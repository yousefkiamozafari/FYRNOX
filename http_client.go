package vpnclient

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"firefox-vpn-client/firefoxtls"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

const (
	controlPlaneHTTPTimeout = 75 * time.Second
	errorBodyLogLimit       = 16 * 1024
)

// ControlPlaneTransportOptions configures the unified control-plane transport.
type ControlPlaneTransportOptions struct {
	ProxyURL         *url.URL
	ProxyAuth        *proxy.Auth
	CustomDialer     func(ctx context.Context, network, addr string) (net.Conn, error)
	TLSFragment      bool
	HandshakeTimeout time.Duration
	SessionCache     utls.ClientSessionCache
	Roots            *x509.CertPool
	ALPN             []string
	ForceStandardTLS bool // For comparative A/B diagnostic testing
}

var defaultControlPlaneSessionCache = utls.NewLRUClientSessionCache(64)

// DefaultControlPlaneTransport returns the control-plane transport.
// On Windows with direct or HTTP/HTTPS proxy settings it returns a
// winHTTPTransport; otherwise it returns an *http.Transport with uTLS.
func DefaultControlPlaneTransport() http.RoundTripper {
	tr, err := NewControlPlaneTransport(ControlPlaneTransportOptions{
		SessionCache: defaultControlPlaneSessionCache,
	})
	if err != nil {
		panic(fmt.Sprintf("creating default control plane transport: %v", err))
	}
	return tr
}

// controlPlaneHTTPClient is the isolated HTTP client for all API requests
// (FxA, Guardian, Remote Settings). It carries an explicit timeout so
// requests can never hang indefinitely; SetControlPlaneTransport may replace
// its Transport. It is deliberately kept separate from http.DefaultClient.
// The cookie jar holds Fastly anti-bot challenge cookies earned by
// solveFastlyChallenge.
var controlPlaneHTTPClient = func() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(fmt.Sprintf("creating control plane cookie jar: %v", err))
	}
	return &http.Client{
		Timeout:   controlPlaneHTTPTimeout,
		Jar:       jar,
		Transport: DefaultControlPlaneTransport(), // http.RoundTripper (winHTTPTransport or *http.Transport)
	}
}()

// NewControlPlaneTransport builds the unified control-plane transport.
// On Windows, when WinHTTP is eligible (direct or HTTP/HTTPS proxy, no
// custom dialer, no TLS fragmentation), it returns a winHTTPTransport backed
// by the Windows Schannel TLS stack, which bypasses the Fastly WAF HTTP 406.
// In all other cases it returns an *http.Transport with an authentic Firefox
// uTLS ClientHello.
func NewControlPlaneTransport(opts ControlPlaneTransportOptions) (http.RoundTripper, error) {
	// ── Windows WinHTTP fast-path ────────────────────────────────────────
	if shouldUseWinHTTP(opts) {
		return newWinHTTPTransport(opts)
	}
	// ── Go uTLS path (all proxy/dialer/frag modes) ───────────────────────
	return newGoTLSTransport(opts)
}

