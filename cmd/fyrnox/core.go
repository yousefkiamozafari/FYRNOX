package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	vpnclient "firefox-vpn-client"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

const (
	defaultTunnelWriteBufSize    = 2 << 20 // 2 MB buffer for maximum download throughput
	upstreamKeepAliveInterval    = 10 * time.Second
	upstreamKeepAlivePingTimeout = 5 * time.Second
	proxyPassRenewLead           = 2 * time.Minute
	proxyPassRetryDelay          = 30 * time.Second
	proxyPassRenewTimeout        = 2 * time.Minute
	oauthRefreshLead             = 2 * time.Minute
	maxOpenTunnelRebuildRetries  = 3
	defaultExitCheckTimeout      = 10 * time.Second
	defaultExitCheckURL          = "https://www.cloudflare.com/cdn-cgi/trace"
	apiRequestTimeout            = 60 * time.Second
)

var (
	errProxyHTTP2Unavailable = errors.New("proxy did not negotiate HTTP/2")
	errProxyHTTP3Unavailable = errors.New("proxy did not negotiate HTTP/3")
	errProxySessionUnhealthy = errors.New("upstream proxy session became unhealthy")
	errNoUsableProxySession  = errors.New("no usable upstream proxy session")
	errTunnelDeadline        = errors.New("deadlines are not supported for HTTP CONNECT tunnel streams")
)

type tunnelOpener interface {
	OpenTunnel(authority string) (net.Conn, error)
}

type proxySession interface {
	tunnelOpener
	Close() error
}

type proxySessionTokenUpdater interface {
	UpdateToken(token string) error
}

type proxySessionLivenessChecker interface {
	IsAlive() bool
}

func isSessionAlive(session proxySession) bool {
	if session == nil {
		return false
	}
	if checker, ok := session.(proxySessionLivenessChecker); ok {
		return checker.IsAlive()
	}
	return true
}

type runtimeAuth struct {
	Token      *vpnclient.TokenResponse
	ObtainedAt time.Time
}

func (a *runtimeAuth) accessTokenValid(now time.Time) bool {
	if a == nil || a.Token == nil || a.Token.AccessToken == "" {
		return false
	}
	expiry := a.ObtainedAt.Add(time.Duration(a.Token.ExpiresIn) * time.Second)
	return now.Before(expiry.Add(-oauthRefreshLead))
}

// --- tunnelWriteBuffer & tunnelConn ---

type tunnelWriteBuffer struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	buf      []byte
	n        int
	off      int
	closed   bool
	readErr  error
}

func newTunnelWriteBuffer(size int) *tunnelWriteBuffer {
	if size < 1 {
		size = defaultTunnelWriteBufSize
	}
	b := &tunnelWriteBuffer{buf: make([]byte, size)}
	b.notEmpty = sync.NewCond(&b.mu)
	b.notFull = sync.NewCond(&b.mu)
	return b
}

func (b *tunnelWriteBuffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.n == 0 && b.readErr == nil && !b.closed {
		b.notEmpty.Wait()
	}
	if b.n == 0 {
		if b.readErr != nil {
			return 0, b.readErr
		}
		return 0, io.EOF
	}

	end := b.off + b.n
	if end > len(b.buf) {
		end = len(b.buf)
	}
	n := copy(p, b.buf[b.off:end])
	b.off = (b.off + n) % len(b.buf)
	b.n -= n
	b.notFull.Broadcast()
	return n, nil
}

func (b *tunnelWriteBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	total := 0
	for len(p) > 0 {
		for b.n == len(b.buf) && b.readErr == nil && !b.closed {
			b.notFull.Wait()
		}
		if b.readErr != nil {
			return total, b.readErr
		}
		if b.closed {
			return total, io.ErrClosedPipe
		}

		space := len(b.buf) - b.n
		head := (b.off + b.n) % len(b.buf)
		writable := len(b.buf) - head
		if space < writable {
			writable = space
		}
		if writable > len(p) {
			writable = len(p)
		}
		copy(b.buf[head:], p[:writable])
		b.n += writable
		p = p[writable:]
		total += writable
		b.notEmpty.Broadcast()
	}
	return total, nil
}

func (b *tunnelWriteBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	b.notEmpty.Broadcast()
	b.notFull.Broadcast()
	return nil
}

