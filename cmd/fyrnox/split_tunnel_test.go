package main

import (
	"testing"
)

func TestSplitTunnelRules(t *testing.T) {
	// 1. Split tunnel disabled -> all traffic allowed through VPN
	sDisabled := VpnSettings{
		SplitTunnelEnabled: false,
		SplitTunnelApps:    []string{"chrome.exe", "telegram.exe"},
		SplitTunnelMode:    "bypass",
	}
	if !IsAppAllowedBySplitTunnel(0, "example.com:443", false, sDisabled) {
		t.Errorf("Expected all traffic allowed when split tunnel is disabled")
	}

	// 2. Mock lookup with a registered test target in cache
	globalPortCache.mu.Lock()
	globalPortCache.remoteTargets["203.0.113.10:443"] = "telegram.exe"
	globalPortCache.remoteTargets["203.0.113.20:443"] = "chrome.exe"
	globalPortCache.tcpPorts[45678] = "discord.exe"
	globalPortCache.mu.Unlock()

	// Bypass mode: telegram should bypass (return false), chrome should bypass (return false), others proxied (return true)
	sBypass := VpnSettings{
		SplitTunnelEnabled: true,
		SplitTunnelMode:    "bypass",
		SplitTunnelApps:    []string{"Telegram.exe", "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"},
	}

	if IsAppAllowedBySplitTunnel(0, "203.0.113.10:443", false, sBypass) {
		t.Errorf("Telegram should be bypassed (false) in bypass mode")
	}
	if IsAppAllowedBySplitTunnel(0, "203.0.113.20:443", false, sBypass) {
		t.Errorf("Chrome should be bypassed (false) in bypass mode")
	}
	if !IsAppAllowedBySplitTunnel(45678, "some-ip:443", false, sBypass) {
		t.Errorf("Discord should be proxied (true) in bypass mode when not in list")
	}

	// Whitelist / proxy_only mode: only telegram & chrome should be proxied (return true)
	sWhitelist := VpnSettings{
		SplitTunnelEnabled: true,
		SplitTunnelMode:    "proxy_only",
		SplitTunnelApps:    []string{"telegram.exe"},
	}

	if !IsAppAllowedBySplitTunnel(0, "203.0.113.10:443", false, sWhitelist) {
		t.Errorf("Telegram should be proxied (true) in proxy_only mode")
	}
	if IsAppAllowedBySplitTunnel(45678, "some-ip:443", false, sWhitelist) {
		t.Errorf("Discord should bypass (false) in proxy_only mode when not in list")
	}
}
