package vpnclient

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// closeIdleConnections closes idle connections on any transport that supports it.
func closeIdleConnections(t http.RoundTripper) {
	type idleCloser interface{ CloseIdleConnections() }
	if ic, ok := t.(idleCloser); ok {
		ic.CloseIdleConnections()
	}
}

// Helper: spins up a mock TLS server with a given handler and cert pool
func setupMockTLSServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *x509.CertPool) {
	server := httptest.NewUnstartedServer(handler)
	server.TLS = &tls.Config{
		NextProtos: []string{"http/1.1"},
	}
	server.StartTLS()
	t.Cleanup(server.Close)

	certPool := x509.NewCertPool()
	certPool.AddCert(server.Certificate())
	return server, certPool
}

// 1. Direct connection: NewControlPlaneTransport establishes TLS and negotiates HTTP/1.1
func TestTransportFactory_DirectConnection(t *testing.T) {
	server, certPool := setupMockTLSServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Proto", r.Proto)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("direct-ok"))
	})

	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		Roots: certPool,
	})
	if err != nil {
		t.Fatalf("NewControlPlaneTransport failed: %v", err)
	}
	defer closeIdleConnections(tr)

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "direct-ok" {
		t.Errorf("unexpected body: %s", body)
	}
	if resp.Header.Get("X-Proto") != "HTTP/1.1" {
		t.Errorf("expected proto HTTP/1.1, got %s", resp.Header.Get("X-Proto"))
	}
}

// 2. Custom dialer: Transport invokes custom dialer and wraps it with uTLS
func TestTransportFactory_CustomDialer(t *testing.T) {
	server, certPool := setupMockTLSServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("custom-dialer-ok"))
	})

	customDialerCalled := false
	stdDialer := &net.Dialer{Timeout: 5 * time.Second}

	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		Roots: certPool,
		CustomDialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			customDialerCalled = true
			return stdDialer.DialContext(ctx, network, addr)
		},
	})
	if err != nil {
		t.Fatalf("NewControlPlaneTransport failed: %v", err)
	}
	defer closeIdleConnections(tr)

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if !customDialerCalled {
		t.Errorf("custom dialer was not called")
	}
}

// 3. SOCKS5 proxy: Transport dials through mock SOCKS5 proxy and wraps with uTLS
func TestTransportFactory_SOCKS5Proxy(t *testing.T) {
	server, certPool := setupMockTLSServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("socks5-ok"))
	})

	// Minimal hermetic mock SOCKS5 proxy listener
	socksListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen socks: %v", err)
	}
	defer socksListener.Close()

	go func() {
		for {
			conn, err := socksListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 256)
				// 1. Handshake: read version and methods
				n, err := c.Read(buf)
				if err != nil || n < 2 || buf[0] != 0x05 {
					return
				}
				// Reply: SOCKS5, no auth (0x00)
				c.Write([]byte{0x05, 0x00})

				// 2. Request: CONNECT
				n, err = c.Read(buf)
				if err != nil || n < 7 || buf[1] != 0x01 {
					return
				}

				// Resolve target address from SOCKS5 request
				var targetAddr string
				switch buf[3] {
				case 0x01: // IPv4
					ip := net.IP(buf[4:8])
					port := int(buf[8])<<8 | int(buf[9])
					targetAddr = fmt.Sprintf("%s:%d", ip.String(), port)
				case 0x03: // Domain
					domainLen := int(buf[4])
					domain := string(buf[5 : 5+domainLen])
					port := int(buf[5+domainLen])<<8 | int(buf[6+domainLen])
					targetAddr = fmt.Sprintf("%s:%d", domain, port)
				}

				targetConn, err := net.DialTimeout("tcp", targetAddr, 5*time.Second)
				if err != nil {
					c.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
					return
				}
				defer targetConn.Close()

				// Success reply
				c.Write([]byte{0x05, 0x00, 0x00, 0x01, 127, 0, 0, 1, 0, 0})

				// Bidirectional copy
				done := make(chan struct{})
				go func() { io.Copy(targetConn, c); close(done) }()
				io.Copy(c, targetConn)
				<-done
			}(conn)
		}
	}()

	proxyURL, _ := url.Parse("socks5://" + socksListener.Addr().String())
	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		ProxyURL: proxyURL,
		Roots:    certPool,
	})
	if err != nil {
		t.Fatalf("NewControlPlaneTransport failed: %v", err)
	}
	defer closeIdleConnections(tr)

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request through SOCKS5 failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "socks5-ok" {
		t.Errorf("unexpected body: %s", body)
	}
}