func (b *tunnelWriteBuffer) failRead(err error) {
	b.mu.Lock()
	if b.readErr == nil {
		b.readErr = err
	}
	b.closed = true
	b.notEmpty.Broadcast()
	b.notFull.Broadcast()
	b.mu.Unlock()
}

type tunnelAddr string

func (a tunnelAddr) Network() string { return "tcp" }
func (a tunnelAddr) String() string  { return string(a) }

type tunnelConn struct {
	reader    io.ReadCloser
	writer    *tunnelWriteBuffer
	reqBody   *tunnelWriteBuffer
	cancel    context.CancelFunc
	name      tunnelAddr
	closeOnce sync.Once
	closeErr  error
	readErr   error
	writeErr  error
}

func (c *tunnelConn) Read(p []byte) (int, error)  { return c.reader.Read(p) }
func (c *tunnelConn) Write(p []byte) (int, error) { return c.writer.Write(p) }
func (c *tunnelConn) Close() error {
	c.closeOnce.Do(func() {
		wErr := c.writer.Close()
		rErr := c.reader.Close()
		if c.reqBody != nil {
			c.reqBody.failRead(io.ErrClosedPipe)
		}
		if c.cancel != nil {
			c.cancel()
		}
		if wErr != nil {
			c.closeErr = wErr
		} else {
			c.closeErr = rErr
		}
	})
	return c.closeErr
}
func (c *tunnelConn) LocalAddr() net.Addr              { return c.name }
func (c *tunnelConn) RemoteAddr() net.Addr             { return c.name }
func (c *tunnelConn) SetDeadline(time.Time) error      { return errTunnelDeadline }
func (c *tunnelConn) SetReadDeadline(time.Time) error  { return errTunnelDeadline }
func (c *tunnelConn) SetWriteDeadline(time.Time) error { return errTunnelDeadline }

// --- HTTP/2 Proxy Session ---

type h2ProxySession struct {
	raw           *tls.Conn
	cc            *http2.ClientConn
	proxyHost     string
	token         string
	timeout       time.Duration
	tokenMu       sync.RWMutex
	alive         atomic.Bool
	stopKeepAlive context.CancelFunc
	closeOnce     sync.Once
	closeErr      error
}

func createProxyTLSConfig(serverHostname string, nextProtos []string) *tls.Config {
	return &tls.Config{
		ServerName: serverHostname,
		InsecureSkipVerify: true, // Verification is performed securely in VerifyPeerCertificate
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("tls: no certificates presented by proxy server")
			}

			certs := make([]*x509.Certificate, len(rawCerts))
			for i, asn1Data := range rawCerts {
				cert, err := x509.ParseCertificate(asn1Data)
				if err != nil {
					return fmt.Errorf("tls: failed to parse peer certificate: %w", err)
				}
				certs[i] = cert
			}

			leaf := certs[0]
			opts := x509.VerifyOptions{
				Intermediates: x509.NewCertPool(),
				CurrentTime:   time.Now(),
			}
			for _, cert := range certs[1:] {
				opts.Intermediates.AddCert(cert)
			}

			// 1. Direct standard hostname verification
			opts.DNSName = serverHostname
			if _, err := leaf.Verify(opts); err == nil {
				return nil
			}

			// 2. Fastly MASQUE edge servers present Fastly's edge certificates (default.ssl.fastly.net, *.hosts.fastly.net, *.fastly.com).
			// If connecting to a Fastly MASQUE node (*.fastly-masque.net or *.fastly.net), verify against Fastly's SANs using system roots.
			isFastly := strings.Contains(serverHostname, "fastly-masque.net") || strings.Contains(serverHostname, "fastly.net")
			if isFastly {
				for _, san := range leaf.DNSNames {
					if strings.HasSuffix(san, ".fastly.net") || strings.HasSuffix(san, ".fastly.com") || san == "default.ssl.fastly.net" {
						opts.DNSName = san
						if _, err := leaf.Verify(opts); err == nil {
							return nil
						}
					}
				}
				for _, defaultSAN := range []string{"default.ssl.fastly.net", "hosts.fastly.net", "fastly.com"} {
					opts.DNSName = defaultSAN
					if _, err := leaf.Verify(opts); err == nil {
						return nil
					}
				}
			}

			return fmt.Errorf("tls: certificate verification failed for %s", serverHostname)
		},
		MinVersion: tls.VersionTLS12,
		NextProtos: nextProtos,
	}
}

func createProxyDialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout: timeout,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 4 * time.Second}
				dnsServers := []string{
					"1.1.1.1:53",
					"8.8.8.8:53",
					"9.9.9.9:53",
					"1.0.0.1:53",
					"8.8.4.4:53",
				}
				for _, dns := range dnsServers {
					conn, err := d.DialContext(ctx, "udp", dns)
					if err == nil {
						return conn, nil
					}
					conn, err = d.DialContext(ctx, "tcp", dns)
					if err == nil {
						return conn, nil
					}
				}
				return d.DialContext(ctx, network, address)
			},
		},
	}
}

// fragmentingConn wraps a net.Conn to fragment the TLS ClientHello across
// multiple TCP segments, bypassing SNI-based Deep Packet Inspection (DPI)
// censorship (matching Vauth/Fyrnox v1.0.7).
type fragmentingConn struct {
	net.Conn
	mu             sync.Mutex
	firstWriteDone bool
	mode           string
}

func newFragmentingConn(c net.Conn, mode string) *fragmentingConn {
	if mode == "" {
		mode = "sni"
	}
	return &fragmentingConn{
		Conn: c,
		mode: mode,
	}
}

func (fc *fragmentingConn) Write(b []byte) (int, error) {
	fc.mu.Lock()
	if fc.firstWriteDone {
		fc.mu.Unlock()
		return fc.Conn.Write(b)
	}
	fc.firstWriteDone = true
	fc.mu.Unlock()

	// Check if this write looks like a TLS ClientHello
	// TLS record header: 0x16 (Handshake), Version (0x03, 0x01..0x03)
	// Handshake type at offset 5: 0x01 (ClientHello)
	if len(b) > 9 && b[0] == 0x16 && b[1] == 0x03 && b[5] == 0x01 {
		switch fc.mode {
		case "first_byte":
			n1, err := fc.Conn.Write(b[:1])
			if err != nil {
				return n1, err
			}
			time.Sleep(8 * time.Millisecond)
			n2, err := fc.Conn.Write(b[1:])
			return n1 + n2, err

		case "sni", "chunks":
			fallthrough
		default:
			// Split into 3 segments to break TLS header, version, and SNI extension:
			// Segment 1: Record header (5 bytes)
			// Segment 2: Handshake header + prefix (30 bytes)
			// Segment 3: The remainder (contains cipher suites and SNI extension)
			split1 := 5
			split2 := 35
			if len(b) > split2 {
				n1, err := fc.Conn.Write(b[:split1])
				if err != nil {
					return n1, err
				}
				time.Sleep(6 * time.Millisecond)

				n2, err := fc.Conn.Write(b[split1:split2])
				if err != nil {
					return n1 + n2, err
				}
				time.Sleep(6 * time.Millisecond)

				n3, err := fc.Conn.Write(b[split2:])
				return n1 + n2 + n3, err
			} else if len(b) > split1 {
				n1, err := fc.Conn.Write(b[:split1])
				if err != nil {
					return n1, err
				}
				time.Sleep(6 * time.Millisecond)

				n2, err := fc.Conn.Write(b[split1:])
				return n1 + n2, err
			}
		}
	}

	return fc.Conn.Write(b)
}