// newGoTLSTransport is the original transport factory, now used only when
// WinHTTP is not eligible.
func newGoTLSTransport(opts ControlPlaneTransportOptions) (*http.Transport, error) {
	if opts.HandshakeTimeout <= 0 {
		opts.HandshakeTimeout = 15 * time.Second
	}
	if len(opts.ALPN) == 0 {
		opts.ALPN = []string{"http/1.1"}
	}
	if opts.SessionCache == nil {
		opts.SessionCache = defaultControlPlaneSessionCache
	}

	dialTLSContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}

		// 1. Establish raw connection based on configured proxy or dialer mode
		var raw net.Conn
		if opts.ProxyURL != nil {
			scheme := strings.ToLower(opts.ProxyURL.Scheme)
			switch scheme {
			case "socks5", "socks5h":
				d, err := proxy.SOCKS5("tcp", opts.ProxyURL.Host, opts.ProxyAuth, &net.Dialer{Timeout: 20 * time.Second})
				if err != nil {
					return nil, fmt.Errorf("creating socks5 dialer: %w", err)
				}
				contextDialer, ok := d.(proxy.ContextDialer)
				if !ok {
					return nil, errors.New("socks5 dialer does not implement ContextDialer")
				}
				raw, err = contextDialer.DialContext(ctx, network, address)
				if err != nil {
					return nil, fmt.Errorf("socks5 dial to %s: %w", address, err)
				}
			case "http", "https":
				dialer := &net.Dialer{Timeout: 20 * time.Second}
				proxyConn, err := dialer.DialContext(ctx, "tcp", opts.ProxyURL.Host)
				if err != nil {
					return nil, fmt.Errorf("dialing http proxy %s: %w", opts.ProxyURL.Host, err)
				}
				if err := httpConnectHandshake(ctx, proxyConn, address, opts.ProxyURL, opts.ProxyAuth); err != nil {
					proxyConn.Close()
					return nil, err
				}
				raw = proxyConn
			default:
				return nil, fmt.Errorf("unsupported control plane proxy scheme: %s", scheme)
			}
		} else if opts.CustomDialer != nil {
			raw, err = opts.CustomDialer(ctx, network, address)
			if err != nil {
				return nil, err
			}
		} else {
			dialer := &net.Dialer{Timeout: 20 * time.Second}
			raw, err = dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
		}

		// 2. Wrap with TLS ClientHello fragmentation if requested
		if opts.TLSFragment && strings.HasPrefix(network, "tcp") {
			raw = newFragmentingConn(raw, "sni")
		}

		// 3. Complete TLS handshake: either standard crypto/tls (for A/B testing) or uTLS Firefox ClientHello
		if opts.ForceStandardTLS {
			tlsConfig := &tls.Config{
				ServerName:         host,
				RootCAs:            opts.Roots,
				NextProtos:         opts.ALPN,
				InsecureSkipVerify: false,
			}
			tlsConn := tls.Client(raw, tlsConfig)
			shakeCtx, cancel := context.WithTimeout(ctx, opts.HandshakeTimeout)
			defer cancel()
			if err := tlsConn.HandshakeContext(shakeCtx); err != nil {
				raw.Close()
				return nil, fmt.Errorf("standard crypto/tls handshake to %s failed: %w", host, err)
			}
			return tlsConn, nil
		}

		shakeCtx, cancel := context.WithTimeout(ctx, opts.HandshakeTimeout)
		defer cancel()
		uConn, err := firefoxtls.Handshake(shakeCtx, raw, host, opts.Roots, opts.SessionCache, opts.ALPN...)
		if err != nil {
			raw.Close()
			return nil, fmt.Errorf("firefox uTLS handshake to %s failed: %w", host, err)
		}
		return uConn, nil
	}

	tr := &http.Transport{
		Proxy:               nil, // Ignore ambient environment proxies; proxies are handled directly in DialTLSContext
		DialTLSContext:      dialTLSContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: opts.HandshakeTimeout,
	}

	return tr, nil
}

// httpConnectHandshake performs HTTP CONNECT tunnel establishment to an HTTP proxy.
func httpConnectHandshake(ctx context.Context, conn net.Conn, targetAddr string, proxyURL *url.URL, auth *proxy.Auth) error {
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", targetAddr, targetAddr)
	if auth != nil && auth.User != "" {
		cred := base64.StdEncoding.EncodeToString([]byte(auth.User + ":" + auth.Password))
		req += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", cred)
	} else if proxyURL != nil && proxyURL.User != nil {
		cred := base64.StdEncoding.EncodeToString([]byte(proxyURL.User.String()))
		req += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", cred)
	}
	req += "User-Agent: " + FirefoxUserAgent() + "\r\n\r\n"

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}

	if _, err := io.WriteString(conn, req); err != nil {
		return fmt.Errorf("writing CONNECT to proxy: %w", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		return fmt.Errorf("reading CONNECT response from proxy: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy CONNECT rejected with HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// SetControlPlaneTransport routes Firefox Accounts, Guardian, and Remote
// Settings requests through transport. It must be called before any control
// plane request is issued, because the client is shared process-wide.
func SetControlPlaneTransport(transport http.RoundTripper) {
	if transport == nil {
		controlPlaneHTTPClient.Transport = DefaultControlPlaneTransport()
		return
	}
	if tr, ok := transport.(*http.Transport); ok {
		func() {
			defer func() { _ = recover() }()
			_ = http2.ConfigureTransport(tr)
		}()
	}
	controlPlaneHTTPClient.Transport = transport
}

// ControlPlaneTransport returns the currently active control-plane RoundTripper.
func ControlPlaneTransport() http.RoundTripper {
	return controlPlaneHTTPClient.Transport
}

func doControlPlane(req *http.Request) (*http.Response, error) {
	return controlPlaneHTTPClient.Do(req)
}

func controlPlaneDo(ctx context.Context, newRequest func() (*http.Request, error)) (*http.Response, error) {
	ctx, cancel := contextWithMinDeadline(ctx, 60*time.Second)
	defer cancel()

	for attempt := 1; attempt <= maxChallengeAttempts; attempt++ {
		req, err := newRequest()
		if err != nil {
			return nil, err
		}
		resp, err := doControlPlane(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusNotAcceptable {
			return resp, nil
		}
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()

		if err := solveFastlyChallenge(ctx); err != nil {
			return nil, fmt.Errorf("Fastly anti-bot challenge: %w", err)
		}
	}
	return nil, fmt.Errorf("Fastly anti-bot challenge still failing after %d attempts (HTTP 406)", maxChallengeAttempts)
}

func readErrorBody(r io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(r, errorBodyLogLimit))
	if err != nil {
		return "reading error body failed: " + err.Error()
	}
	return string(body)
}
