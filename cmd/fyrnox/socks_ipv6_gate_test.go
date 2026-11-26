package main

import (
	"errors"
	"net"
	"testing"
	"time"
)

type mockTunnelOpener struct {
	openCount  int
	lastTarget string
}

func (m *mockTunnelOpener) OpenTunnel(target string) (net.Conn, error) {
	m.openCount++
	m.lastTarget = target
	c1, c2 := net.Pipe()
	go func() {
		_ = c2.Close()
	}()
	return c1, nil
}

func TestSocksIPv6Gate_BlockedWhenNotOperational(t *testing.T) {
	opener := &mockTunnelOpener{}
	server := NewSocksProxyServer("127.0.0.1:0", opener, nil)
	server.SetIPv6ForwardingAllowed(false) // blocked

	// 1. Test direct dialUpstream
	_, err := server.dialUpstream("[2606:4700:4700::1111]:443", 12345)
	if err == nil {
		t.Fatal("Expected error when dialUpstream called for IPv6 with ipv6Allowed=false")
	}
	if opener.openCount != 0 {
		t.Fatalf("OpenTunnel should not be called when IPv6 is blocked, got openCount=%d", opener.openCount)
	}

	// 2. Test bypass target with IPv6
	server.SetBypassRules("<local>;*.example.com")
	_, err = server.dialUpstream("[2606:4700:4700::1111]:443", 12345)
	if err == nil {
		t.Fatal("Expected error when bypass target is IPv6 with ipv6Allowed=false")
	}
	if opener.openCount != 0 {
		t.Fatalf("OpenTunnel should not be called, got openCount=%d", opener.openCount)
	}
}

func TestSocksIPv6Gate_AllowedWhenOperational(t *testing.T) {
	opener := &mockTunnelOpener{}
	server := NewSocksProxyServer("127.0.0.1:0", opener, nil)
	server.SetIPv6ForwardingAllowed(true) // operational

	conn, err := server.dialUpstream("[2606:4700:4700::1111]:443", 12345)
	if err != nil {
		t.Fatalf("Expected dialUpstream to succeed for IPv6 when allowed, got error: %v", err)
	}
	if conn != nil {
		_ = conn.Close()
	}
	if opener.openCount != 1 {
		t.Fatalf("Expected OpenTunnel to be called once, got %d", opener.openCount)
	}
	if opener.lastTarget != "[2606:4700:4700::1111]:443" {
		t.Fatalf("Expected target [2606:4700:4700::1111]:443, got %s", opener.lastTarget)
	}
}

func TestSocksIPv6Gate_IPv4UnblockedRegardless(t *testing.T) {
	opener := &mockTunnelOpener{}
	server := NewSocksProxyServer("127.0.0.1:0", opener, nil)
	server.SetIPv6ForwardingAllowed(false) // IPv6 blocked

	conn, err := server.dialUpstream("1.1.1.1:443", 12345)
	if err != nil {
		t.Fatalf("IPv4 dialUpstream should succeed even when IPv6 is blocked, got error: %v", err)
	}
	if conn != nil {
		_ = conn.Close()
	}
	if opener.openCount != 1 {
		t.Fatalf("Expected OpenTunnel to be called for IPv4, got %d", opener.openCount)
	}
}

func TestSocksIPv6Gate_ActiveConnectionsClosedOnTransition(t *testing.T) {
	server := NewSocksProxyServer("127.0.0.1:0", &mockTunnelOpener{}, nil)
	server.SetIPv6ForwardingAllowed(true)

	// Create a pipe and register it in server.conns with an IPv6 remote address
	clientPipe, serverPipe := net.Pipe()
	defer clientPipe.Close()

	// Mock a net.Conn with an IPv6 RemoteAddr
	mockConn := &mockIPv6Conn{
		Conn:       serverPipe,
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("2606:4700:4700::1111"), Port: 443},
	}

	server.mu.Lock()
	server.conns[mockConn] = struct{}{}
	server.mu.Unlock()

	// Transition to blocked
	server.SetIPv6ForwardingAllowed(false)

	// Verify that reading from clientPipe fails because serverPipe was closed
	buf := make([]byte, 10)
	_ = clientPipe.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, err := clientPipe.Read(buf)
	if err == nil {
		t.Fatal("Expected read error on closed pipe after IPv6 gate closed active connection")
	}
}

type mockIPv6Conn struct {
	net.Conn
	remoteAddr net.Addr
}

func (m *mockIPv6Conn) RemoteAddr() net.Addr {
	return m.remoteAddr
}

func (m *mockIPv6Conn) Close() error {
	if m.Conn != nil {
		return m.Conn.Close()
	}
	return errors.New("closed")
}
