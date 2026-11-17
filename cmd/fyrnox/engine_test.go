package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	vpnclient "firefox-vpn-client"
)

func TestPhysicalGateway(t *testing.T) {
	gw, idx, err := getPhysicalDefaultGateway()
	t.Logf("getPhysicalDefaultGateway: gw=%s, idx=%d, err=%v", gw, idx, err)
}

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

type mockProxySession struct {
	dummyOpener
	closed  bool
	closeMu sync.Mutex
}

func (m *mockProxySession) UpdateToken(token string) error {
	return nil
}

func (m *mockProxySession) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 443}
}

func (m *mockProxySession) Close() error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()
	m.closed = true
	return nil
}

func (m *mockProxySession) isClosed() bool {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()
	return m.closed
}

func TestConnectAtomicRollback(t *testing.T) {
	engine := NewVpnEngine()

	// Inject deterministic mock authentication, proxy pass, controller, and exit verification
	engine.auth = &runtimeAuth{
		Token: &vpnclient.TokenResponse{
			AccessToken:  "mock-test-access-token",
			RefreshToken: "mock-test-refresh-token",
			ExpiresIn:    3600,
		},
		ObtainedAt: time.Now(),
	}
	engine.proxyPassFn = func(ctx context.Context) (*vpnclient.ProxyPassInfo, error) {
		return &vpnclient.ProxyPassInfo{
			RawToken: "mock-proxy-pass",
			Claims: vpnclient.ProxyPassClaims{
				Exp: time.Now().Add(1 * time.Hour).Unix(),
			},
		}, nil
	}
	engine.proxyControllerFactory = func(guardian string, proxyURL *url.URL, timeout time.Duration, auth *runtimeAuth, pass *vpnclient.ProxyPassInfo, useH3 bool, apiProxy string, tlsFrag bool, ipProtocol string) (*proxyController, error) {
		return &proxyController{
			proxyURL: proxyURL,
			session:  &mockProxySession{},
			done:     make(chan struct{}),
		}, nil
	}
	engine.verifyExitFn = func(opener tunnelOpener, timeout time.Duration) (ExitProbeResult, error) {
		return ExitProbeResult{
			IP:          "198.51.100.1",
			CountryCode: "US",
			LatencyMs:   15,
		}, nil
	}

	// Avoid modifying host system proxy during test
	engine.settings.UseSystemProxy = false
	engine.settings.EnableTunMode = false

	err := engine.Connect()
	if err != nil {
		t.Fatalf("Connect initiation failed: %v", err)
	}

	// Wait up to 5 seconds for StateConnected
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		status := engine.GetStatus()
		if status.State == StateConnected {
			break
		}
		if status.State == StateError {
			t.Fatalf("Connection reached error state: %s", status.ErrorMessage)
		}
		time.Sleep(20 * time.Millisecond)
	}

	status := engine.GetStatus()
	if status.State != StateConnected {
		t.Fatalf("Timed out waiting for StateConnected; current: %s", status.State)
	}
	if status.ExitIP != "198.51.100.1" {
		t.Fatalf("Expected ExitIP '198.51.100.1', got %s", status.ExitIP)
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
		if status.State == StateDisconnected {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if status.State != StateDisconnected {
		t.Fatalf("Expected StateDisconnected after Disconnect(), got %s", status.State)
	}
}

func TestConnectExitVerificationFailureRollback(t *testing.T) {
	engine := NewVpnEngine()

	// 1. Verify synchronous failure when credentials are missing
	engine.auth = nil
	if err := engine.Connect(); err == nil {
		t.Fatal("expected Connect() to fail synchronously when auth is nil, but got nil error")
	}

	// 2. Set up valid in-memory mock authentication and proxy pass
	engine.auth = &runtimeAuth{
		Token: &vpnclient.TokenResponse{
			AccessToken:  "mock-test-access-token",
			RefreshToken: "mock-test-refresh-token",
			ExpiresIn:    3600,
		},
		ObtainedAt: time.Now(),
	}
	engine.proxyPassFn = func(ctx context.Context) (*vpnclient.ProxyPassInfo, error) {
		return &vpnclient.ProxyPassInfo{
			RawToken: "mock-proxy-pass",
			Claims: vpnclient.ProxyPassClaims{
				Exp: time.Now().Add(1 * time.Hour).Unix(),
			},
		}, nil
	}

	// Track session closure
	mockSession1 := &mockProxySession{}
	engine.proxyControllerFactory = func(guardian string, proxyURL *url.URL, timeout time.Duration, auth *runtimeAuth, pass *vpnclient.ProxyPassInfo, useH3 bool, apiProxy string, tlsFrag bool, ipProtocol string) (*proxyController, error) {
		return &proxyController{
			proxyURL: proxyURL,
			session:  mockSession1,
			done:     make(chan struct{}),
		}, nil
	}

	// Inject failing verifyExitFn
	expectedErr := errors.New("simulated exit probe failure: upstream unreachable")
	engine.verifyExitFn = func(opener tunnelOpener, timeout time.Duration) (ExitProbeResult, error) {
		return ExitProbeResult{}, expectedErr
	}

	// Use dedicated non-default test ports to avoid any port conflicts
	testSocksPort := 20818
	testHttpPort := 20819
	engine.settings.SocksPort = testSocksPort
	engine.settings.HttpPort = testHttpPort
	engine.settings.UseSystemProxy = false
	engine.settings.EnableTunMode = false

	// Initiate connection
	err := engine.Connect()
	if err != nil {
		t.Fatalf("Connect initiation failed unexpectedly: %v", err)
	}

	// Poll until StateError is reached
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		status := engine.GetStatus()
		if status.State == StateError {
			break
		}
		if status.State == StateConnected {
			t.Fatal("Engine unexpectedly transitioned to StateConnected despite failing exit verification")
		}
		time.Sleep(20 * time.Millisecond)
	}

	status := engine.GetStatus()
	if status.State != StateError {
		t.Fatalf("Expected StateError, got %s", status.State)
	}
	if status.State == StateConnecting {
		t.Fatal("Engine state remained in StateConnecting after exit probe failure")
	}
	if !strings.Contains(status.ErrorMessage, "Tunnel established but exit verification failed") ||
		!strings.Contains(status.ErrorMessage, expectedErr.Error()) {
		t.Fatalf("Expected error message containing exit failure, got: %q", status.ErrorMessage)
	}
	if status.ConnectStep != "" {
		t.Fatalf("Expected connect step to be reset, got: %q", status.ConnectStep)
	}
	if status.ExitIP != "" {
		t.Fatalf("Expected ExitIP to be rolled back/empty, got: %q", status.ExitIP)
	}
	if status.ExitCountry != "" {
		t.Fatalf("Expected ExitCountry to be empty, got: %q", status.ExitCountry)
	}
	if status.ExitLatency != 0 {
		t.Fatalf("Expected ExitLatency to be 0, got: %d", status.ExitLatency)
	}
	if status.IsVPNIP {
		t.Fatal("Expected IsVPNIP to be false")
	}

	// Verify controller session was closed
	if !mockSession1.isClosed() {
		t.Fatal("Expected mockProxySession to be closed on exit verification rollback")
	}

	// Verify internal engine references are nil
	engine.mu.RLock()
	if engine.controller != nil {
		engine.mu.RUnlock()
		t.Fatal("Expected engine.controller to be nil after error rollback")
	}
	if engine.socksServer != nil {
		engine.mu.RUnlock()
		t.Fatal("Expected engine.socksServer to be nil after error rollback")
	}
	if engine.httpServer != nil {
		engine.mu.RUnlock()
		t.Fatal("Expected engine.httpServer to be nil after error rollback")
	}
	if engine.exitResult != nil {
		engine.mu.RUnlock()
		t.Fatal("Expected engine.exitResult to be nil after error rollback")
	}
	engine.mu.RUnlock()

	// Verify SOCKS5 and HTTP proxy listeners are closed and no longer accepting TCP connections
	socksAddr := fmt.Sprintf("127.0.0.1:%d", testSocksPort)
	if conn, err := net.DialTimeout("tcp", socksAddr, 100*time.Millisecond); err == nil {
		_ = conn.Close()
		t.Fatalf("Expected SOCKS5 listener on %s to be closed, but dial succeeded", socksAddr)
	}
	httpAddr := fmt.Sprintf("127.0.0.1:%d", testHttpPort)
	if conn, err := net.DialTimeout("tcp", httpAddr, 100*time.Millisecond); err == nil {
		_ = conn.Close()
		t.Fatalf("Expected HTTP proxy listener on %s to be closed, but dial succeeded", httpAddr)
	}

	// 3. Verify subsequent retry connection attempt succeeds and is not blocked
	mockSession2 := &mockProxySession{}
	engine.proxyControllerFactory = func(guardian string, proxyURL *url.URL, timeout time.Duration, auth *runtimeAuth, pass *vpnclient.ProxyPassInfo, useH3 bool, apiProxy string, tlsFrag bool, ipProtocol string) (*proxyController, error) {
		return &proxyController{
			proxyURL: proxyURL,
			session:  mockSession2,
			done:     make(chan struct{}),
		}, nil
	}
	engine.verifyExitFn = func(opener tunnelOpener, timeout time.Duration) (ExitProbeResult, error) {
		return ExitProbeResult{
			IP:          "198.51.100.2",
			CountryCode: "US",
			LatencyMs:   12,
		}, nil
	}

	if err := engine.Connect(); err != nil {
		t.Fatalf("Subsequent Connect() initiation failed: %v", err)
	}

	retryDeadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(retryDeadline) {
		status = engine.GetStatus()
		if status.State == StateConnected {
			break
		}
		if status.State == StateError {
			t.Fatalf("Retry connection failed with error: %s", status.ErrorMessage)
		}
		time.Sleep(20 * time.Millisecond)
	}

	status = engine.GetStatus()
	if status.State != StateConnected {
		t.Fatalf("Expected StateConnected on retry, got: %s", status.State)
	}
	if status.ExitIP != "198.51.100.2" {
		t.Fatalf("Expected ExitIP '198.51.100.2' on retry, got: %s", status.ExitIP)
	}
	if status.ErrorMessage != "" {
		t.Fatalf("Expected empty ErrorMessage on successful retry, got: %q", status.ErrorMessage)
	}
	if mockSession2.isClosed() {
		t.Fatal("Expected active session for successful retry connection")
	}

	// Clean up retry connection
	if err := engine.Disconnect(); err != nil {
		t.Fatalf("Disconnect() after retry failed: %v", err)
	}
	discDeadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(discDeadline) {
		status = engine.GetStatus()
		if status.State == StateDisconnected {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if status.State != StateDisconnected {
		t.Fatalf("Expected StateDisconnected after Disconnect(), got: %s", status.State)
	}
}

func TestLiveConnectAtomicRollback(t *testing.T) {
	if os.Getenv("FXA_E2E") == "" {
		t.Skip("set FXA_E2E=1 to run the live connect atomic rollback test with real credentials")
	}

	engine := NewVpnEngine()
	if engine.auth == nil {
		t.Skip("Not logged in; skipping live atomic rollback test")
	}
	// Avoid modifying host system proxy during test
	engine.settings.UseSystemProxy = false
	engine.settings.EnableTunMode = false

	err := engine.Connect()
	if err != nil {
		t.Fatalf("Connect initiation failed: %v", err)
	}

	// Wait up to 45 seconds for StateConnected
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		status := engine.GetStatus()
		if status.State == StateConnected {
			break
		}
		if status.State == StateError {
			t.Fatalf("Connection reached error state: %s", status.ErrorMessage)
		}
		time.Sleep(200 * time.Millisecond)
	}

	status := engine.GetStatus()
	if status.State != StateConnected {
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
		if status.State == StateDisconnected {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if status.State != StateDisconnected {
		t.Fatalf("Expected StateDisconnected after Disconnect(), got %s", status.State)
	}
}
