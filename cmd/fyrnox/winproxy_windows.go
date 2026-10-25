package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	wininet                = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOption  = wininet.NewProc("InternetSetOptionW")
	user32                 = syscall.NewLazyDLL("user32.dll")
	procSendMessageTimeout = user32.NewProc("SendMessageTimeoutW")
)

const (
	internetOptionSettingsChanged = 39
	internetOptionRefresh         = 37
)

type ProxySettingsBackup struct {
	Enabled    bool
	Server     string
	Override   string
	AutoDetect bool
}

var lastBackup *ProxySettingsBackup

// BackupCurrentProxySettings reads and saves the current Windows proxy settings
func BackupCurrentProxySettings() (*ProxySettingsBackup, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	backup := &ProxySettingsBackup{}
	enabled, _, err := k.GetIntegerValue("ProxyEnable")
	if err == nil {
		backup.Enabled = enabled != 0
	}
	server, _, err := k.GetStringValue("ProxyServer")
	if err == nil {
		backup.Server = server
	}
	override, _, err := k.GetStringValue("ProxyOverride")
	if err == nil {
		backup.Override = override
	}

	lastBackup = backup
	return backup, nil
}

const defaultBypassRules = "<local>;localhost;127.*;10.*;172.16.*;192.168.*"

// FormatBypassRules combines system default LAN bypass with any user-defined bypass domains
func FormatBypassRules(customBypass string) string {
	clean := strings.TrimSpace(customBypass)
	if clean == "" {
		return defaultBypassRules
	}
	clean = strings.ReplaceAll(clean, "\r\n", ";")
	clean = strings.ReplaceAll(clean, "\n", ";")
	clean = strings.ReplaceAll(clean, ",", ";")
	parts := strings.Split(clean, ";")

	unique := make(map[string]struct{})
	var list []string

	for _, p := range strings.Split(defaultBypassRules, ";") {
		p = strings.TrimSpace(p)
		if p != "" {
			unique[strings.ToLower(p)] = struct{}{}
			list = append(list, p)
		}
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			low := strings.ToLower(p)
			if _, exists := unique[low]; !exists {
				unique[low] = struct{}{}
				list = append(list, p)
			}
		}
	}
	return strings.Join(list, ";")
}

// SetSystemProxy enables or disables Windows system-wide proxy with default bypass rules
func SetSystemProxy(enable bool, httpProxyAddr, socksProxyAddr string) error {
	return SetSystemProxyWithBypass(enable, httpProxyAddr, socksProxyAddr, "")
}

// SetSystemProxyWithBypass enables or disables Windows system-wide proxy with custom bypass rules
func SetSystemProxyWithBypass(enable bool, httpProxyAddr, socksProxyAddr, customBypass string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening registry key: %w", err)
	}
	defer k.Close()

	if enable {
		if lastBackup == nil {
			_, _ = BackupCurrentProxySettings()
		}

		// Explicit multi-protocol proxy format for Windows:
		// http=127.0.0.1:20809;https=127.0.0.1:20809;socks=127.0.0.1:20808
		var proxyString string
		if httpProxyAddr != "" && socksProxyAddr != "" {
			proxyString = fmt.Sprintf("http=%s;https=%s;socks=%s", httpProxyAddr, httpProxyAddr, socksProxyAddr)
		} else if httpProxyAddr != "" {
			proxyString = fmt.Sprintf("http=%s;https=%s", httpProxyAddr, httpProxyAddr)
		} else if socksProxyAddr != "" {
			proxyString = fmt.Sprintf("socks=%s", socksProxyAddr)
		}

		if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
			return err
		}
		if err := k.SetStringValue("ProxyServer", proxyString); err != nil {
			return err
		}

		// Configure bypass rules (default LAN + user defined)
		bypassStr := FormatBypassRules(customBypass)
		_ = k.SetStringValue("ProxyOverride", bypassStr)

		// Also configure WinHTTP proxy so Windows system services & PowerShell route through proxy
		if httpProxyAddr != "" {
			_ = newSilentCmd("netsh", "winhttp", "set", "proxy", fmt.Sprintf("http=%s;https=%s", httpProxyAddr, httpProxyAddr), bypassStr).Run()
		}

		// Also set user environment variables for CLI tools
		socksURL := fmt.Sprintf("socks5://%s", socksProxyAddr)
		httpURL := fmt.Sprintf("http://%s", httpProxyAddr)
		setEnvVarsDirect(socksURL, httpURL)
	} else {
		if err := k.SetDWordValue("ProxyEnable", 0); err != nil {
			return err
		}
		_ = k.SetStringValue("ProxyServer", "")

		// Reset WinHTTP proxy
		_ = newSilentCmd("netsh", "winhttp", "reset", "proxy").Run()

		// Clear user environment variables
		clearEnvVarsDirect()
	}

	refreshSystemProxy()
	FlushDNSCache()
	return nil
}

// RestoreSystemProxy restores the proxy settings from backup cleanly
func RestoreSystemProxy() error {
	if lastBackup == nil {
		return SetSystemProxy(false, "", "")
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	enableVal := uint32(0)
	if lastBackup.Enabled {
		enableVal = 1
	}
	_ = k.SetDWordValue("ProxyEnable", enableVal)
	if enableVal == 1 && lastBackup.Server != "" {
		_ = k.SetStringValue("ProxyServer", lastBackup.Server)
	} else {
		_ = k.SetStringValue("ProxyServer", "")
	}
	if lastBackup.Override != "" {
		_ = k.SetStringValue("ProxyOverride", lastBackup.Override)
	}

	// Reset WinHTTP proxy
	_ = newSilentCmd("netsh", "winhttp", "reset", "proxy").Run()

	// Clear user environment variables
	clearEnvVarsDirect()

	refreshSystemProxy()
	FlushDNSCache()
	return nil
}

// CleanOrphanedProxy checks if an old/crashed Fyrnox session left system proxy active
func CleanOrphanedProxy() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	server, _, err := k.GetStringValue("ProxyServer")
	if err == nil && (strings.Contains(server, "20808") || strings.Contains(server, "20809") || strings.Contains(server, "10808") || strings.Contains(server, "10809")) {
		LogInfo("Cleaning orphaned proxy settings from previous session: %s", server)
		_ = k.SetDWordValue("ProxyEnable", 0)
		_ = k.SetStringValue("ProxyServer", "")
		refreshSystemProxy()
	}
}

func refreshSystemProxy() {
	procInternetSetOption.Call(0, uintptr(internetOptionSettingsChanged), 0, 0)
	procInternetSetOption.Call(0, uintptr(internetOptionRefresh), 0, 0)

	// Broadcast WM_SETTINGCHANGE to HWND_BROADCAST so running browsers (Chrome, Edge, Brave, etc.) reload proxy settings immediately
	subKey, _ := syscall.UTF16PtrFromString("Internet Settings")
	var result uintptr
	procSendMessageTimeout.Call(
		uintptr(0xFFFF), // HWND_BROADCAST
		uintptr(0x001A), // WM_SETTINGCHANGE
		0,
		uintptr(unsafe.Pointer(subKey)),
		uintptr(0x0002), // SMTO_ABORTIFHUNG
		1000,            // 1000ms timeout
		uintptr(unsafe.Pointer(&result)),
	)
}
