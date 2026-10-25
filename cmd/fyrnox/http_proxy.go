package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HttpProxyServer struct {
	listenAddr  string
	listener    net.Listener
	opener      tunnelOpener
	meter       *TrafficMeter
	bypassRules string
	settings    VpnSettings
	closed      bool
	conns       map[net.Conn]struct{}
	mu          sync.Mutex
}

func NewHttpProxyServer(listenAddr string, opener tunnelOpener, meter *TrafficMeter) *HttpProxyServer {
	return &HttpProxyServer{
		listenAddr: listenAddr,
		opener:     opener,
		meter:      meter,
		conns:      make(map[net.Conn]struct{}),
	}
}

func (s *HttpProxyServer) SetBypassRules(rules string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bypassRules = rules
	s.settings.BypassRules = rules
}

func (s *HttpProxyServer) SetSettings(settings VpnSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
	s.bypassRules = settings.BypassRules
}

func (s *HttpProxyServer) dialUpstream(target string, clientPort int) (net.Conn, error) {
	s.mu.Lock()
	bypass := s.bypassRules
	settings := s.settings
	s.mu.Unlock()

	// 1. Windows default bypass rules (<local>, loopback, private ranges, custom domains)
	if isBypassTarget(target, bypass) {
		return net.DialTimeout("tcp", target, 5*time.Second)
	}

	// 2. Smart Iran / Domestic Traffic Direct Routing (Zero-Config Bypass)
	if settings.DirectIranTraffic && IsDomesticIranTarget(target) {
		LogInfo("[HTTP Bypass] Direct route domestic Iran target: %s", target)
		return net.DialTimeout("tcp", target, 5*time.Second)
	}

	// 3. Process-Based Per-App Split Tunneling
	if settings.SplitTunnelEnabled && !IsAppAllowedBySplitTunnel(clientPort, target, false, settings) {
		LogInfo("[HTTP SplitTunnel] Direct bypass for socket target: %s (clientPort: %d)", target, clientPort)
		return net.DialTimeout("tcp", target, 5*time.Second)
	}

	return s.opener.OpenTunnel(target)
}

// isBypassTarget checks whether host/port should bypass the VPN tunnel and connect directly
func isBypassTarget(target, customBypass string) bool {
	host := target
	if h, _, err := net.SplitHostPort(target); err == nil {
		host = h
	}
	hostLower := strings.ToLower(strings.TrimSpace(host))

	// Loopback and local hostnames
	if hostLower == "localhost" || hostLower == "127.0.0.1" || hostLower == "::1" {
		return true
	}

	// Private IP addresses
	if ip := net.ParseIP(hostLower); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return true
		}
	}

	// Plain intranet hosts without dots or standard local network TLDs (.local, .lan, .internal, .home)
	if !strings.Contains(hostLower, ".") || strings.HasSuffix(hostLower, ".local") || strings.HasSuffix(hostLower, ".lan") || strings.HasSuffix(hostLower, ".internal") || strings.HasSuffix(hostLower, ".home") {
		return true
	}

	if customBypass != "" {
		rules := strings.FieldsFunc(customBypass, func(r rune) bool {
			return r == ';' || r == ',' || r == '\n' || r == '\r'
		})
		for _, rule := range rules {
			pattern := strings.ToLower(strings.TrimSpace(rule))
			if pattern == "" {
				continue
			}
			if pattern == "<local>" && !strings.Contains(hostLower, ".") {
				return true
			}
			if pattern == hostLower {
				return true
			}
			if strings.HasPrefix(pattern, "*.") && strings.HasSuffix(hostLower, pattern[1:]) {
				return true
			}
			if strings.HasPrefix(pattern, ".") && strings.HasSuffix(hostLower, pattern) {
				return true
			}
			if strings.HasSuffix(pattern, "*") && strings.HasPrefix(hostLower, pattern[:len(pattern)-1]) {
				return true
			}
		}
	}
	return false
}