func dialProxyTLS(ctx context.Context, proxyHost string, tlsCfg *tls.Config, timeout time.Duration, apiProxy string, tlsFrag bool, ipProtocol string) (*tls.Conn, error) {
	apiProxy = strings.TrimSpace(apiProxy)
	var rawConn net.Conn
	var err error

	if apiProxy != "" {
		formatted := apiProxy
		if !strings.Contains(formatted, "://") {
			if strings.Contains(formatted, ":1080") || strings.Contains(formatted, ":10808") {
				formatted = "socks5://" + formatted
			} else {
				formatted = "http://" + formatted
			}
		}
		pURL, parseErr := url.Parse(formatted)
		if parseErr == nil {
			switch strings.ToLower(pURL.Scheme) {
			case "socks5", "socks5h":
				var auth *proxy.Auth
				if pURL.User != nil {
					pwd, _ := pURL.User.Password()
					auth = &proxy.Auth{User: pURL.User.Username(), Password: pwd}
				}
				socksDialer, sErr := proxy.SOCKS5("tcp", pURL.Host, auth, &net.Dialer{Timeout: timeout})
				if sErr == nil {
					if cd, ok := socksDialer.(proxy.ContextDialer); ok {
						rawConn, err = cd.DialContext(ctx, "tcp", proxyHost)
					} else {
						rawConn, err = socksDialer.Dial("tcp", proxyHost)
					}
				} else {
					err = sErr
				}
			case "http", "https":
				d := &net.Dialer{Timeout: timeout}
				pConn, pErr := d.DialContext(ctx, "tcp", pURL.Host)
				if pErr == nil {
					req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", proxyHost, proxyHost)
					if pURL.User != nil {
						auth := base64.StdEncoding.EncodeToString([]byte(pURL.User.String()))
						req = fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: Basic %s\r\n\r\n", proxyHost, proxyHost, auth)
					}
					_, _ = pConn.Write([]byte(req))
					buf := make([]byte, 1024)
					n, rErr := pConn.Read(buf)
					if rErr == nil && strings.Contains(string(buf[:n]), "200") {
						rawConn = pConn
					} else {
						_ = pConn.Close()
						err = fmt.Errorf("http proxy connect failed: %s", string(buf[:n]))
					}
				} else {
					err = pErr
				}
			}
		}
	}

	if rawConn == nil && err == nil {
		dialer := createProxyDialer(timeout)
		network := "tcp"
		switch strings.ToLower(ipProtocol) {
		case "ipv4":
			network = "tcp4"
		case "ipv6":
			network = "tcp6"
		}
		rawConn, err = dialer.DialContext(ctx, network, proxyHost)
		if err != nil && network == "tcp6" {
			return nil, fmt.Errorf("IPv6 connection failed: %w (your network or the server may not support IPv6 routing; select Dual Stack or IPv4 Only)", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("dialing %s: %w", proxyHost, err)
	}

	if tc, ok := rawConn.(*net.TCPConn); ok {
		_ = tc.SetNoDelay(true)
		_ = tc.SetKeepAlive(true)
		_ = tc.SetKeepAlivePeriod(15 * time.Second)
		_ = tc.SetReadBuffer(2 * 1024 * 1024)
		_ = tc.SetWriteBuffer(2 * 1024 * 1024)
	}

	if tlsFrag {
		rawConn = newFragmentingConn(rawConn, "sni")
	}

	tlsConn := tls.Client(rawConn, tlsCfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = rawConn.Close()
		return nil, fmt.Errorf("tls handshake with %s failed: %w", proxyHost, err)
	}

	return tlsConn, nil
}

func newH2ProxySession(proxyURL *url.URL, token string, timeout time.Duration, apiProxy string, tlsFrag bool, ipProtocol string) (*h2ProxySession, error) {
	proxyHost := proxyURL.Host
	if proxyURL.Port() == "" {
		proxyHost = proxyURL.Hostname() + ":443"
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tlsCfg := createProxyTLSConfig(proxyURL.Hostname(), []string{"h2"})
	proxyTLS, err := dialProxyTLS(ctx, proxyHost, tlsCfg, timeout, apiProxy, tlsFrag, ipProtocol)
	if err != nil {
		return nil, err
	}

	if proxyTLS.ConnectionState().NegotiatedProtocol != "h2" {
		_ = proxyTLS.Close()
		return nil, errProxyHTTP2Unavailable
	}

	transport := &http2.Transport{
		ReadIdleTimeout:  30 * time.Second,
		PingTimeout:      15 * time.Second,
		WriteByteTimeout: 30 * time.Second,
	}
	transport.CloseIdleConnections() // Initialize underlying transport to prevent nil dereference in x/net/http2 on Go 1.26
	cc, err := transport.NewClientConn(proxyTLS)
	if err != nil {
		_ = proxyTLS.Close()
		return nil, err
	}

	kCtx, kCancel := context.WithCancel(context.Background())
	s := &h2ProxySession{
		raw:           proxyTLS,
		cc:            cc,
		proxyHost:     proxyHost,
		token:         token,
		timeout:       timeout,
		stopKeepAlive: kCancel,
	}
	s.alive.Store(true)
	go s.runKeepAlive(kCtx)
	return s, nil
}

func (s *h2ProxySession) runKeepAlive(ctx context.Context) {
	ticker := time.NewTicker(upstreamKeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, upstreamKeepAlivePingTimeout)
			err := s.cc.Ping(pingCtx)
			cancel()
			if err != nil {
				s.alive.Store(false)
				return
			}
		}
	}
}

func (s *h2ProxySession) IsAlive() bool {
	if !s.alive.Load() {
		return false
	}
	return s.cc.CanTakeNewRequest()
}

func (s *h2ProxySession) OpenTunnel(authority string) (net.Conn, error) {
	reqBody := newTunnelWriteBuffer(defaultTunnelWriteBufSize)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)

	req := &http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Host: authority,
		},
		Host:          authority,
		Header:        make(http.Header),
		Body:          reqBody,
		ContentLength: -1, // CRITICAL: ContentLength must be -1 so Fastly h2o accepts the duplex stream without 400 Bad Request
	}
	req = req.WithContext(ctx)

	s.tokenMu.RLock()
	curToken := s.token
	s.tokenMu.RUnlock()
	req.Header.Set("Proxy-Authorization", "Bearer "+curToken)

	resp, err := s.cc.RoundTrip(req)
	if err != nil {
		cancel()
		reqBody.failRead(io.ErrClosedPipe)
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		cancel()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		_ = resp.Body.Close()
		reqBody.failRead(io.ErrClosedPipe)
		return nil, fmt.Errorf("proxy CONNECT HTTP %d: %s", resp.StatusCode, string(body))
	}

	return &tunnelConn{
		reader:  resp.Body,
		writer:  reqBody,
		reqBody: reqBody,
		name:    tunnelAddr(authority),
		cancel:  cancel,
	}, nil
}

