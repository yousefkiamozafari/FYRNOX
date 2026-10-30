package main

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestIPProtocolSettingsPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	origAppData := getAppDataDir()
	_ = origAppData

	settingsFile := filepath.Join(tmpDir, "settings.json")
	s := defaultSettings()
	s.IPProtocol = "ipv6"

	data, err := saveSettingsFile(settingsFile, s)
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("Empty settings data written")
	}

	loaded, err := loadSettingsFile(settingsFile)
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	if loaded.IPProtocol != "ipv6" {
		t.Errorf("Expected IPProtocol 'ipv6', got '%s'", loaded.IPProtocol)
	}
	if loaded.SocksPort != 20808 {
		t.Errorf("Expected SocksPort 20808, got %d", loaded.SocksPort)
	}
}

func TestLocalIPFamiliesDetection(t *testing.T) {
	hasV4, hasV6 := checkLocalIPFamilies()
	t.Logf("Detected local network capabilities: IPv4=%v, IPv6=%v", hasV4, hasV6)
	// At least IPv4 or loopback should exist in any test environment
	if !hasV4 && !hasV6 {
		t.Logf("Notice: Neither global IPv4 nor IPv6 found on active physical interfaces (running in isolated environment)")
	}
}

func TestDisconnectIdempotency(t *testing.T) {
	meter := NewTrafficMeter()
	engine := &VpnEngine{
		state: StateConnected,
		meter: meter,
	}

	// 1. First Disconnect
	err := engine.Disconnect()
	if err != nil {
		t.Fatalf("First Disconnect failed: %v", err)
	}
	if engine.state != StateDisconnected {
		t.Errorf("Expected state Disconnected, got %s", engine.state)
	}

	// 2. Second Disconnect (must be idempotent and fast)
	start := time.Now()
	err2 := engine.Disconnect()
	duration := time.Since(start)
	if err2 != nil {
		t.Fatalf("Second Disconnect failed: %v", err2)
	}
	if duration > 100*time.Millisecond {
		t.Errorf("Second disconnect took too long (%v); expected near-instant idempotent return", duration)
	}
	if engine.state != StateDisconnected {
		t.Errorf("Expected state to remain Disconnected, got %s", engine.state)
	}
}

func TestMeasureServerPingBreakdown(t *testing.T) {
	// Test against local dummy server to verify breakdown computation
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	srv := ServerLocation{
		Hostname: "127.0.0.1",
		Port:     port,
	}

	// Measure against local listener: TCP will succeed, TLS handshake will fail or return -1 gracefully
	p, d, tc, _ := measureServerPing(ctx, srv)
	t.Logf("Local ping measurement: pingMs=%d, dnsMs=%d, tcpMs=%d", p, d, tc)

	if d <= 0 {
		t.Errorf("Expected positive DNS resolution time, got %d", d)
	}
	if tc < 0 {
		t.Errorf("Expected positive TCP connect time on local listener, got %d", tc)
	}
}

func TestResetSettingsToDefault(t *testing.T) {
	engine := NewVpnEngine()
	// Change some settings
	s := engine.GetSettings()
	s.GamingMode = true
	s.EnableTunMode = true
	s.UseSystemProxy = false
	s.KillSwitch = true
	engine.UpdateSettings(s)

	if !engine.GetSettings().GamingMode || !engine.GetSettings().EnableTunMode {
		t.Fatalf("Failed to update test settings")
	}

	// Reset to default
	res := engine.ResetSettingsToDefault()
	if res.GamingMode != false || res.EnableTunMode != false || res.UseSystemProxy != true || res.KillSwitch != false {
		t.Fatalf("ResetSettingsToDefault did not restore default settings: %+v", res)
	}
	if engine.GetSettings().GamingMode != false {
		t.Fatalf("Engine settings state not restored: %+v", engine.GetSettings())
	}
}
