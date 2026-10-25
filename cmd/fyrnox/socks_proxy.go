package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type SocksProxyServer struct {
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

func NewSocksProxyServer(listenAddr string, opener tunnelOpener, meter *TrafficMeter) *SocksProxyServer {
	return &SocksProxyServer{
		listenAddr: listenAddr,
		opener:     opener,
		meter:      meter,
		conns:      make(map[net.Conn]struct{}),
	}
}

func (s *SocksProxyServer) SetBypassRules(rules string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bypassRules = rules
	s.settings.BypassRules = rules
}

func (s *SocksProxyServer) SetSettings(settings VpnSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = settings
	s.bypassRules = settings.BypassRules
}

func (s *SocksProxyServer) dialUpstream(target string, clientPort int) (net.Conn, error) {
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
		LogInfo("[SOCKS5 Bypass] Direct route domestic Iran target: %s", target)
		return net.DialTimeout("tcp", target, 5*time.Second)
	}

	// 3. Process-Based Per-App Split Tunneling
	if settings.SplitTunnelEnabled && !IsAppAllowedBySplitTunnel(clientPort, target, false, settings) {
		LogInfo("[SOCKS5 SplitTunnel] Direct bypass for socket target: %s (clientPort: %d)", target, clientPort)
		return net.DialTimeout("tcp", target, 5*time.Second)
	}

	return s.opener.OpenTunnel(target)
}

func (s *SocksProxyServer) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		LogWarn("SOCKS5 port %s was unavailable (%v); finding open port...", s.listenAddr, err)
		altPorts := []int{20808, 20888, 30808, 40808, 50808, 1088, 0}
		for _, p := range altPorts {
			candidate := fmt.Sprintf("127.0.0.1:%d", p)
			altLn, altErr := net.Listen("tcp", candidate)
			if altErr == nil {
				ln = altLn
				s.listenAddr = altLn.Addr().String()
				LogInfo("Automatically bound SOCKS5 to free port: %s", s.listenAddr)
				err = nil
				break
			}
		}
		if err != nil {
			return fmt.Errorf("failed to bind SOCKS5 listener: %w", err)
		}
	}
	s.listener = ln
	LogInfo("SOCKS5 Proxy listening on %s", s.listenAddr)

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

func (s *SocksProxyServer) Close() error {
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

func (s *SocksProxyServer) handleClient(clientConn net.Conn) {
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

	// 1. Negotiation
	buf := make([]byte, 258)
	if _, err := io.ReadFull(clientConn, buf[:2]); err != nil {
		return
	}
	if buf[0] != 0x05 {
		return
	}
	nMethods := int(buf[1])
	if _, err := io.ReadFull(clientConn, buf[:nMethods]); err != nil {
		return
	}
	// Select NO AUTH (0x00)
	if _, err := clientConn.Write([]byte{0x05, 0x00}); err != nil {
		return
	}

	// 2. Request Details
	header := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, header); err != nil {
		return
	}
	if header[0] != 0x05 || header[1] != 0x01 { // CMD != CONNECT
		clientConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var host string
	switch header[3] {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(clientConn, ip); err != nil {
			return
		}
		host = net.IP(ip).String()
	case 0x03: // Domain
		var l [1]byte
		if _, err := io.ReadFull(clientConn, l[:]); err != nil {
			return
		}
		domain := make([]byte, int(l[0]))
		if _, err := io.ReadFull(clientConn, domain); err != nil {
			return
		}
		host = string(domain)
	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(clientConn, ip); err != nil {
			return
		}
		host = net.IP(ip).String()
	default:
		clientConn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, portBuf); err != nil {
		return
	}
	port := binary.BigEndian.Uint16(portBuf)
	target := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// 3. Dial Upstream through VPN tunnel (or directly if local/bypassed)
	upstreamConn, err := s.dialUpstream(target, clientPort)
	if err != nil {
		clientConn.Write([]byte{0x05, 0x04, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer upstreamConn.Close()

	// 4. Success reply
	if _, err := clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}

	// 5. Pipe
	s.pipe(clientConn, upstreamConn)
}

var proxyBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 64*1024)
		return &b
	},
}

func (s *SocksProxyServer) pipe(client, upstream net.Conn) {
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
		bufPtr := proxyBufferPool.Get().(*[]byte)
		buf := *bufPtr
		defer proxyBufferPool.Put(bufPtr)

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
		bufPtr := proxyBufferPool.Get().(*[]byte)
		buf := *bufPtr
		defer proxyBufferPool.Put(bufPtr)

		for {
			nr, err := client.Read(buf)
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