// 4. HTTP CONNECT proxy: Transport performs CONNECT handshake, then executes uTLS over the tunnel
func TestTransportFactory_HTTPCONNECTProxy(t *testing.T) {
	server, certPool := setupMockTLSServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("http-connect-ok"))
	})

	// Minimal hermetic mock HTTP CONNECT proxy listener
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen proxy: %v", err)
	}
	defer proxyListener.Close()

	connectSeen := false
	go func() {
		for {
			conn, err := proxyListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				br := bufio.NewReader(c)
				req, err := http.ReadRequest(br)
				if err != nil {
					return
				}
				if req.Method != http.MethodConnect {
					c.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
					return
				}
				connectSeen = true
				targetConn, err := net.DialTimeout("tcp", req.RequestURI, 5*time.Second)
				if err != nil {
					c.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
					return
				}
				defer targetConn.Close()

				c.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

				done := make(chan struct{})
				go func() { io.Copy(targetConn, br); close(done) }()
				io.Copy(c, targetConn)
				<-done
			}(conn)
		}
	}()

	proxyURL, _ := url.Parse("http://" + proxyListener.Addr().String())
	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		ProxyURL: proxyURL,
		Roots:    certPool,
	})
	if err != nil {
		t.Fatalf("NewControlPlaneTransport failed: %v", err)
	}
	defer closeIdleConnections(tr)

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request through HTTP CONNECT proxy failed: %v", err)
	}
	defer resp.Body.Close()

	if !connectSeen {
		t.Errorf("CONNECT request was not received by mock proxy")
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "http-connect-ok" {
		t.Errorf("unexpected body: %s", body)
	}
}

// 5. TLS fragmentation compatibility: Transport wraps connection with fragmenting conn before uTLS
func TestTransportFactory_TLSFragmentation(t *testing.T) {
	server, certPool := setupMockTLSServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("frag-ok"))
	})

	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		Roots:       certPool,
		TLSFragment: true,
	})
	if err != nil {
		t.Fatalf("NewControlPlaneTransport failed: %v", err)
	}
	defer closeIdleConnections(tr)

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request with TLS fragmentation failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "frag-ok" {
		t.Errorf("unexpected body: %s", body)
	}
}

// 6. Environment proxy immunity: Default transport has Proxy: nil and ignores HTTP_PROXY
func TestTransportFactory_IgnoresEnvironmentProxy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:59999")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:59999")

	server, certPool := setupMockTLSServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("env-ignored-ok"))
	})

	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		Roots: certPool,
	})
	if err != nil {
		t.Fatalf("NewControlPlaneTransport failed: %v", err)
	}
	defer closeIdleConnections(tr)

	// When ForceStandardTLS is false and no SOCKS/CustomDialer, Windows may return
	// winHTTPTransport. On all platforms the resulting transport must not silently
	// honour the ambient HTTP_PROXY environment variable.
	// For *http.Transport we can verify the Proxy field directly; for winHTTPTransport
	// the guarantee is structural (WINHTTP_ACCESS_TYPE_NO_PROXY is set at session open).
	if htr, ok := tr.(*http.Transport); ok {
		if htr.Proxy != nil {
			t.Errorf("expected tr.Proxy to be nil on *http.Transport")
		}
	}

	client := &http.Client{Transport: tr, Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request should succeed by ignoring invalid environment proxy, got: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "env-ignored-ok" {
		t.Errorf("unexpected body: %s", body)
	}
}
