package main

import (
	"net"
	"testing"
)

func TestGetLocalSubnetIP(t *testing.T) {
	ip := GetLocalSubnetIP()
	if ip == "" {
		t.Errorf("Expected non-empty IP")
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		t.Errorf("Expected valid IP format, got '%s'", ip)
	}
	t.Logf("Detected local LAN subnet IP: %s", ip)
}

func TestLANShareManagerLifecycle(t *testing.T) {
	meter := NewTrafficMeter()
	opener := &dummyOpener{}

	status1 := globalLANShare.GetStatus()
	if status1.Enabled {
		t.Errorf("Expected initial state disabled")
	}

	err := globalLANShare.SetEnabled(true, opener, meter)
	if err != nil {
		t.Fatalf("Failed to enable LAN share: %v", err)
	}

	status2 := globalLANShare.GetStatus()
	if !status2.Enabled {
		t.Errorf("Expected state enabled")
	}
	if status2.HttpPort != defaultLANHttpPort {
		t.Errorf("Expected HTTP port %d, got %d", defaultLANHttpPort, status2.HttpPort)
	}

	// Disable
	err = globalLANShare.SetEnabled(false, opener, meter)
	if err != nil {
		t.Fatalf("Failed to disable LAN share: %v", err)
	}

	status3 := globalLANShare.GetStatus()
	if status3.Enabled {
		t.Errorf("Expected state disabled after stop")
	}
}
