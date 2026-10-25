package main

import (
	"context"
	"net"
	"testing"
	"time"
)

type dummyOpener struct{}

func (d *dummyOpener) OpenTunnel(target string) (net.Conn, error) {
	c1, c2 := net.Pipe()
	go func() {
		defer c2.Close()
		buf := make([]byte, 1024)
		for {
			_, err := c2.Read(buf)
			if err != nil {
				return
			}
		}
	}()
	return c1, nil
}

func TestProxyServerCloseTerminatesActiveConnections(t *testing.T) {
	// Start HTTP Proxy
	meter := NewTrafficMeter()
	httpServer := NewHttpProxyServer("127.0.0.1:0", &dummyOpener{}, meter)
	if err := httpServer.Start(); err != nil {
		t.Fatalf("Failed to start HTTP proxy: %v", err)
	}

	addr := httpServer.listener.Addr().String()

	// Connect client
	clientConn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientConn.Close()

	// Send CONNECT request
	_, _ = clientConn.Write([]byte("CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n"))

	// Read the 200 Connection Established response
	clientConn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	buf := make([]byte, 512)
	_, _ = clientConn.Read(buf)

	start := time.Now()
	if err := httpServer.Close(); err != nil {
		t.Fatalf("HttpProxyServer.Close() error: %v", err)
	}

	duration := time.Since(start)
	if duration > 500*time.Millisecond {
		t.Errorf("HttpProxyServer.Close took too long: %v", duration)
	}

	// Verify client connection was terminated by server (subsequent read gives EOF or error)
	clientConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, readErr := clientConn.Read(buf)
	if readErr == nil {
		t.Errorf("Expected client read error after proxy close, got nil")
	}
}

func TestDisconnectSpeed(t *testing.T) {
	engine := &VpnEngine{
		state: StateConnected,
		meter: NewTrafficMeter(),
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- engine.Disconnect()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("engine.Disconnect() returned error: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("engine.Disconnect() timed out (> 2s)")
	}

	elapsed := time.Since(start)
	t.Logf("Disconnect completed cleanly in %v", elapsed)
	if elapsed > 1*time.Second {
		t.Errorf("Disconnect took too long: %v", elapsed)
	}
	if engine.state != StateDisconnected {
		t.Errorf("Expected state %s, got %s", StateDisconnected, engine.state)
	}
}
