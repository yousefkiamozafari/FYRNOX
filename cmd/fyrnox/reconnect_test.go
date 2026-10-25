package main

import (
	"testing"
	"time"
)

func TestSelectServerHotReconnect(t *testing.T) {
	engine := NewVpnEngine()
	if len(engine.servers) < 2 {
		t.Fatal("Need at least 2 servers for switching test")
	}

	firstServer := engine.servers[0].Hostname
	secondServer := engine.servers[1].Hostname

	// Select first server initially
	err := engine.SelectServer(firstServer)
	if err != nil {
		t.Fatalf("Failed to select initial server: %v", err)
	}
	if engine.selected == nil || engine.selected.Hostname != firstServer {
		t.Fatalf("Expected selected server %s, got %v", firstServer, engine.selected)
	}

	// Simulate connected state
	engine.mu.Lock()
	engine.state = StateConnected
	engine.mu.Unlock()

	// Switch to second server while connected
	err = engine.SelectServer(secondServer)
	if err != nil {
		t.Fatalf("Failed to select second server: %v", err)
	}

	// Verify selected server updated immediately
	if engine.selected == nil || engine.selected.Hostname != secondServer {
		t.Fatalf("Expected selected server %s, got %v", secondServer, engine.selected)
	}

	// Verify that state immediately transitioned or initiated hot reconnect
	time.Sleep(50 * time.Millisecond)
	status := engine.GetStatus()
	t.Logf("Status after hot reconnect switch: state=%s, server=%+v", status.State, status.Server)
	if status.Server == nil || status.Server.Hostname != secondServer {
		t.Fatalf("Expected Server.Hostname %s, got %+v", secondServer, status.Server)
	}
}
