package main

import (
	"testing"
)

func TestGamingAcceleratorMetrics(t *testing.T) {
	gamer := &GamingAccelerator{
		historyRTT: make([]int, 0, 20),
	}

	gamer.recordProbe(40)
	gamer.recordProbe(44)
	gamer.recordProbe(42)

	stats := gamer.GetStats()
	if stats.PingMs != 42 {
		t.Errorf("Expected last ping 42, got %d", stats.PingMs)
	}
	if stats.JitterMs <= 0 {
		t.Errorf("Expected non-zero jitter, got %d", stats.JitterMs)
	}
	if stats.PacketLoss != 0 {
		t.Errorf("Expected 0%% packet loss, got %f", stats.PacketLoss)
	}

	// Record a lost packet (-1)
	gamer.recordProbe(-1)
	statsLost := gamer.GetStats()
	if statsLost.PacketLoss <= 0 {
		t.Errorf("Expected positive packet loss after failure, got %f", statsLost.PacketLoss)
	}
}
