package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

type mockDnsTunnelOpener struct{}

func (m *mockDnsTunnelOpener) OpenTunnel(authority string) (net.Conn, error) {
	client, server := net.Pipe()

	go func() {
		defer server.Close()

		if strings.Contains(authority, ":443") {
			// Mock DoH (RFC 8484) HTTP server
			buf := make([]byte, 4096)
			n, err := server.Read(buf)
			if err != nil {
				return
			}
			req := string(buf[:n])
			// Split headers from body
			parts := bytes.Split(buf[:n], []byte("\r\n\r\n"))
			var query []byte
			if len(parts) > 1 {
				query = parts[1]
			}
			_ = req

			if len(query) >= 12 {
				mockResp := make([]byte, len(query)+16)
				copy(mockResp, query)
				mockResp[2] = 0x81 // Flags: standard response
				mockResp[3] = 0x80

				httpResp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/dns-message\r\nContent-Length: %d\r\nConnection: close\r\n\r\n", len(mockResp))
				_, _ = server.Write([]byte(httpResp))
				_, _ = server.Write(mockResp)
			}
			return
		}

		// Mock DNS over TCP (RFC 7766)
		var lenBuf [2]byte
		if _, err := io.ReadFull(server, lenBuf[:]); err != nil {
			return
		}
		qLen := binary.BigEndian.Uint16(lenBuf[:])
		query := make([]byte, qLen)
		if _, err := io.ReadFull(server, query); err != nil {
			return
		}

		mockResp := make([]byte, len(query)+16)
		copy(mockResp, query)
		mockResp[2] = 0x81 // Flags: standard response
		mockResp[3] = 0x80

		respLenBuf := make([]byte, 2)
		binary.BigEndian.PutUint16(respLenBuf, uint16(len(mockResp)))
		_, _ = server.Write(respLenBuf)
		_, _ = server.Write(mockResp)
	}()

	return client, nil
}

func TestSocks5UDPAssociateAndDnsTunneling(t *testing.T) {
	meter := NewTrafficMeter()
	opener := &mockDnsTunnelOpener{}

	srv := NewSocksProxyServer("127.0.0.1:0", opener, meter)
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start socks proxy: %v", err)
	}
	defer srv.Close()

	// 1. Connect to SOCKS5 server via TCP
	conn, err := net.Dial("tcp", srv.listenAddr)
	if err != nil {
		t.Fatalf("Failed to dial SOCKS5: %v", err)
	}
	defer conn.Close()

	// 2. Auth Handshake (Method: No Auth)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		t.Fatalf("Failed to write auth handshake: %v", err)
	}
	authResp := make([]byte, 2)
	if _, err := io.ReadFull(conn, authResp); err != nil {
		t.Fatalf("Failed to read auth response: %v", err)
	}
	if authResp[0] != 0x05 || authResp[1] != 0x00 {
		t.Fatalf("Auth failed: %v", authResp)
	}

	// 3. Send SOCKS5 UDP ASSOCIATE request (CMD 0x03)
	udpAssocReq := []byte{
		0x05, 0x03, 0x00, 0x01,
		0, 0, 0, 0,
		0, 0,
	}
	if _, err := conn.Write(udpAssocReq); err != nil {
		t.Fatalf("Failed to send UDP ASSOCIATE request: %v", err)
	}

	assocResp := make([]byte, 10)
	if _, err := io.ReadFull(conn, assocResp); err != nil {
		t.Fatalf("Failed to read UDP ASSOCIATE response: %v", err)
	}
	if assocResp[1] != 0x00 {
		t.Fatalf("UDP ASSOCIATE failed with rep code %x", assocResp[1])
	}

	relayIP := net.IP(assocResp[4:8]).String()
	relayPort := binary.BigEndian.Uint16(assocResp[8:10])
	relayAddr := fmt.Sprintf("%s:%d", relayIP, relayPort)
	t.Logf("SOCKS5 UDP ASSOCIATE succeeded! Relay address: %s", relayAddr)

	// 4. Send DNS query packet to UDP relay
	udpRelayConn, err := net.Dial("udp", relayAddr)
	if err != nil {
		t.Fatalf("Failed to dial UDP relay: %v", err)
	}
	defer udpRelayConn.Close()

	dnsQueryPayload := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Standard query
		0x00, 0x01, // QDCOUNT = 1
		0x00, 0x00, // ANCOUNT = 0
		0x00, 0x00, // NSCOUNT = 0
		0x00, 0x00, // ARCOUNT = 0
		0x06, 'g', 'o', 'o', 'g', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // End of name
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	// SOCKS5 UDP Encapsulation targeting 1.1.1.1:53
	socksUdpPacket := []byte{
		0x00, 0x00, // RSV
		0x00,       // FRAG
		0x01,       // ATYP: IPv4
		1, 1, 1, 1, // Target IP: 1.1.1.1
		0x00, 0x35, // Target Port: 53
	}
	socksUdpPacket = append(socksUdpPacket, dnsQueryPayload...)

	if _, err := udpRelayConn.Write(socksUdpPacket); err != nil {
		t.Fatalf("Failed to send UDP packet: %v", err)
	}

	// 5. Read DNS response from UDP relay
	_ = udpRelayConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	respBuf := make([]byte, 4096)
	n, err := udpRelayConn.Read(respBuf)
	if err != nil {
		t.Fatalf("Failed to receive DNS response from relay: %v", err)
	}

	if n < 10+len(dnsQueryPayload) {
		t.Fatalf("DNS response too short: %d bytes", n)
	}

	payloadResp := respBuf[10:n]
	if payloadResp[0] != 0x12 || payloadResp[1] != 0x34 {
		t.Fatalf("Transaction ID mismatch in DNS response: got %x %x", payloadResp[0], payloadResp[1])
	}

	t.Logf("DNS Over Tunneling Verified! Response size: %d bytes, Transaction ID: %x%x", len(payloadResp), payloadResp[0], payloadResp[1])
}