func (s *h2ProxySession) UpdateToken(token string) error {
	s.tokenMu.Lock()
	s.token = token
	s.tokenMu.Unlock()
	return nil
}

func (s *h2ProxySession) Close() error {
	s.closeOnce.Do(func() {
		s.alive.Store(false)
		if s.stopKeepAlive != nil {
			s.stopKeepAlive()
		}
		if s.cc != nil {
			s.closeErr = s.cc.Close()
		}
		if s.raw != nil {
			_ = s.raw.Close()
		}
	})
	return s.closeErr
}

// --- HTTP/3 Proxy Session ---

type h3ProxySession struct {
	conn      *quic.Conn
	udpConn   *net.UDPConn
	rt        *http3.Transport
	proxyHost string
	token     string
	timeout   time.Duration
	tokenMu   sync.RWMutex
	closeOnce sync.Once
	closeErr  error
}

func newH3ProxySession(proxyURL *url.URL, token string, timeout time.Duration, apiProxy string, ipProtocol string) (*h3ProxySession, error) {
	host := proxyURL.Hostname()
	portStr := proxyURL.Port()
	if portStr == "" {
		portStr = "443"
	}
	proxyHost := net.JoinHostPort(host, portStr)

	resolver := createProxyDialer(timeout).Resolver
	ips, err := resolver.LookupIP(context.Background(), "ip", host)
	var udpAddr *net.UDPAddr
	if err == nil && len(ips) > 0 {
		port, _ := strconv.Atoi(portStr)
		udpAddr = &net.UDPAddr{IP: ips[0], Port: port}
	} else {
		udpAddr, err = net.ResolveUDPAddr("udp", proxyHost)
		if err != nil {
			return nil, err
		}
	}

	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tlsCfg := createProxyTLSConfig(host, []string{"h3"})
	tlsCfg.MinVersion = tls.VersionTLS13

	qConn, err := quic.Dial(ctx, udpConn, udpAddr, tlsCfg, &quic.Config{
		KeepAlivePeriod: 15 * time.Second,
	})
	if err != nil {
		_ = udpConn.Close()
		return nil, err
	}

	rt := &http3.Transport{}
	return &h3ProxySession{
		conn:      qConn,
		udpConn:   udpConn,
		rt:        rt,
		proxyHost: proxyHost,
		token:     token,
		timeout:   timeout,
	}, nil
}

func (s *h3ProxySession) IsAlive() bool {
	return s.conn != nil && s.conn.Context().Err() == nil
}

func (s *h3ProxySession) OpenTunnel(authority string) (net.Conn, error) {
	reqBody := newTunnelWriteBuffer(defaultTunnelWriteBufSize)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)

	req, err := http.NewRequestWithContext(ctx, http.MethodConnect, "https://"+s.proxyHost, reqBody)
	if err != nil {
		cancel()
		return nil, err
	}
	req.Host = authority

	s.tokenMu.RLock()
	tok := s.token
	s.tokenMu.RUnlock()
	req.Header.Set("Proxy-Authorization", "Bearer "+tok)

	resp, err := s.rt.RoundTrip(req)
	if err != nil {
		cancel()
		reqBody.failRead(io.ErrClosedPipe)
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		cancel()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		_ = resp.Body.Close()
		reqBody.failRead(io.ErrClosedPipe)
		return nil, fmt.Errorf("h3 CONNECT HTTP %d: %s", resp.StatusCode, string(body))
	}

	return &tunnelConn{
		reader:  resp.Body,
		writer:  reqBody,
		reqBody: reqBody,
		name:    tunnelAddr(authority),
		cancel:  cancel,
	}, nil
}

