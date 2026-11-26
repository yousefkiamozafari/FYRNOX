package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type mockProbeSession struct {
	dialErr    error
	dialConn   net.Conn
	dialTarget string
}

func (m *mockProbeSession) OpenTunnel(authority string) (net.Conn, error) {
	m.dialTarget = authority
	if m.dialErr != nil {
		return nil, m.dialErr
	}
	if m.dialConn != nil {
		return m.dialConn, nil
	}
	c1, c2 := net.Pipe()
	go func() { _ = c2.Close() }()
	return c1, nil
}

func (m *mockProbeSession) Close() error { return nil }

func TestIPv6Probe_DialFailureFailsClosed(t *testing.T) {
	ctrl := &proxyController{
		session: &mockProbeSession{dialErr: errors.New("masque dial error")},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := ctrl.ProbeIPv6DataPath(ctx)
	if err == nil {
		t.Fatal("Expected error on masque dial failure, got nil")
	}
	if ctrl.IsIPv6Operational(ctx) {
		t.Fatal("IsIPv6Operational should return false on dial error")
	}
}

func TestIPv6Probe_TLSHandshakeFailureFailsClosed(t *testing.T) {
	c1, c2 := net.Pipe()
	go func() {
		// Server closes without completing TLS handshake
		_ = c2.Close()
	}()

	ctrl := &proxyController{
		session: &mockProbeSession{dialConn: c1},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := ctrl.ProbeIPv6DataPath(ctx)
	if err == nil {
		t.Fatal("Expected error on broken TLS handshake, got nil")
	}
	if ctrl.IsIPv6Operational(ctx) {
		t.Fatal("IsIPv6Operational should return false on TLS handshake failure")
	}
}

func TestIPv6Probe_TargetEndpointIsHTTPS(t *testing.T) {
	sess := &mockProbeSession{dialErr: errors.New("halt")}
	ctrl := &proxyController{session: sess}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_ = ctrl.ProbeIPv6DataPath(ctx)

	expectedTarget := "[2606:4700:4700::1111]:443"
	if sess.dialTarget != expectedTarget {
		t.Fatalf("Expected target %s, got %s", expectedTarget, sess.dialTarget)
	}
}
