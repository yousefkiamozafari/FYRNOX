package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	modIphlpapi = syscall.NewLazyDLL("iphlpapi.dll")
	modPsapi    = syscall.NewLazyDLL("psapi.dll")

	procGetExtendedTcpTable          = modIphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable          = modIphlpapi.NewProc("GetExtendedUdpTable")
	procCreateToolhelp32Snapshot     = kernel32Mod.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW              = kernel32Mod.NewProc("Process32FirstW")
	procProcess32NextW               = kernel32Mod.NewProc("Process32NextW")
	procOpenProcess                  = kernel32Mod.NewProc("OpenProcess")
	procCloseHandle                  = kernel32Mod.NewProc("CloseHandle")
	procQueryFullProcessImageNameW   = kernel32Mod.NewProc("QueryFullProcessImageNameW")
)

const (
	afInet                 = 2
	tcpTableOwnerPidAll    = 5
	udpTableOwnerPid       = 1
	th32csSnapProcess      = 0x00000002
	processQueryLimitedInfo = 0x1000
)

type ProcessInfo struct {
	PID         uint32 `json:"pid"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	IsRunning   bool   `json:"is_running"`
}

type mibTcpRowOwnerPid struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

type mibUdpRowOwnerPid struct {
	dwLocalAddr uint32
	dwLocalPort uint32
	dwOwningPid uint32
}

type process32EntryW struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

type portOwnerCache struct {
	mu            sync.RWMutex
	tcpPorts      map[uint16]string // localPort -> exeName (lowercase)
	udpPorts      map[uint16]string // localPort -> exeName (lowercase)
	remoteTargets map[string]string // remoteAddr:port -> exeName (lowercase)
	lastScan      time.Time
}

var globalPortCache = &portOwnerCache{
	tcpPorts:      make(map[uint16]string),
	udpPorts:      make(map[uint16]string),
	remoteTargets: make(map[string]string),
}

// refreshPortCache populates port -> process executable name mappings
func (c *portOwnerCache) refresh(force bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if !force && now.Sub(c.lastScan) < 800*time.Millisecond {
		return
	}
	c.lastScan = now

	// 1. Scan TCP table
	newTcp := make(map[uint16]string)
	newRemote := make(map[string]string)
	var tcpSize uint32
	r, _, _ := procGetExtendedTcpTable.Call(0, uintptr(unsafe.Pointer(&tcpSize)), 1, afInet, tcpTableOwnerPidAll, 0)
	if tcpSize > 0 {
		buf := make([]byte, tcpSize)
		r, _, _ = procGetExtendedTcpTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&tcpSize)), 1, afInet, tcpTableOwnerPidAll, 0)
		if r == 0 && len(buf) >= 4 {
			numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
			rowSize := unsafe.Sizeof(mibTcpRowOwnerPid{})
			for i := uint32(0); i < numEntries; i++ {
				offset := 4 + uintptr(i)*rowSize
				if offset+rowSize <= uintptr(len(buf)) {
					row := (*mibTcpRowOwnerPid)(unsafe.Pointer(&buf[offset]))
					port := ((row.dwLocalPort >> 8) & 0xFF) | ((row.dwLocalPort & 0xFF) << 8)
					if exe := getProcessExeName(row.dwOwningPid); exe != "" {
						cleanExe := strings.ToLower(exe)
						newTcp[uint16(port)] = cleanExe

						remPort := ((row.dwRemotePort >> 8) & 0xFF) | ((row.dwRemotePort & 0xFF) << 8)
						if remPort > 0 && row.dwRemoteAddr != 0 {
							remIP := net.IPv4(byte(row.dwRemoteAddr), byte(row.dwRemoteAddr>>8), byte(row.dwRemoteAddr>>16), byte(row.dwRemoteAddr>>24)).String()
							newRemote[fmt.Sprintf("%s:%d", remIP, remPort)] = cleanExe
						}
					}
				}
			}
		}
	}
	c.tcpPorts = newTcp
	c.remoteTargets = newRemote

	// 2. Scan UDP table
	newUdp := make(map[uint16]string)
	var udpSize uint32
	r, _, _ = procGetExtendedUdpTable.Call(0, uintptr(unsafe.Pointer(&udpSize)), 1, afInet, udpTableOwnerPid, 0)
	if udpSize > 0 {
		buf := make([]byte, udpSize)
		r, _, _ = procGetExtendedUdpTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&udpSize)), 1, afInet, udpTableOwnerPid, 0)
		if r == 0 && len(buf) >= 4 {
			numEntries := *(*uint32)(unsafe.Pointer(&buf[0]))
			rowSize := unsafe.Sizeof(mibUdpRowOwnerPid{})
			for i := uint32(0); i < numEntries; i++ {
				offset := 4 + uintptr(i)*rowSize
				if offset+rowSize <= uintptr(len(buf)) {
					row := (*mibUdpRowOwnerPid)(unsafe.Pointer(&buf[offset]))
					port := ((row.dwLocalPort >> 8) & 0xFF) | ((row.dwLocalPort & 0xFF) << 8)
					if exe := getProcessExeName(row.dwOwningPid); exe != "" {
						newUdp[uint16(port)] = strings.ToLower(exe)
					}
				}
			}
		}
	}
	c.udpPorts = newUdp
}

func (c *portOwnerCache) lookup(port uint16, isUdp bool) string {
	c.mu.RLock()
	var exe string
	if isUdp {
		exe = c.udpPorts[port]
	} else {
		exe = c.tcpPorts[port]
	}
	c.mu.RUnlock()

	if exe != "" {
		return exe
	}

	// Not found, do quick refresh and check once more
	c.refresh(true)

	c.mu.RLock()
	defer c.mu.RUnlock()
	if isUdp {
		return c.udpPorts[port]
	}
	return c.tcpPorts[port]
}

func (c *portOwnerCache) lookupTarget(target string) string {
	c.mu.RLock()
	exe := c.remoteTargets[target]
	c.mu.RUnlock()
	if exe != "" {
		return exe
	}

	c.refresh(true)

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.remoteTargets[target]
}

// pidExeCache caches PID -> ExeName
var (
	pidCacheLock sync.RWMutex
	pidExeCache  = make(map[uint32]string)
)

func getProcessExeName(pid uint32) string {
	if pid == 0 || pid == 4 {
		return "system"
	}

	pidCacheLock.RLock()
	name, exists := pidExeCache[pid]
	pidCacheLock.RUnlock()
	if exists {
		return name
	}

	hProc, _, _ := procOpenProcess.Call(processQueryLimitedInfo, 0, uintptr(pid))
	if hProc == 0 {
		return ""
	}
	defer procCloseHandle.Call(hProc)

	var buf [1024]uint16
	size := uint32(len(buf))
	r, _, _ := procQueryFullProcessImageNameW.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if r != 0 && size > 0 {
		fullPath := syscall.UTF16ToString(buf[:size])
		exe := strings.ToLower(filepath.Base(fullPath))
		pidCacheLock.Lock()
		pidExeCache[pid] = exe
		pidCacheLock.Unlock()
		return exe
	}
	return ""
}

// IsAppAllowedBySplitTunnel checks whether a socket originating from clientPort or connecting to target should be routed via VPN
func IsAppAllowedBySplitTunnel(clientPort int, target string, isUdp bool, settings VpnSettings) bool {
	if !settings.SplitTunnelEnabled {
		return true // Split tunneling disabled: all traffic routes through VPN
	}

	targetExe := ""
	if clientPort > 0 {
		targetExe = globalPortCache.lookup(uint16(clientPort), isUdp)
	}
	if targetExe == "" && target != "" {
		targetExe = globalPortCache.lookupTarget(target)
	}
	if targetExe == "" {
		// If process cannot be determined, default to allowing traffic
		return true
	}

	// Normalize configured list
	configured := make(map[string]struct{})
	for _, app := range settings.SplitTunnelApps {
		clean := strings.ToLower(strings.TrimSpace(filepath.Base(app)))
		if clean != "" {
			configured[clean] = struct{}{}
		}
	}

	_, match := configured[targetExe]

	mode := strings.ToLower(settings.SplitTunnelMode)
	if mode == "proxy_only" || mode == "whitelist" {
		// Only apps in list are proxied; all others bypass direct
		return match
	}

	// Default is "bypass" or "blacklist": apps in list bypass direct; all others are proxied
	return !match
}

// GetRunningProcesses returns a list of unique active processes with UI windows or background networking
func GetRunningProcesses() []ProcessInfo {
	hSnap, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if hSnap == uintptr(syscall.InvalidHandle) {
		return nil
	}
	defer procCloseHandle.Call(hSnap)

	var entry process32EntryW
	entry.dwSize = uint32(unsafe.Sizeof(entry))

	r, _, _ := procProcess32FirstW.Call(hSnap, uintptr(unsafe.Pointer(&entry)))
	if r == 0 {
		return nil
	}

	seen := make(map[string]ProcessInfo)
	currentPid := uint32(os.Getpid())

	for {
		pid := entry.th32ProcessID
		if pid > 4 && pid != currentPid {
			name := syscall.UTF16ToString(entry.szExeFile[:])
			cleanName := strings.ToLower(filepath.Base(name))

			// Exclude common OS internal non-interactive host workers
			if cleanName != "" && !isSystemProcessExcluded(cleanName) {
				if _, ok := seen[cleanName]; !ok {
					fullPath := ""
					hProc, _, _ := procOpenProcess.Call(processQueryLimitedInfo, 0, uintptr(pid))
					if hProc != 0 {
						var buf [1024]uint16
						size := uint32(len(buf))
						r, _, _ := procQueryFullProcessImageNameW.Call(hProc, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
						if r != 0 && size > 0 {
							fullPath = syscall.UTF16ToString(buf[:size])
						}
						procCloseHandle.Call(hProc)
					}

					seen[cleanName] = ProcessInfo{
						PID:         pid,
						Name:        name,
						Path:        fullPath,
						Description: getFriendlyAppName(cleanName),
						IsRunning:   true,
					}
				}
			}
		}

		r, _, _ = procProcess32NextW.Call(hSnap, uintptr(unsafe.Pointer(&entry)))
		if r == 0 {
			break
		}
	}

	result := make([]ProcessInfo, 0, len(seen))
	for _, p := range seen {
		result = append(result, p)
	}
	return result
}

// GetInstalledApplications reads Windows Registry uninstall keys to list user software
func GetInstalledApplications() []ProcessInfo {
	roots := []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER}
	paths := []string{
		`Software\Microsoft\Windows\CurrentVersion\Uninstall`,
		`Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	seen := make(map[string]ProcessInfo)

	for _, root := range roots {
		for _, p := range paths {
			k, err := registry.OpenKey(root, p, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			names, err := k.ReadSubKeyNames(-1)
			if err == nil {
				for _, sub := range names {
					subKey, err := registry.OpenKey(k, sub, registry.QUERY_VALUE)
					if err != nil {
						continue
					}
					dispName, _, err1 := subKey.GetStringValue("DisplayName")
					dispIcon, _, _ := subKey.GetStringValue("DisplayIcon")
					instLocation, _, _ := subKey.GetStringValue("InstallLocation")
					subKey.Close()

					if err1 == nil && strings.TrimSpace(dispName) != "" {
						exeName := ""
						if dispIcon != "" {
							clean := strings.Trim(strings.Split(dispIcon, ",")[0], "\"")
							if strings.HasSuffix(strings.ToLower(clean), ".exe") {
								exeName = filepath.Base(clean)
							}
						}
						if exeName == "" && instLocation != "" {
							// Search root of install location for main exe
							if files, err := os.ReadDir(instLocation); err == nil {
								for _, f := range files {
									if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".exe") {
										exeName = f.Name()
										break
									}
								}
							}
						}

						if exeName != "" {
							cleanExe := strings.ToLower(exeName)
							if _, exists := seen[cleanExe]; !exists {
								seen[cleanExe] = ProcessInfo{
									Name:        exeName,
									Path:        dispIcon,
									Description: dispName,
									IsRunning:   false,
								}
							}
						}
					}
				}
			}
			k.Close()
		}
	}

	result := make([]ProcessInfo, 0, len(seen))
	for _, p := range seen {
		result = append(result, p)
	}
	return result
}