func (s *h3ProxySession) UpdateToken(token string) error {
	s.tokenMu.Lock()
	s.token = token
	s.tokenMu.Unlock()
	return nil
}

func (s *h3ProxySession) Close() error {
	s.closeOnce.Do(func() {
		if s.conn != nil {
			s.closeErr = (*s.conn).CloseWithError(0, "closed")
		}
		if s.udpConn != nil {
			_ = s.udpConn.Close()
		}
	})
	return s.closeErr
}

// --- Controller & Session Management ---

type proxyController struct {
	guardian       string
	proxyURL       *url.URL
	timeout        time.Duration
	sessionFactory func(token string) (proxySession, error)
	auth           *runtimeAuth
	mu             sync.Mutex
	session        proxySession
	rawToken       string
	expiresAt      time.Time
	closed         bool
	done           chan struct{}
	closeOnce      sync.Once
	fallbackMgr    *FallbackManager
}

func newProxyController(guardian string, proxyURL *url.URL, timeout time.Duration, auth *runtimeAuth, pass *vpnclient.ProxyPassInfo, useH3 bool, apiProxy string, tlsFrag bool, ipProtocol string) (*proxyController, error) {
	var factory func(token string) (proxySession, error)
	if useH3 {
		factory = func(token string) (proxySession, error) {
			return newH3ProxySession(proxyURL, token, timeout, apiProxy, ipProtocol)
		}
	} else {
		factory = func(token string) (proxySession, error) {
			return newH2ProxySession(proxyURL, token, timeout, apiProxy, tlsFrag, ipProtocol)
		}
	}

	sess, err := factory(pass.RawToken)
	if err != nil {
		return nil, err
	}

	c := &proxyController{
		guardian:       guardian,
		proxyURL:       proxyURL,
		timeout:        timeout,
		sessionFactory: factory,
		auth:           auth,
		session:        sess,
		rawToken:       pass.RawToken,
		expiresAt:      pass.ExpiresAt(),
		done:           make(chan struct{}),
	}
	c.fallbackMgr = NewFallbackManager(proxyURL, apiProxy, ipProtocol, useH3, func(switchToH3 bool) error {
		c.mu.Lock()
		defer c.mu.Unlock()
		if switchToH3 {
			c.sessionFactory = func(token string) (proxySession, error) {
				return newH3ProxySession(proxyURL, token, timeout, apiProxy, ipProtocol)
			}
		} else {
			c.sessionFactory = func(token string) (proxySession, error) {
				return newH2ProxySession(proxyURL, token, timeout, apiProxy, true, ipProtocol)
			}
		}
		newSess, err := c.sessionFactory(c.rawToken)
		if err != nil {
			return err
		}
		if c.session != nil {
			_ = c.session.Close()
		}
		c.session = newSess
		return nil
	})
	go c.runRenewalLoop()
	go c.runKeepAliveWatchdog()
	return c, nil
}

func (c *proxyController) runKeepAliveWatchdog() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			closed := c.closed
			sess := c.session
			c.mu.Unlock()

			if closed {
				return
			}

			if sess == nil || !isSessionAlive(sess) {
				LogWarn("Proxy session keepalive failed; proactively rebuilding in-memory session...")
				_ = c.rebuild()
			}
		}
	}
}

func (c *proxyController) OpenTunnel(authority string) (net.Conn, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errNoUsableProxySession
	}
	sess := c.session
	c.mu.Unlock()

	if sess == nil || !isSessionAlive(sess) {
		if err := c.rebuild(); err != nil {
			return nil, err
		}
		c.mu.Lock()
		sess = c.session
		c.mu.Unlock()
	}

	conn, err := sess.OpenTunnel(authority)
	if c.fallbackMgr != nil {
		c.fallbackMgr.RecordUDPResult(err == nil, 0)
	}
	return conn, err
}

