package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestConnectionProgressSteps(t *testing.T) {
	engine := NewVpnEngine()
	status := engine.GetStatus()
	if status.ConnectStep != "" {
		t.Fatalf("expected empty initial connect step, got %q", status.ConnectStep)
	}
	engine.setConnectStep("authorizing")
	if engine.GetStatus().ConnectStep != "authorizing" {
		t.Fatalf("expected connect step 'authorizing', got %q", engine.GetStatus().ConnectStep)
	}
}

func TestLiveConnectStageByStage(t *testing.T) {
	engine := NewVpnEngine()
	if engine.auth == nil {
		t.Skip("Not logged in; skipping live connect test")
	}

	t.Logf("1. Testing getOrRenewProxyPass...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	start := time.Now()
	pass, err := engine.getOrRenewProxyPass(ctx)
	t.Logf("getOrRenewProxyPass took %v | Err: %v | Pass: %v", time.Since(start), err, pass != nil)
	if err != nil {
		t.Skipf("getOrRenewProxyPass failed (%v); skipping live probe", err)
		return
	}

	t.Logf("2. Testing newProxyController with server %s...", engine.selected.Hostname)
	targetURL, _ := url.Parse(fmt.Sprintf("https://%s:%d", engine.selected.Hostname, engine.selected.Port))
	start2 := time.Now()
	controller, err := newProxyController(engine.guardian, targetURL, 15*time.Second, engine.auth, pass, engine.settings.UseH3, engine.settings.ApiProxy, engine.settings.TlsFragmentation, engine.settings.IPProtocol)
	t.Logf("newProxyController took %v | Err: %v", time.Since(start2), err)
	if err != nil {
		t.Skipf("Upstream server %s unreachable (%v); skipping live probe", engine.selected.Hostname, err)
		return
	}
	defer controller.Close()

	t.Logf("3. Testing VerifyExit probe...")
	start3 := time.Now()
	probe, err := VerifyExit(controller, 10*time.Second)
	t.Logf("VerifyExit took %v | Err: %v | Probe: %+v", time.Since(start3), err, probe)
}

func TestEngineInitializesWithRealServers(t *testing.T) {
	engine := NewVpnEngine()
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}

	servers := engine.GetServers()
	if len(servers) < 30 {
		t.Fatalf("expected at least 30 real servers, got %d", len(servers))
	}

	for _, s := range servers {
		if strings.HasPrefix(s.Hostname, "fra1.") || strings.HasPrefix(s.Hostname, "ams1.") || strings.HasPrefix(s.Hostname, "lhr1.") {
			t.Fatalf("found invalid placeholder hostname in server list: %s", s.Hostname)
		}
		if !strings.HasSuffix(s.Hostname, ".m1.fastly-masque.net") && !strings.Contains(s.Hostname, ".") {
			t.Fatalf("unexpected server hostname format: %s", s.Hostname)
		}
		if s.CountryName == "" || s.CountryCode == "" {
			t.Fatalf("server missing country metadata: %+v", s)
		}
	}

	status := engine.GetStatus()
	if status.Server == nil {
		t.Fatal("expected selected server to be non-nil")
	}
	if status.Server.Hostname == "" {
		t.Fatal("selected server hostname is empty")
	}
	if strings.HasPrefix(status.Server.Hostname, "fra1.") {
		t.Fatalf("selected server is invalid placeholder: %s", status.Server.Hostname)
	}
}

func TestProxyTLSConfig(t *testing.T) {
	cfg := createProxyTLSConfig("eddf247.m1.fastly-masque.net", []string{"h2"})
	if cfg == nil {
		t.Fatal("expected non-nil tls.Config")
	}
	if cfg.ServerName != "eddf247.m1.fastly-masque.net" {
		t.Fatalf("expected ServerName eddf247.m1.fastly-masque.net, got %s", cfg.ServerName)
	}
	if len(cfg.NextProtos) != 1 || cfg.NextProtos[0] != "h2" {
		t.Fatalf("expected NextProtos ['h2'], got %v", cfg.NextProtos)
	}
}

func TestFragmentingConnPreservesDataAndFragments(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	fc := newFragmentingConn(client, "sni")

	payload := make([]byte, 100)
	payload[0] = 0x16 // Handshake
	payload[1] = 0x03 // TLS 1.x
	payload[2] = 0x01
	payload[3] = 0x00
	payload[4] = 95
	payload[5] = 0x01 // ClientHello
	for i := 6; i < 100; i++ {
		payload[i] = byte(i)
	}

	received := make([]byte, 100)
	errCh := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(server, received)
		errCh <- err
	}()

	n, err := fc.Write(payload)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("Expected wrote %d bytes, got %d", len(payload), n)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("ReadFull failed: %v", err)
	}

	if !bytes.Equal(payload, received) {
		t.Fatalf("Payload corrupted during fragmentation!")
	}
}

func TestEngineStatusIncludesCurrentIP(t *testing.T) {
	engine := NewVpnEngine()
	status := engine.GetStatus()
	if status.CurrentIP == "" {
		t.Fatal("expected status.CurrentIP to be non-empty")
	}
}

func TestConnectAtomicRollback(t *testing.T) {
	engine := NewVpnEngine()
	if engine.auth == nil {
		t.Skip("Not logged in; skipping atomic rollback test")
	}
	// Avoid modifying host system proxy during test
	engine.settings.UseSystemProxy = false
	engine.settings.EnableTunMode = false

	err := engine.Connect()
	if err != nil {
		t.Fatalf("Connect initiation failed: %v", err)
	}

	// Wait up to 12 seconds for StateConnected
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		status := engine.GetStatus()
		if status.State == "connected" {
			break
		}
		if status.State == "error" {
			t.Fatalf("Connection reached error state: %s", status.ErrorMessage)
		}
		time.Sleep(200 * time.Millisecond)
	}

	status := engine.GetStatus()
	if status.State != "connected" {
		t.Fatalf("Timed out waiting for StateConnected; current: %s", status.State)
	}
	if status.ExitIP == "" {
		t.Fatalf("Expected non-empty ExitIP on connected status")
	}

	// Now test atomic disconnect
	err = engine.Disconnect()
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	// Verify status is disconnected within short grace period
	deadlineDisc := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadlineDisc) {
		status = engine.GetStatus()
		if status.State == "disconnected" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if status.State != "disconnected" {
		t.Fatalf("Expected StateDisconnected after Disconnect(), got %s", status.State)
	}
}
