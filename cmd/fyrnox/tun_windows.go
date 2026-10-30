package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/tun"
	"github.com/xjasonlyu/tun2socks/v2/proxy/socks5"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

var (
	tunModeActive   bool
	wintunActive    bool
	tunMu           sync.Mutex
	tunDev          device.Device
	tunStack        *stack.Stack
	remoteServerIPs []string
	defaultGateway  string
	defaultIfaceIdx int

	modDnsapi                 = syscall.NewLazyDLL("dnsapi.dll")
	procDnsFlushResolverCache = modDnsapi.NewProc("DnsFlushResolverCache")
)

// FlushDNSCache flushes the Windows DNS resolver cache via DnsFlushResolverCache
func FlushDNSCache() {
	if procDnsFlushResolverCache.Find() == nil {
		_, _, _ = procDnsFlushResolverCache.Call()
	} else {
		_, _ = runSilentCmdWithTimeout(800*time.Millisecond, "ipconfig", "/flushdns")
	}
	LogInfo("Windows DNS resolver cache flushed successfully")
}

const (
	tunDeviceName = "FyrnoxTun"
	tunIPAddr     = "10.254.1.2"
	tunNetmask    = "255.255.255.0"
	tunGateway    = "10.254.1.1"
	tunMTU        = 1500
)

// newSilentCmdContext creates an exec.Cmd with context configured without visible window
func newSilentCmdContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
}

// runSilentCmdWithTimeout executes a command silently with bounded context timeout
func runSilentCmdWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := newSilentCmdContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// newSilentCmd creates an exec.Cmd configured to run without creating a visible console window
func newSilentCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd
}

// runSilentCmd runs a command silently without showing a console window and returns combined output
func runSilentCmd(name string, args ...string) (string, error) {
	cmd := newSilentCmd(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// IsAdmin checks if the current process is running with elevated administrator rights
func IsAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

// IsTunModeActive returns whether TUN mode or full system proxy mode is currently enabled
func IsTunModeActive() bool {
	tunMu.Lock()
	defer tunMu.Unlock()
	return tunModeActive
}

// IsWintunActive returns whether the Layer-3 Wintun network stack is running
func IsWintunActive() bool {
	tunMu.Lock()
	defer tunMu.Unlock()
	return wintunActive
}

// getPhysicalDefaultGateway finds the default gateway and interface index for the physical network
func getPhysicalDefaultGateway() (string, int, error) {
	// 1. Try route print 0.0.0.0 first for fast lookup
	out, err := newSilentCmd("route", "print", "0.0.0.0").CombinedOutput()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 5 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
				gw := fields[2]
				ip := net.ParseIP(gw)
				if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() && !strings.HasPrefix(gw, "172.18.") {
					// Use powershell to get interface index reliably
					break
				}
			}
		}
	}

	// 2. Query PowerShell Get-NetRoute for precise physical adapter information
	psCmd := `(Get-NetRoute -DestinationPrefix '0.0.0.0/0' | Where-Object { $_.InterfaceAlias -notmatch 'tun|tap|vpn|virtual|fyrnox' -and $_.NextHop -ne '0.0.0.0' } | Sort-Object RouteMetric | Select-Object -First 1 NextHop, InterfaceIndex) | ConvertTo-Json -Compress`
	psOut, psErr := newSilentCmd("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", psCmd).Output()
	if psErr == nil {
		var res struct {
			NextHop        string `json:"NextHop"`
			InterfaceIndex int    `json:"InterfaceIndex"`
		}
		if jsonErr := json.Unmarshal(psOut, &res); jsonErr == nil && res.NextHop != "" {
			return res.NextHop, res.InterfaceIndex, nil
		}
	}

	return "", 0, fmt.Errorf("could not determine default physical gateway")
}