func (c *proxyController) GetFallbackStatus() (active bool, state string, lossRate float64) {
	if c.fallbackMgr == nil {
		return false, "tcp", 0
	}
	return c.fallbackMgr.GetStatus()
}

func (c *proxyController) rebuild() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	LogInfo("Rebuilding upstream proxy session...")
	sess, err := c.sessionFactory(c.rawToken)
	if err != nil {
		return err
	}
	if c.session != nil {
		_ = c.session.Close()
	}
	c.session = sess
	return nil
}

func (c *proxyController) runRenewalLoop() {
	for {
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return
		}
		exp := c.expiresAt
		c.mu.Unlock()

		renewAt := exp.Add(-proxyPassRenewLead)
		wait := time.Until(renewAt)
		if wait < 0 {
			wait = 0
		}

		select {
		case <-c.done:
			return
		case <-time.After(wait):
		}

		LogInfo("Renewing Guardian proxy pass token...")
		ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
		pass, err := vpnclient.FetchProxyPass(ctx, c.guardian, c.auth.Token.AccessToken)
		cancel()
		if err != nil {
			LogWarn("Proxy pass renewal error: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		c.mu.Lock()
		c.rawToken = pass.RawToken
		c.expiresAt = pass.ExpiresAt()
		updater, ok := c.session.(proxySessionTokenUpdater)
		c.mu.Unlock()

		if ok {
			_ = updater.UpdateToken(pass.RawToken)
			LogInfo("Proxy pass token updated in-place (expires: %s)", pass.ExpiresAt().Format("15:04:05"))
		} else {
			_ = c.rebuild()
		}
	}
}

func (c *proxyController) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
		if c.fallbackMgr != nil {
			c.fallbackMgr.Stop()
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		c.closed = true
		if c.session != nil {
			_ = c.session.Close()
		}
	})
	return nil
}

// --- Exit Verification ---

type ExitProbeResult struct {
	IP          string `json:"ip"`
	CountryCode string `json:"country_code"`
	LatencyMs   int64  `json:"latency_ms"`
}

func VerifyExit(opener tunnelOpener, timeout time.Duration) (ExitProbeResult, error) {
	var streamLatency int64
	transport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: timeout,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		DialContext: func(ctx context.Context, _, authority string) (net.Conn, error) {
			t0 := time.Now()
			c, err := opener.OpenTunnel(authority)
			if err == nil {
				streamLatency = time.Since(t0).Milliseconds()
			}
			return c, err
		},
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: timeout}

	// 1. Try 1.1.1.1/cdn-cgi/trace directly (no DNS lookup delay over newly established tunnel)
	probeTrace := func(targetURL string) (ExitProbeResult, bool) {
		r, err := client.Get(targetURL)
		if err != nil {
			return ExitProbeResult{}, false
		}
		defer r.Body.Close()
		body, rErr := io.ReadAll(io.LimitReader(r.Body, 16384))
		if rErr != nil {
			return ExitProbeResult{}, false
		}
		ip := ""
		country := ""
		for _, line := range strings.Split(string(body), "\n") {
			k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
			if ok {
				switch strings.ToLower(k) {
				case "ip":
					ip = v
				case "loc":
					country = v
				}
			}
		}
		if ip != "" {
			lat := streamLatency
			if lat <= 0 {
				lat = 1
			}
			return ExitProbeResult{
				IP:          ip,
				CountryCode: country,
				LatencyMs:   lat,
			}, true
		}
		return ExitProbeResult{}, false
	}

	if res, ok := probeTrace("https://1.1.1.1/cdn-cgi/trace"); ok {
		return res, nil
	}
	if res, ok := probeTrace(defaultExitCheckURL); ok {
		return res, nil
	}

	// 2. Fallback to icanhazip.com through tunnel
	resp2, err2 := client.Get("https://icanhazip.com")
	if err2 == nil {
		defer resp2.Body.Close()
		body, rErr := io.ReadAll(io.LimitReader(resp2.Body, 128))
		if rErr == nil {
			ip := strings.TrimSpace(string(body))
			if ip != "" {
				lat := streamLatency
				if lat <= 0 {
					lat = 1
				}
				return ExitProbeResult{
					IP:          ip,
					CountryCode: "",
					LatencyMs:   lat,
				}, nil
			}
		}
	}

	if err2 != nil {
		return ExitProbeResult{}, err2
	}
	return ExitProbeResult{}, fmt.Errorf("exit verification failed to detect IP")
}
