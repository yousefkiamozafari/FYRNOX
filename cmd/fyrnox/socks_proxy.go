package main

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
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
	s.listenAddr = ln.Addr().String()
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

	// 2. Request Details (supports CMD 0x01 CONNECT and CMD 0x03 UDP ASSOCIATE for Wintun TUN)
	header := make([]byte, 4)
	if _, err := io.ReadFull(clientConn, header); err != nil {
		return
	}
	if header[0] != 0x05 || (header[1] != 0x01 && header[1] != 0x03) {
		clientConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	if header[1] == 0x03 {
		s.handleUDPAssociate(clientConn, header)
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
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = upstreamConn.Close()
		return
	}
	s.conns[upstreamConn] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.conns, upstreamConn)
		s.mu.Unlock()
		_ = upstreamConn.Close()
	}()

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
// handleUDPAssociate implements SOCKS5 UDP ASSOCIATE (RFC 1928) for full Wintun TUN mode
func (s *SocksProxyServer) handleUDPAssociate(clientConn net.Conn, header []byte) {
	// Read address and port specified by client in UDP ASSOCIATE request
	var clientExpectedHost string
	switch header[3] {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(clientConn, ip); err != nil {
			return
		}
		clientExpectedHost = net.IP(ip).String()
	case 0x03: // Domain
		var l [1]byte
		if _, err := io.ReadFull(clientConn, l[:]); err != nil {
			return
		}
		domain := make([]byte, int(l[0]))
		if _, err := io.ReadFull(clientConn, domain); err != nil {
			return
		}
		clientExpectedHost = string(domain)
	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(clientConn, ip); err != nil {
			return
		}
		clientExpectedHost = net.IP(ip).String()
	default:
		clientConn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	_ = clientExpectedHost

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(clientConn, portBuf); err != nil {
		return
	}

	// Listen for UDP datagrams on loopback
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		clientConn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer udpConn.Close()

	relayAddr := udpConn.LocalAddr().(*net.UDPAddr)
	relayPort := uint16(relayAddr.Port)

	// Send success response (0x00) with BND.ADDR and BND.PORT
	resp := []byte{
		0x05, 0x00, 0x00, 0x01,
		127, 0, 0, 1,
		byte(relayPort >> 8), byte(relayPort & 0xff),
	}
	if _, err := clientConn.Write(resp); err != nil {
		return
	}

	// Association terminates when TCP stream closes
	tcpDone := make(chan struct{})
	go func() {
		defer close(tcpDone)
		buf := make([]byte, 256)
		for {
			if _, err := clientConn.Read(buf); err != nil {
				return
			}
		}
	}()

	packetBuf := make([]byte, 65535)
	for {
		select {
		case <-tcpDone:
			return
		default:
		}

		_ = udpConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, senderAddr, err := udpConn.ReadFrom(packetBuf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return
		}

		if n < 10 {
			continue
		}

		// Check SOCKS5 UDP header: RSV(2) == 00 00, FRAG(1) == 00
		if packetBuf[0] != 0 || packetBuf[1] != 0 || packetBuf[2] != 0 {
			continue
		}

		atyp := packetBuf[3]
		offset := 4
		var dstHost string

		switch atyp {
		case 0x01: // IPv4
			if n < offset+4+2 {
				continue
			}
			dstHost = net.IP(packetBuf[offset : offset+4]).String()
			offset += 4
		case 0x03: // Domain
			dLen := int(packetBuf[offset])
			offset++
			if n < offset+dLen+2 {
				continue
			}
			dstHost = string(packetBuf[offset : offset+dLen])
			offset += dLen
		case 0x04: // IPv6
			if n < offset+16+2 {
				continue
			}
			dstHost = net.IP(packetBuf[offset : offset+16]).String()
			offset += 16
		default:
			continue
		}

		dstPort := int(binary.BigEndian.Uint16(packetBuf[offset : offset+2]))
		offset += 2
		payload := packetBuf[offset:n]

		// Port 53 -> Tunnel DNS query securely over MASQUE TCP stream
		if dstPort == 53 {
			go s.forwardDNSOverTunnel(udpConn, senderAddr, dstHost, dstPort, payload)
		} else {
			go s.forwardGenericUDP(udpConn, senderAddr, dstHost, dstPort, payload)
		}
	}
}

// queryDoH executes an RFC 8484 DNS-over-HTTPS request over the encrypted MASQUE tunnel (port 443)
func (s *SocksProxyServer) queryDoH(query []byte) ([]byte, error) {
	endpoints := []struct {
		target string
		host   string
		path   string
	}{
		{"1.1.1.1:443", "cloudflare-dns.com", "/dns-query"},
		{"8.8.8.8:443", "dns.google", "/dns-query"},
		{"9.9.9.9:443", "dns.quad9.net", "/dns-query"},
	}

	for _, ep := range endpoints {
		conn, err := s.opener.OpenTunnel(ep.target)
		if err != nil {
			continue
		}

		tlsCfg := &tls.Config{
			ServerName: ep.host,
			MinVersion: tls.VersionTLS12,
		}
		tlsConn := tls.Client(conn, tlsCfg)
		_ = tlsConn.SetDeadline(time.Now().Add(4 * time.Second))
		if err := tlsConn.Handshake(); err != nil {
			_ = conn.Close()
			continue
		}

		reqStr := fmt.Sprintf("POST %s HTTP/1.1\r\nHost: %s\r\nContent-Type: application/dns-message\r\nAccept: application/dns-message\r\nContent-Length: %d\r\nConnection: close\r\n\r\n", ep.path, ep.host, len(query))
		if _, err := tlsConn.Write([]byte(reqStr)); err != nil {
			_ = tlsConn.Close()
			continue
		}
		if _, err := tlsConn.Write(query); err != nil {
			_ = tlsConn.Close()
			continue
		}

		resp, err := http.ReadResponse(bufio.NewReader(tlsConn), nil)
		if err != nil {
			_ = tlsConn.Close()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		_ = tlsConn.Close()

		if resp.StatusCode == http.StatusOK && len(body) >= 12 {
			return body, nil
		}
	}
	return nil, errors.New("all DoH endpoints failed")
}

// queryTcpDns executes an RFC 7766 DNS-over-TCP query
func (s *SocksProxyServer) queryTcpDns(dnsHost string, dnsPort int, query []byte) ([]byte, error) {
	target := fmt.Sprintf("%s:%d", dnsHost, dnsPort)
	conn, err := s.opener.OpenTunnel(target)
	if err != nil {
		conn, err = s.opener.OpenTunnel("1.1.1.1:53")
		if err != nil {
			return nil, err
		}
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	reqLen := uint16(len(query))
	reqBuf := make([]byte, 2+len(query))
	binary.BigEndian.PutUint16(reqBuf[:2], reqLen)
	copy(reqBuf[2:], query)

	if _, err := conn.Write(reqBuf); err != nil {
		return nil, err
	}

	var respLenBuf [2]byte
	if _, err := io.ReadFull(conn, respLenBuf[:]); err != nil {
		return nil, err
	}
	respLen := binary.BigEndian.Uint16(respLenBuf[:])
	if respLen == 0 || respLen > 4096 {
		return nil, errors.New("invalid tcp dns length")
	}

	dnsResp := make([]byte, respLen)
	if _, err := io.ReadFull(conn, dnsResp); err != nil {
		return nil, err
	}
	return dnsResp, nil
}

// queryDirectUdpDns attempts direct UDP query to public resolver
func (s *SocksProxyServer) queryDirectUdpDns(dnsHost string, dnsPort int, query []byte) ([]byte, error) {
	rAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", dnsHost, dnsPort))
	if err != nil {
		rAddr, _ = net.ResolveUDPAddr("udp", "1.1.1.1:53")
	}
	conn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write(query); err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n < 12 {
		return nil, errors.New("direct udp dns failed")
	}
	return buf[:n], nil
}

// forwardDNSOverTunnel sends a UDP DNS query with RFC 8484 DoH, TCP fallback, and returns SOCKS5 UDP answer
func (s *SocksProxyServer) forwardDNSOverTunnel(udpConn *net.UDPConn, clientAddr net.Addr, dnsHost string, dnsPort int, query []byte) {
	if dnsHost == "" || dnsHost == "0.0.0.0" {
		dnsHost = "1.1.1.1"
	}
	if dnsPort <= 0 {
		dnsPort = 53
	}

	var dnsResp []byte
	var err error

	// 1. Primary: RFC 8484 DNS-over-HTTPS (DoH) via MASQUE tunnel (port 443 - 100% open & unfiltered)
	dnsResp, err = s.queryDoH(query)

	// 2. Fallback: RFC 7766 DNS-over-TCP via MASQUE tunnel (port 53)
	if err != nil || len(dnsResp) == 0 {
		dnsResp, err = s.queryTcpDns(dnsHost, dnsPort, query)
	}

	// 3. Fallback: Direct UDP to public resolver
	if err != nil || len(dnsResp) == 0 {
		dnsResp, err = s.queryDirectUdpDns(dnsHost, dnsPort, query)
	}

	if err != nil || len(dnsResp) == 0 {
		return
	}

	// Pack DNS response in SOCKS5 UDP encapsulation (RFC 1928)
	respHeader := []byte{0x00, 0x00, 0x00, 0x01}
	ipBytes := net.ParseIP(dnsHost).To4()
	if ipBytes == nil {
		ipBytes = net.IPv4(1, 1, 1, 1).To4()
	}
	respHeader = append(respHeader, ipBytes...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(dnsPort))
	respHeader = append(respHeader, portBytes...)

	fullPacket := append(respHeader, dnsResp...)
	if s.meter != nil {
		s.meter.AddUp(int64(len(query)))
		s.meter.AddDown(int64(len(dnsResp)))
	}
	_, _ = udpConn.WriteTo(fullPacket, clientAddr)
}

// forwardGenericUDP dials direct UDP for non-DNS traffic (STUN, NTP, local network)
func (s *SocksProxyServer) forwardGenericUDP(udpConn *net.UDPConn, clientAddr net.Addr, host string, port int, data []byte) {
	rAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return
	}
	directConn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		return
	}
	defer directConn.Close()

	_ = directConn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := directConn.Write(data); err != nil {
		return
	}

	buf := make([]byte, 4096)
	n, err := directConn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	respHeader := []byte{0x00, 0x00, 0x00, 0x01}
	ipBytes := rAddr.IP.To4()
	if ipBytes == nil {
		ipBytes = net.IPv4(0, 0, 0, 0).To4()
	}
	respHeader = append(respHeader, ipBytes...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	respHeader = append(respHeader, portBytes...)

	fullPacket := append(respHeader, buf[:n]...)
	if s.meter != nil {
		s.meter.AddUp(int64(len(data))); s.meter.AddDown(int64(n))
	}
	_, _ = udpConn.WriteTo(fullPacket, clientAddr)
}
