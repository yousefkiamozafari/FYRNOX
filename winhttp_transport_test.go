//go:build windows

package vpnclient

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func setupMockHTTPServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func newWinHTTPTestClient(t *testing.T, opts ControlPlaneTransportOptions) (*http.Client, *winHTTPTransport) {
	t.Helper()
	rt, err := newWinHTTPTransport(opts)
	if err != nil {
		t.Fatalf("newWinHTTPTransport: %v", err)
	}
	wt := rt.(*winHTTPTransport)
	t.Cleanup(wt.CloseIdleConnections)
	return &http.Client{Transport: wt, Timeout: 10 * time.Second}, wt
}

func mustParseTestURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func TestWinHTTP_RoutingMatrix(t *testing.T) {
	cases := []struct {
		name string
		opts ControlPlaneTransportOptions
		want bool
	}{
		{
			name: "Direct (no proxy, no dialer, no frag, no custom roots)",
			opts: ControlPlaneTransportOptions{},
			want: true,
		},
		{
			name: "HTTP forward proxy",
			opts: ControlPlaneTransportOptions{ProxyURL: mustParseTestURL("http://proxy.example.com:8080")},
			want: true,
		},
		{
			name: "HTTPS forward proxy — stays on Go transport",
			opts: ControlPlaneTransportOptions{ProxyURL: mustParseTestURL("https://proxy.example.com:8443")},
			want: false,
		},
		{
			name: "SOCKS5 proxy — must NOT use WinHTTP",
			opts: ControlPlaneTransportOptions{ProxyURL: mustParseTestURL("socks5://127.0.0.1:1080")},
			want: false,
		},
		{
			name: "SOCKS5h proxy — must NOT use WinHTTP",
			opts: ControlPlaneTransportOptions{ProxyURL: mustParseTestURL("socks5h://127.0.0.1:1080")},
			want: false,
		},
		{
			name: "Custom dialer — must NOT use WinHTTP",
			opts: ControlPlaneTransportOptions{
				CustomDialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return nil, nil
				},
			},
			want: false,
		},
		{
			name: "TLS fragmentation — must NOT use WinHTTP",
			opts: ControlPlaneTransportOptions{TLSFragment: true},
			want: false,
		},
		{
			name: "ForceStandardTLS — must NOT use WinHTTP",
			opts: ControlPlaneTransportOptions{ForceStandardTLS: true},
			want: false,
		},
		{
			name: "Custom Roots CertPool — must NOT use WinHTTP",
			opts: ControlPlaneTransportOptions{Roots: x509.NewCertPool()},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldUseWinHTTP(tc.opts); got != tc.want {
				t.Errorf("shouldUseWinHTTP(%+v) = %v, want %v", tc.opts, got, tc.want)
			}
			rt, err := NewControlPlaneTransport(tc.opts)
			if err != nil {
				t.Fatalf("NewControlPlaneTransport: %v", err)
			}
			if ic, ok := rt.(interface{ CloseIdleConnections() }); ok {
				defer ic.CloseIdleConnections()
			}
			if got := IsWinHTTPTransport(rt); got != tc.want {
				t.Errorf("IsWinHTTPTransport = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWinHTTP_DirectGETAndHeaders(t *testing.T) {
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Method", r.Method)
		w.Header().Add("X-Multi", "v1")
		w.Header().Add("X-Multi", "v2")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "winhttp-get-ok")
	})

	client, _ := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test?a=1", nil)
	req.Header.Add("X-Custom", "one")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify zero-length read guard never panics.
	if n, rerr := resp.Body.Read(nil); n != 0 || rerr != nil {
		t.Fatalf("Read(nil) = (%d, %v), want (0, nil)", n, rerr)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "winhttp-get-ok" {
		t.Errorf("unexpected body: %q", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %d", resp.StatusCode)
	}
	if resp.ContentLength != int64(len("winhttp-get-ok")) {
		t.Errorf("ContentLength = %d, want %d", resp.ContentLength, len("winhttp-get-ok"))
	}
	if got := resp.Header.Values("X-Multi"); len(got) != 2 {
		t.Errorf("X-Multi values = %v, want 2 entries", got)
	}
}

func TestWinHTTP_DirectPOST(t *testing.T) {
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	})

	client, _ := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})
	payload := `{"client_id":"test","grant_type":"refresh_token"}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/create", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("unexpected status: %d", resp.StatusCode)
	}
	if string(body) != payload {
		t.Errorf("unexpected echo body: %q", string(body))
	}
}

func TestWinHTTP_HTTPProxyAndAuth(t *testing.T) {
	targetSrv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "proxied-target-ok")
	})
	targetParsed, _ := url.Parse(targetSrv.URL)

	var proxyHits atomic.Int32

	proxySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHits.Add(1)

		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:s3cr3t"))
		if r.Header.Get("Proxy-Authorization") != expected {
			w.Header().Set("Proxy-Authenticate", `Basic realm="test"`)
			w.WriteHeader(http.StatusProxyAuthRequired)
			return
		}

		// Forward to the local target server regardless of test hostname.
		fwdURL := "http://" + targetParsed.Host + r.URL.RequestURI()
		req, err := http.NewRequest(r.Method, fwdURL, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}))
	t.Cleanup(proxySrv.Close)

	pu, _ := url.Parse(proxySrv.URL)
	pu.User = url.UserPassword("alice", "s3cr3t")

	client, _ := newWinHTTPTestClient(t, ControlPlaneTransportOptions{
		ProxyURL: pu,
	})

	// Use both a non-loopback unresolvable host (to prove DNS doesn't run locally and
	// traffic routes via the HTTP proxy) and verify Basic Proxy-Authorization succeeded.
	resp, err := client.Get("http://fyrnox-proxy-target.invalid/check")
	if err != nil {
		t.Fatalf("GET via HTTP proxy failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "proxied-target-ok" {
		t.Fatalf("unexpected proxy response: %d %q", resp.StatusCode, string(body))
	}
	if proxyHits.Load() == 0 {
		t.Fatalf("expected request to traverse mock HTTP proxy")
	}
}

func TestWinHTTP_DisablesInternalCookiesAndRedirects(t *testing.T) {
	var sawCookie atomic.Bool
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "" {
			sawCookie.Store(true)
		}
		switch r.URL.Path {
		case "/setcookie":
			http.SetCookie(w, &http.Cookie{Name: "winhttp_leak", Value: "1", Path: "/"})
			http.Redirect(w, r, "/after", http.StatusFound)
		default:
			w.WriteHeader(http.StatusOK)
		}
	})

	_, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})

	req1, _ := http.NewRequest(http.MethodGet, srv.URL+"/setcookie", nil)
	resp1, err := wt.RoundTrip(req1)
	if err != nil {
		t.Fatalf("RoundTrip 1: %v", err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusFound {
		t.Fatalf("expected raw 302 from RoundTrip (no auto redirect), got %d", resp1.StatusCode)
	}

	req2, _ := http.NewRequest(http.MethodGet, srv.URL+"/after", nil)
	resp2, err := wt.RoundTrip(req2)
	if err != nil {
		t.Fatalf("RoundTrip 2: %v", err)
	}
	resp2.Body.Close()
	if sawCookie.Load() {
		t.Fatalf("WinHTTP leaked internal session cookie to subsequent request")
	}
}

func TestWinHTTP_PreCancelledContext(t *testing.T) {
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	_, err := wt.RoundTrip(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWinHTTP_CancelInFlightHeaders(t *testing.T) {
	releaseHandler := make(chan struct{})
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-releaseHandler:
		case <-time.After(10 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	})
	defer close(releaseHandler)

	_, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(120 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	_, err := wt.RoundTrip(req)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("async cancellation took too long (%v), expected < 3s", elapsed)
	}
}

func TestWinHTTP_CancelInFlightBodyRead(t *testing.T) {
	// Raw TCP HTTP server: sends 200 OK headers + Content-Length: 1000, sends no body
	// (or partial body) and stalls until stopConn closes. Cancelling ctx during Read()
	// must abort promptly with context.Canceled.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	stopConn := make(chan struct{})
	defer close(stopConn)

	go func() {
		conn, aerr := ln.Accept()
		if aerr != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		_, _ = conn.Read(buf)
		_, _ = io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n")
		<-stopConn
	}()

	_, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+ln.Addr().String()+"/stream", nil)

	resp, err := wt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	go func() {
		time.Sleep(120 * time.Millisecond)
		cancel()
	}()

	buf := make([]byte, 64)
	start := time.Now()
	n, err := resp.Body.Read(buf)
	elapsed := time.Since(start)
	t.Logf("Read during stalled body returned n=%d err=%v in %v", n, err, elapsed)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled during Read, got %v", err)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("body read cancellation took too long (%v)", elapsed)
	}
}

func TestWinHTTP_DeadlineExceeded(t *testing.T) {
	release := make(chan struct{})
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	})
	defer close(release)

	_, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	_, err := wt.RoundTrip(req)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestWinHTTP_ConcurrentRequestsAndIdleCycle(t *testing.T) {
	const goroutines = 12
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			b, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "concurrent-ok")
	})

	client, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})

	var wg sync.WaitGroup
	errs := make(chan error, goroutines*2)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == goroutines/2 {
				wt.CloseIdleConnections()
			}
			resp, err := client.Get(srv.URL)
			if err != nil {
				errs <- err
				return
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				errs <- err
				return
			}
			if string(body) != "concurrent-ok" {
				errs <- &unexpectedBodyError{got: string(body)}
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent request error: %v", err)
	}
}

type unexpectedBodyError struct{ got string }

func (e *unexpectedBodyError) Error() string { return "unexpected body: " + e.got }

func TestWinHTTP_CertValidationEnforced(t *testing.T) {
	// Untrusted self-signed TLS server: WinHTTP MUST reject it with security error.
	tlsSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(tlsSrv.Close)

	client, _ := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})
	resp, err := client.Get(tlsSrv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected TLS certificate validation failure against untrusted self-signed cert")
	}
	t.Logf("untrusted cert rejected as expected: %v", err)
}

func TestWinHTTP_CloseIdleIdempotentAndReusable(t *testing.T) {
	srv := setupMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "reused-ok")
	})

	client, wt := newWinHTTPTestClient(t, ControlPlaneTransportOptions{})
	wt.CloseIdleConnections()
	wt.CloseIdleConnections()

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("GET after CloseIdleConnections failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "reused-ok" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}