// EnableTunMode sets up Layer-3 Wintun routing (when Administrator) and WinHTTP + user environment proxies
func EnableTunMode(socksAddr, httpAddr, remoteHost, remoteIP, ipProtocol string) error {
	tunMu.Lock()
	defer tunMu.Unlock()

	LogInfo("Configuring System-Wide Routing (Target Server: %s)...", remoteHost)
	cleanupHostRoutes()
	FlushDNSCache()

	// Ensure DLL search path includes executable directory for wintun.dll
	if exePath, err := os.Executable(); err == nil {
		_ = windows.SetDllDirectory(filepath.Dir(exePath))
	}

	// 1. Always configure WinHTTP system proxy (if elevated or available)
	if httpAddr != "" {
		out, err := newSilentCmd("netsh", "winhttp", "set", "proxy", httpAddr).CombinedOutput()
		if err != nil {
			LogWarn("WinHTTP set proxy notice: %v (%s)", err, stringsTrim(string(out)))
		} else {
			LogInfo("WinHTTP system proxy configured: %s", httpAddr)
		}
	}

	// 2. Configure user environment variables for CLI and development tools
	socksHost, socksPort, _ := net.SplitHostPort(socksAddr)
	if socksHost == "" {
		socksHost = "127.0.0.1"
	}
	if socksPort == "" {
		socksPort = "20808"
	}

	httpHost, httpPort, _ := net.SplitHostPort(httpAddr)
	if httpHost == "" {
		httpHost = "127.0.0.1"
	}
	if httpPort == "" {
		httpPort = "20809"
	}

	socksProxyURL := fmt.Sprintf("socks5://%s:%s", socksHost, socksPort)
	httpProxyURL := fmt.Sprintf("http://%s:%s", httpHost, httpPort)

	// Direct Windows Registry configuration (< 1ms, zero subprocess, zero popup)
	setEnvVarsDirect(socksProxyURL, httpProxyURL)

	// 3. Layer-3 Wintun Routing
	if !IsAdmin() {
		LogWarn("Running as Standard User: Kernel TUN (Wintun) requires Administrator privileges.")
		LogInfo("Full System Proxy (WinINet + WinHTTP + Environment Variables) is active.")
		tunModeActive = true
		wintunActive = false
		return nil
	}

	LogInfo("Administrator privileges confirmed: Initializing Wintun Layer-3 Network Stack...")

	// Find physical gateway to prevent routing loop
	gw, ifaceIdx, err := getPhysicalDefaultGateway()
	if err != nil {
		LogWarn("Could not determine physical gateway (%v); falling back to System Proxy mode", err)
		tunModeActive = true
		wintunActive = false
		return nil
	}
	defaultGateway = gw
	defaultIfaceIdx = ifaceIdx
	LogInfo("Physical Default Gateway identified: %s (Interface Index: %d)", gw, ifaceIdx)

	// Add explicit host routes for the active tunnel server to prevent routing loops
	remoteServerIPs = nil
	addHostRoute := func(ipStr string) {
		if ipStr == "" || ipStr == "127.0.0.1" || strings.HasPrefix(ipStr, "10.") {
			return
		}
		for _, existing := range remoteServerIPs {
			if existing == ipStr {
				return
			}
		}
		remoteServerIPs = append(remoteServerIPs, ipStr)
		routeCmd := newSilentCmd("route", "add", ipStr, "mask", "255.255.255.255", gw, "metric", "1")
		if ifaceIdx > 0 {
			routeCmd = newSilentCmd("route", "add", ipStr, "mask", "255.255.255.255", gw, "if", strconv.Itoa(ifaceIdx), "metric", "1")
		}
		_ = routeCmd.Run()
		LogInfo("Routing loop prevention route added for %s -> %s", ipStr, gw)
	}

	if remoteIP != "" {
		addHostRoute(remoteIP)
	}
	if remoteHost != "" {
		resolvedIPs, err := net.LookupIP(remoteHost)
		if err == nil {
			for _, ip := range resolvedIPs {
				if ip4 := ip.To4(); ip4 != nil {
					addHostRoute(ip4.String())
				}
			}
		}
	}

	// Create and start tun2socks Wintun adapter
	dev, err := tun.Open(tunDeviceName, tunMTU)
	if err != nil {
		LogWarn("Failed to create Wintun adapter: %v (using System Proxy mode)", err)
		cleanupHostRoutes()
		tunModeActive = true
		wintunActive = false
		return nil
	}
	tunDev = dev

	// Configure SOCKS5 upstream proxy for tun2socks
	s5, err := socks5.New(socksAddr, "", "")
	if err != nil {
		LogWarn("Failed to initialize SOCKS5 upstream for TUN: %v", err)
		closeDevice(dev)
		tunDev = nil
		cleanupHostRoutes()
		tunModeActive = true
		wintunActive = false
		return nil
	}
	tunnel.T().SetProxy(s5)

	// Create gVisor TCP/IP network stack
	netStack, err := core.CreateStack(&core.Config{
		LinkEndpoint:     dev,
		TransportHandler: tunnel.T(),
	})
	if err != nil {
		LogWarn("Failed to initialize gVisor netstack: %v", err)
		closeDevice(dev)
		tunDev = nil
		cleanupHostRoutes()
		tunModeActive = true
		wintunActive = false
		return nil
	}
	tunStack = netStack

	// Wait briefly for Windows NDIS to register the newly created interface
	time.Sleep(300 * time.Millisecond)

	// Configure IP and DNS on Wintun interface
	_ = newSilentCmd("netsh", "interface", "ipv4", "set", "address",
		fmt.Sprintf("name=%s", tunDeviceName),
		"source=static",
		fmt.Sprintf("addr=%s", tunIPAddr),
		fmt.Sprintf("mask=%s", tunNetmask),
		"gateway=none",
	).Run()

	_ = newSilentCmd("netsh", "interface", "ipv4", "set", "dnsservers",
		fmt.Sprintf("name=%s", tunDeviceName),
		"source=static",
		"address=1.1.1.1",
		"register=none",
		"validate=no",
	).Run()

	_ = newSilentCmd("netsh", "interface", "ipv4", "add", "dnsservers",
		fmt.Sprintf("name=%s", tunDeviceName),
		"address=8.8.8.8",
		"index=2",
		"validate=no",
	).Run()

	// Prioritize Wintun interface metric to 1
	_ = newSilentCmd("netsh", "interface", "ipv4", "set", "interface",
		fmt.Sprintf("name=%s", tunDeviceName),
		"metric=1",
	).Run()

	// Find the interface index of FyrnoxTun for explicit route binding
	tunIfIndex := 0
	if tunIface, err := net.InterfaceByName(tunDeviceName); err == nil {
		tunIfIndex = tunIface.Index
	}

	// Add split default routes (0.0.0.0/1 and 128.0.0.0/1) to redirect all traffic into TUN
	if tunIfIndex > 0 {
		_ = newSilentCmd("route", "add", "0.0.0.0", "mask", "128.0.0.0", tunGateway, "if", strconv.Itoa(tunIfIndex), "metric", "1").Run()
		_ = newSilentCmd("route", "add", "128.0.0.0", "mask", "128.0.0.0", tunGateway, "if", strconv.Itoa(tunIfIndex), "metric", "1").Run()
	} else {
		_ = newSilentCmd("route", "add", "0.0.0.0", "mask", "128.0.0.0", tunGateway, "metric", "1").Run()
		_ = newSilentCmd("route", "add", "128.0.0.0", "mask", "128.0.0.0", tunGateway, "metric", "1").Run()
	}

	// IPv6 leak protection: blackhole non-tunnel IPv6 traffic when in IPv4-only mode
	if ipProtocol == "ipv4" {
		if tunIfIndex > 0 {
			_, _ = runSilentCmdWithTimeout(2*time.Second, "route", "-6", "add", "::/1", "fe80::1", "if", strconv.Itoa(tunIfIndex), "metric", "1")
			_, _ = runSilentCmdWithTimeout(2*time.Second, "route", "-6", "add", "8000::/1", "fe80::1", "if", strconv.Itoa(tunIfIndex), "metric", "1")
		} else {
			_, _ = runSilentCmdWithTimeout(2*time.Second, "route", "-6", "add", "::/1", "fe80::1", "metric", "1")
			_, _ = runSilentCmdWithTimeout(2*time.Second, "route", "-6", "add", "8000::/1", "fe80::1", "metric", "1")
		}
		LogInfo("IPv6 leak protection blackhole active in TUN mode")
	}

	wintunActive = true
	tunModeActive = true
	FlushDNSCache()
	LogInfo("Wintun Layer-3 routing successfully activated (%s, IP: %s, IF: %d)", tunDeviceName, tunIPAddr, tunIfIndex)
	return nil
}