func isSystemProcessExcluded(name string) bool {
	excluded := map[string]struct{}{
		"system": {}, "smss.exe": {}, "csrss.exe": {}, "wininit.exe": {},
		"services.exe": {}, "lsass.exe": {}, "svchost.exe": {}, "fontdrvhost.exe": {},
		"dwm.exe": {}, "taskhostw.exe": {}, "sihost.exe": {}, "ctfmon.exe": {},
		"searchhost.exe": {}, "startmenuexperiencehost.exe": {}, "conhost.exe": {},
		"runtimebroker.exe": {}, "werfault.exe": {}, "wermgr.exe": {},
	}
	_, ok := excluded[strings.ToLower(name)]
	return ok
}

func getFriendlyAppName(exe string) string {
	names := map[string]string{
		"chrome.exe":       "Google Chrome",
		"firefox.exe":      "Mozilla Firefox",
		"msedge.exe":       "Microsoft Edge",
		"brave.exe":        "Brave Browser",
		"opera.exe":        "Opera Browser",
		"telegram.exe":     "Telegram Desktop",
		"discord.exe":      "Discord",
		"steam.exe":        "Steam Client",
		"epicgameslauncher.exe": "Epic Games Launcher",
		"spotify.exe":      "Spotify",
		"obs64.exe":        "OBS Studio",
		"code.exe":         "Visual Studio Code",
		"idman.exe":        "Internet Download Manager",
	}
	if n, ok := names[strings.ToLower(exe)]; ok {
		return n
	}
	return strings.Title(strings.TrimSuffix(exe, ".exe"))
}
