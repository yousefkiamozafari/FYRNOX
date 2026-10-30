package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

type mockDnsTunnelOpener struct{}

func (m *mockDnsTunnelOpener) OpenTunnel(authority string) (net.Conn, error) {
	client, server := net.Pipe()

	go func() {
		defer server.Close()
		// Read 2-byte length
		var lenBuf [2]byte
		if _, err := io.ReadFull(server, lenBuf[:]); err != nil {
			return
		}
		qLen := binary.BigEndian.Uint16(lenBuf[:])
		query := make([]byte, qLen)
		if _, err := io.ReadFull(server, query); err != nil {
			return
		}

		// Synthesize a valid DNS response matching Transaction ID
		mockResp := make([]byte, len(query)+16)
		copy(mockResp, query)
		mockResp[2] = 0x81 // Flags: standard response, no error
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

	// 2. Handshake: No Auth
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatalf("Handshake write failed: %v", err)
	}
	hsResp := make([]byte, 2)
	if _, err := io.ReadFull(conn, hsResp); err != nil || hsResp[0] != 0x05 || hsResp[1] != 0x00 {
		t.Fatalf("Handshake failed, got: %v, err: %v", hsResp, err)
	}

	// 3. Send UDP ASSOCIATE request: CMD 0x03
	req := []byte{0x05, 0x03, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
	if _, err := conn.Write(req); err != nil {
		t.Fatalf("UDP ASSOCIATE write failed: %v", err)
	}

	resp := make([]byte, 10)
	if _, err := io.ReadFull(conn, resp); err != nil {
		t.Fatalf("Failed to read UDP ASSOCIATE response: %v", err)
	}
	if resp[0] != 0x05 || resp[1] != 0x00 {
		t.Fatalf("UDP ASSOCIATE rejected with code: %d", resp[1])
	}

	relayPort := binary.BigEndian.Uint16(resp[8:10])
	if relayPort == 0 {
		t.Fatalf("Expected non-zero UDP relay port, got %d", relayPort)
	}
	relayAddr := fmt.Sprintf("127.0.0.1:%d", relayPort)
	t.Logf("SOCKS5 UDP ASSOCIATE succeeded! Relay address: %s", relayAddr)

	// 4. Send DNS query packet to UDP relay
	udpClient, err := net.ListenUDP("udp", nil)
	if err != nil {
		t.Fatalf("Failed to create UDP client: %v", err)
	}
	defer udpClient.Close()

	rAddr, err := net.ResolveUDPAddr("udp", relayAddr)
	if err != nil {
		t.Fatalf("Resolve relay UDP address failed: %v", err)
	}

	// Construct SOCKS5 UDP Encapsulation targeting 1.1.1.1:53
	dnsQueryPayload := []byte{
		0x12, 0x34, // Transaction ID
		0x01, 0x00, // Standard query
		0x00, 0x01, // Questions: 1
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x06, 'g', 'o', 'o', 'g', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00,
		0x00, 0x01, 0x00, 0x01, // Type A, Class IN
	}

	socksUdpPacket := []byte{
		0x00, 0x00, 0x00, 0x01, // RSV(2), FRAG(1), ATYP(1) = IPv4
		1, 1, 1, 1, // Target IP: 1.1.1.1
		0x00, 53, // Target Port: 53
	}
	socksUdpPacket = append(socksUdpPacket, dnsQueryPayload...)

	if _, err := udpClient.WriteTo(socksUdpPacket, rAddr); err != nil {
		t.Fatalf("Write to UDP relay failed: %v", err)
	}

	// 5. Read DNS response from UDP relay
	_ = udpClient.SetReadDeadline(time.Now().Add(2 * time.Second))
	respBuf := make([]byte, 2048)
	n, _, err := udpClient.ReadFrom(respBuf)
	if err != nil {
		t.Fatalf("Failed to receive DNS response from relay: %v", err)
	}
	if n < 10+len(dnsQueryPayload) {
		t.Fatalf("Response packet too short: %d bytes", n)
	}

	// Verify SOCKS5 header in response
	if respBuf[0] != 0 || respBuf[1] != 0 || respBuf[2] != 0 || respBuf[3] != 1 {
		t.Fatalf("Invalid SOCKS5 UDP header in response: %v", respBuf[:4])
	}

	payloadResp := respBuf[10:n]
	if payloadResp[0] != 0x12 || payloadResp[1] != 0x34 {
		t.Fatalf("Transaction ID mismatch in DNS response: got %x %x", payloadResp[0], payloadResp[1])
	}

	t.Logf("DNS Over Tunneling Verified! Response size: %d bytes, Transaction ID: %x%x", len(payloadResp), payloadResp[0], payloadResp[1])
}