// cleanupHostRoutes removes any explicit host routes added for remote server IPs
func cleanupHostRoutes() {
	for _, ip := range remoteServerIPs {
		_, _ = runSilentCmdWithTimeout(1500*time.Millisecond, "route", "delete", ip, "mask", "255.255.255.255")
	}
	remoteServerIPs = nil
}

// DisableTunMode restores routing table, tears down Wintun adapter, and cleans proxies
func DisableTunMode() error {
	tunMu.Lock()
	defer tunMu.Unlock()

	if !tunModeActive && !wintunActive {
		return nil
	}

	LogInfo("Deactivating TUN Mode and restoring system networking...")

	// 1. Tear down Wintun routing and network stack if active
	if wintunActive {
		// 1a. Close Wintun device FIRST so netstack packet loops terminate immediately!
		if tunDev != nil {
			closeDevice(tunDev)
			tunDev = nil
		}

		// Parallelize route deletions for instantaneous teardown (< 100ms)
		var routeWg sync.WaitGroup
		routeWg.Add(5)
		go func() {
			defer routeWg.Done()
			_, _ = runSilentCmdWithTimeout(500*time.Millisecond, "route", "delete", "0.0.0.0", "mask", "128.0.0.0")
		}()
		go func() {
			defer routeWg.Done()
			_, _ = runSilentCmdWithTimeout(500*time.Millisecond, "route", "delete", "128.0.0.0", "mask", "128.0.0.0")
		}()
		go func() {
			defer routeWg.Done()
			_, _ = runSilentCmdWithTimeout(500*time.Millisecond, "route", "-6", "delete", "::/1")
		}()
		go func() {
			defer routeWg.Done()
			_, _ = runSilentCmdWithTimeout(500*time.Millisecond, "route", "-6", "delete", "8000::/1")
		}()
		go func() {
			defer routeWg.Done()
			cleanupHostRoutes()
		}()

		// 1b. Tear down gVisor netstack with bounded timeout
		if tunStack != nil {
			st := tunStack
			tunStack = nil
			st.Close()
			done := make(chan struct{})
			go func() {
				st.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(200 * time.Millisecond):
				LogWarn("Netstack wait timed out during teardown (continuing cleanly)")
			}
		}

		routeWg.Wait()
		wintunActive = false
		LogInfo("Wintun Layer-3 adapter removed and routing table restored")
	}

	// 2. Reset WinHTTP system proxy with bounded timeout
	_, _ = runSilentCmdWithTimeout(500*time.Millisecond, "netsh", "winhttp", "reset", "proxy")

	// 3. Remove user-level environment variables via direct Windows Registry (< 1ms)
	clearEnvVarsDirect()

	FlushDNSCache()

	tunModeActive = false
	LogInfo("System routing and proxies completely restored")
	return nil
}

