package main

import (
	"net"
	"testing"
	"time"
)

func TestTrafficMeterStatsAndMapping(t *testing.T) {
	meter := NewTrafficMeter()
	meter.Reset()

	stats := meter.GetStats()
	if stats.BytesDown != 0 || stats.BytesUp != 0 {
		t.Fatalf("expected 0 bytes, got down=%d up=%d", stats.BytesDown, stats.BytesUp)
	}
	if stats.TotalDownHuman != "0 B" || stats.BytesReadHuman != "0 B" {
		t.Fatalf("expected '0 B' for initial human stats, got %q / %q", stats.TotalDownHuman, stats.BytesReadHuman)
	}

	// Simulate transferring 10.5 MB down and 2.5 MB up
	meter.AddDown(10*1024*1024 + 512*1024)
	meter.AddUp(2*1024*1024 + 512*1024)

	stats = meter.GetStats()
	if stats.BytesDown <= 0 || stats.BytesUp <= 0 {
		t.Fatalf("expected positive byte counters, got down=%d up=%d", stats.BytesDown, stats.BytesUp)
	}
	if stats.TotalDownHuman == "0 B" || stats.BytesReadHuman == "0 B" {
		t.Fatalf("expected formatted human strings, got %q / %q", stats.TotalDownHuman, stats.BytesReadHuman)
	}
	if stats.TotalDownHuman != stats.BytesReadHuman {
		t.Fatalf("expected TotalDownHuman == BytesReadHuman, got %q vs %q", stats.TotalDownHuman, stats.BytesReadHuman)
	}
	if stats.TotalUpHuman != stats.BytesWrittenHuman {
		t.Fatalf("expected TotalUpHuman == BytesWrittenHuman, got %q vs %q", stats.TotalUpHuman, stats.BytesWrittenHuman)
	}

	t.Logf("Traffic stats verified: Down=%s (%d bytes), Up=%s (%d bytes)", stats.TotalDownHuman, stats.BytesDown, stats.TotalUpHuman, stats.BytesUp)
}

func TestTrafficMeterSessionLifecycle(t *testing.T) {
	meter := NewTrafficMeter()

	// 1. Initial state before session
	stats := meter.GetStats()
	if stats.BytesDown != 0 || stats.BytesUp != 0 || stats.DurationSec != 0 {
		t.Fatalf("initial stats should be 0, got down=%d, up=%d, dur=%d", stats.BytesDown, stats.BytesUp, stats.DurationSec)
	}

	// 2. Start session
	meter.StartSession()
	time.Sleep(10 * time.Millisecond)
	meter.AddDown(245*1024*1024 + 800*1024) // ~245.8 MB
	meter.AddUp(18*1024*1024 + 400*1024)   // ~18.4 MB

	activeStats := meter.GetStats()
	if activeStats.BytesDown <= 0 || activeStats.BytesUp <= 0 {
		t.Fatalf("active session bytes should be positive")
	}
	if activeStats.TotalDownHuman == "0 B" || activeStats.TotalUpHuman == "0 B" {
		t.Fatalf("active session formatted totals should not be '0 B'")
	}

	// 3. Stop session (disconnection)
	meter.StopSession()
	frozenStats := meter.GetStats()

	// Totals MUST be preserved upon disconnection
	if frozenStats.BytesDown != activeStats.BytesDown || frozenStats.BytesUp != activeStats.BytesUp {
		t.Fatalf("stopped session must preserve totals: got down=%d vs %d", frozenStats.BytesDown, activeStats.BytesDown)
	}
	if frozenStats.TotalDownHuman != activeStats.TotalDownHuman || frozenStats.TotalUpHuman != activeStats.TotalUpHuman {
		t.Fatalf("stopped session must preserve human totals: got down=%s vs %s", frozenStats.TotalDownHuman, activeStats.TotalDownHuman)
	}
	if frozenStats.SpeedDown != 0 || frozenStats.SpeedUp != 0 {
		t.Fatalf("stopped session must have 0 speed: got down=%d, up=%d", frozenStats.SpeedDown, frozenStats.SpeedUp)
	}

	// 4. Start new session (reconnection)
	meter.StartSession()
	newStats := meter.GetStats()
	if newStats.BytesDown != 0 || newStats.BytesUp != 0 {
		t.Fatalf("new session must reset totals to 0: got down=%d, up=%d", newStats.BytesDown, newStats.BytesUp)
	}
}

func TestProxyTrafficRealIOAccounting(t *testing.T) {
	meter := NewTrafficMeter()
	meter.StartSession()

	httpServer := NewHttpProxyServer("127.0.0.1:0", &dummyOpener{}, meter)
	if err := httpServer.Start(); err != nil {
		t.Fatalf("failed to start http proxy: %v", err)
	}
	defer httpServer.Close()

	addr := httpServer.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to dial http proxy: %v", err)
	}
	defer conn.Close()

	// Send CONNECT request
	connectReq := "CONNECT test.example.com:443 HTTP/1.1\r\nHost: test.example.com:443\r\n\r\n"
	if _, err := conn.Write([]byte(connectReq)); err != nil {
		t.Fatalf("failed to write connect request: %v", err)
	}

	// Read 200 response
	resp := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _ = conn.Read(resp)

	// Send dummy payload
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	_, _ = conn.Write(payload)
	time.Sleep(50 * time.Millisecond)

	stats := meter.GetStats()
	if stats.BytesUp == 0 {
		t.Fatalf("expected BytesUp > 0 after real socket transfer, got %d", stats.BytesUp)
	}
	t.Logf("Real IO Proxy Accounting Verified: BytesUp=%d, BytesDown=%d, TotalUpHuman=%s",
		stats.BytesUp, stats.BytesDown, stats.TotalUpHuman)
}
