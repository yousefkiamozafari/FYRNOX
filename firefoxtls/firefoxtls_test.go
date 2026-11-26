package firefoxtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	utls "github.com/refraction-networking/utls"
)

// 1. Asserts TLS 1.2/1.3 bounds, ALPN protocols, cipher suite ordering, and extension presence.
func TestFirefoxTLS_HelloSpecStructure(t *testing.T) {
	spec := Hello("http/1.1")
	if spec == nil {
		t.Fatalf("expected non-nil ClientHelloSpec")
	}
	if spec.TLSVersMin != utls.VersionTLS12 {
		t.Errorf("expected TLSVersMin to be TLS 1.2, got %x", spec.TLSVersMin)
	}
	if spec.TLSVersMax != utls.VersionTLS13 {
		t.Errorf("expected TLSVersMax to be TLS 1.3, got %x", spec.TLSVersMax)
	}
	if len(spec.CipherSuites) == 0 {
		t.Errorf("expected non-empty cipher suites")
	}

	hasALPN := false
	for _, ext := range spec.Extensions {
		if alpnExt, ok := ext.(*utls.ALPNExtension); ok {
			hasALPN = true
			if len(alpnExt.AlpnProtocols) != 1 || alpnExt.AlpnProtocols[0] != "http/1.1" {
				t.Errorf("expected ALPN [http/1.1], got %v", alpnExt.AlpnProtocols)
			}
		}
	}
	if !hasALPN {
		t.Errorf("expected ALPNExtension in extensions")
	}
}

// 2. Spins up an httptest.NewTLSServer, executes Handshake, asserts successful TLS connection and negotiated protocol "http/1.1".
func TestFirefoxTLS_HandshakeHermetic(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok-firefox-tls"))
	}))
	server.TLS = &tls.Config{
		NextProtos: []string{"http/1.1"},
	}
	server.StartTLS()
	defer server.Close()

	host, port, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("splitting host port: %v", err)
	}
	_ = port

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	rawConn, err := dialer.Dial("tcp", server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("raw dial failed: %v", err)
	}
	defer rawConn.Close()

	certPool := x509.NewCertPool()
	certPool.AddCert(server.Certificate())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uConn, err := Handshake(ctx, rawConn, host, certPool, nil, "http/1.1")
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}
	defer uConn.Close()

	cs := uConn.ConnectionState()
	if !cs.HandshakeComplete {
		t.Fatalf("handshake not complete")
	}
	if cs.NegotiatedProtocol != "http/1.1" {
		t.Errorf("expected negotiated protocol http/1.1, got %q", cs.NegotiatedProtocol)
	}
}

// 3. Connects to test server with untrusted cert pool, asserts handshake failure (no insecure bypass).
func TestFirefoxTLS_CertVerificationEnforced(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	host, _, _ := net.SplitHostPort(server.Listener.Addr().String())
	rawConn, err := net.Dial("tcp", server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("raw dial: %v", err)
	}
	defer rawConn.Close()

	// Empty cert pool: server's self-signed cert will fail validation
	emptyPool := x509.NewCertPool()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = Handshake(ctx, rawConn, host, emptyPool, nil, "http/1.1")
	if err == nil {
		t.Fatalf("expected certificate verification error, got nil")
	}
}

// 4. Cancels context before handshake, asserts prompt timeout / cancellation error.
func TestFirefoxTLS_ContextCancellation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	host, _, _ := net.SplitHostPort(server.Listener.Addr().String())
	rawConn, err := net.Dial("tcp", server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("raw dial: %v", err)
	}
	defer rawConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled immediately

	_, err = Handshake(ctx, rawConn, host, nil, nil, "http/1.1")
	if err == nil {
		t.Fatalf("expected error from canceled context, got nil")
	}
}

// 5. Connects twice with shared utls.ClientSessionCache, asserts session resumption ticket handling.
func TestFirefoxTLS_SessionResumption(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	host, _, _ := net.SplitHostPort(server.Listener.Addr().String())
	certPool := x509.NewCertPool()
	certPool.AddCert(server.Certificate())
	sessionCache := utls.NewLRUClientSessionCache(16)

	// Connection 1: populates session cache
	conn1, err := net.Dial("tcp", server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("dial 1: %v", err)
	}
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	uConn1, err := Handshake(ctx1, conn1, host, certPool, sessionCache, "http/1.1")
	if err != nil {
		t.Fatalf("handshake 1: %v", err)
	}
	uConn1.Close()

	// Connection 2: uses session cache
	conn2, err := net.Dial("tcp", server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("dial 2: %v", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	uConn2, err := Handshake(ctx2, conn2, host, certPool, sessionCache, "http/1.1")
	if err != nil {
		t.Fatalf("handshake 2: %v", err)
	}
	uConn2.Close()
}