// VerifyTunCleanup verifies that TUN routes and interfaces have been removed cleanly
func VerifyTunCleanup() bool {
	out, err := runSilentCmdWithTimeout(2*time.Second, "route", "print", "0.0.0.0")
	if err == nil && (strings.Contains(out, "10.254.1.1") || strings.Contains(out, "128.0.0.0")) {
		LogWarn("Detected lingering TUN routes after disconnect; executing force route delete...")
		_, _ = runSilentCmdWithTimeout(1500*time.Millisecond, "route", "delete", "0.0.0.0", "mask", "128.0.0.0")
		_, _ = runSilentCmdWithTimeout(1500*time.Millisecond, "route", "delete", "128.0.0.0", "mask", "128.0.0.0")
		return false
	}
	return true
}

func setEnvVarsDirect(socksURL, httpURL string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err == nil {
		defer k.Close()
		_ = k.SetStringValue("ALL_PROXY", socksURL)
		_ = k.SetStringValue("HTTP_PROXY", httpURL)
		_ = k.SetStringValue("HTTPS_PROXY", httpURL)
		notifyEnvChange()
	}
}

func clearEnvVarsDirect() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err == nil {
		defer k.Close()
		_ = k.DeleteValue("ALL_PROXY")
		_ = k.DeleteValue("HTTP_PROXY")
		_ = k.DeleteValue("HTTPS_PROXY")
		notifyEnvChange()
	}
}

func notifyEnvChange() {
	envPtr, err := windows.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	_, _, _ = syscall.NewLazyDLL("user32.dll").NewProc("SendMessageTimeoutW").Call(
		0xffff, // HWND_BROADCAST
		0x001A, // WM_SETTINGCHANGE
		0,
		uintptr(unsafe.Pointer(envPtr)),
		2,   // SMTO_ABORTIFHUNG
		100, // 100ms timeout
		0,
	)
}

var ErrUACCancelled = errors.New("UAC elevation prompt was cancelled by user")

func isUACCancelled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, windows.ERROR_CANCELLED) || errors.Is(err, ErrUACCancelled) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "canceled") || strings.Contains(msg, "cancelled")
}

func buildElevationArgs(serverHostname, mode string) string {
	args := "-elevated"
	if serverHostname != "" {
		args += " -server " + serverHostname
	}
	if mode != "" {
		args += " -mode " + mode
	}
	return args
}

// RestartAsAdminWithParams restarts the Fyrnox process with Administrator UAC prompt passing server and mode
func RestartAsAdminWithParams(serverHostname, mode string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("runas")
	path, _ := syscall.UTF16PtrFromString(exe)
	argsStr := buildElevationArgs(serverHostname, mode)
	argsPtr, _ := syscall.UTF16PtrFromString(argsStr)

	err = windows.ShellExecute(0, verb, path, argsPtr, nil, windows.SW_SHOWNORMAL)
	if err != nil {
		if isUACCancelled(err) {
			return ErrUACCancelled
		}
		return err
	}
	return nil
}

// RestartAsAdmin restarts the Fyrnox process with Administrator UAC prompt
func RestartAsAdmin() error {
	return RestartAsAdminWithParams("", "tun")
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\r' || s[len(s)-1] == '\n' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}

func closeDevice(d any) {
	if d == nil {
		return
	}
	if closer, ok := d.(interface{ Close() }); ok {
		closer.Close()
	} else if closer, ok := d.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}