func (s *HttpProxyServer) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		LogWarn("HTTP proxy port %s was unavailable (%v); finding open port...", s.listenAddr, err)
		altPorts := []int{20809, 20889, 30809, 40809, 50809, 1089, 0}
		for _, p := range altPorts {
			candidate := fmt.Sprintf("127.0.0.1:%d", p)
			altLn, altErr := net.Listen("tcp", candidate)
			if altErr == nil {
				ln = altLn
				s.listenAddr = altLn.Addr().String()
				LogInfo("Automatically bound HTTP proxy to free port: %s", s.listenAddr)
				err = nil
				break
			}
		}
		if err != nil {
			return fmt.Errorf("failed to bind HTTP listener: %w", err)
		}
	}
	s.listener = ln
	LogInfo("HTTP Proxy listening on %s", s.listenAddr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				s.mu.Lock()
				closed := s.closed
				s.mu.Unlock()
				if closed {
					return
				}
				time.Sleep(10 * time.Millisecond)
				continue
			}
			s.mu.Lock()
			if s.closed {
				_ = conn.Close()
				s.mu.Unlock()
				return
			}
			s.conns[conn] = struct{}{}
			s.mu.Unlock()

			go s.handleClient(conn)
		}
	}()

	return nil
}

func (s *HttpProxyServer) Close() error {
	s.mu.Lock()
	s.closed = true
	for c := range s.conns {
		_ = c.Close()
	}
	s.conns = make(map[net.Conn]struct{})
	ln := s.listener
	s.listener = nil
	s.mu.Unlock()
	if ln != nil {
		return ln.Close()
	}
	return nil
}

func (s *HttpProxyServer) handleClient(clientConn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.conns, clientConn)
		s.mu.Unlock()
		_ = clientConn.Close()
	}()

	clientPort := 0
	if tcpAddr, ok := clientConn.RemoteAddr().(*net.TCPAddr); ok {
		clientPort = tcpAddr.Port
	}

	reader := bufio.NewReader(clientConn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	target := req.Host
	if req.Method == http.MethodConnect {
		// HTTPS CONNECT tunnel - remote DNS resolution occurs at upstream Fastly edge
		if !strings.Contains(target, ":") {
			target += ":443"
		}

		upstreamConn, err := s.dialUpstream(target, clientPort)
		if err != nil {
			clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
			return
		}
		defer upstreamConn.Close()

		_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			return
		}

		s.pipe(reader, clientConn, upstreamConn)
	} else {
		// Plain HTTP proxy
		if !strings.Contains(target, ":") {
			target += ":80"
		}

		upstreamConn, err := s.dialUpstream(target, clientPort)
		if err != nil {
			clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
			return
		}
		defer upstreamConn.Close()

		// Strip hop-by-hop proxy headers
		req.Header.Del("Proxy-Connection")
		req.Header.Del("Proxy-Authenticate")
		req.Header.Del("Proxy-Authorization")

		// Write request to upstream
		err = req.Write(upstreamConn)
		if err != nil {
			return
		}

		s.pipe(reader, clientConn, upstreamConn)
	}
}

var httpBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 64*1024)
		return &b
	},
}

func (s *HttpProxyServer) pipe(clientReader io.Reader, client, upstream net.Conn) {
	if tc, ok := client.(*net.TCPConn); ok {
		_ = tc.SetNoDelay(true)
		_ = tc.SetReadBuffer(1024 * 1024)
		_ = tc.SetWriteBuffer(1024 * 1024)
	}
	if tc, ok := upstream.(*net.TCPConn); ok {
		_ = tc.SetNoDelay(true)
		_ = tc.SetReadBuffer(1024 * 1024)
		_ = tc.SetWriteBuffer(1024 * 1024)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Upstream -> Client (Download)
	go func() {
		defer wg.Done()
		bufPtr := httpBufferPool.Get().(*[]byte)
		buf := *bufPtr
		defer httpBufferPool.Put(bufPtr)

		for {
			nr, err := upstream.Read(buf)
			if nr > 0 {
				s.meter.AddDown(int64(nr))
				_, werr := client.Write(buf[:nr])
				if werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		_ = client.Close()
	}()

	// Client -> Upstream (Upload)
	go func() {
		defer wg.Done()
		bufPtr := httpBufferPool.Get().(*[]byte)
		buf := *bufPtr
		defer httpBufferPool.Put(bufPtr)

		for {
			nr, err := clientReader.Read(buf)
			if nr > 0 {
				s.meter.AddUp(int64(nr))
				_, werr := upstream.Write(buf[:nr])
				if werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		_ = upstream.Close()
	}()

	wg.Wait()
}
