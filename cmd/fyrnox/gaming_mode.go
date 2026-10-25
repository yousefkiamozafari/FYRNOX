package main

import (
	"fmt"
	"math"
	"net"
	"sync"
	"time"
)

type GamingStats struct {
	Active     bool    `json:"active"`
	PingMs     int     `json:"ping_ms"`
	JitterMs   int     `json:"jitter_ms"`
	PacketLoss float64 `json:"packet_loss"`
	Mtu        int     `json:"mtu"`
}

type GamingAccelerator struct {
	mu          sync.RWMutex
	active      bool
	stopCh      chan struct{}
	lastPing    int
	jitter      int
	packetLoss  float64
	historyRTT  []int
	historyLost int
}

var globalGamer = &GamingAccelerator{
	historyRTT: make([]int, 0, 20),
}

// SetGamingMode enables or disables low-latency gaming mode
func (g *GamingAccelerator) SetGamingMode(enable bool) {
	g.mu.Lock()
	if g.active == enable {
		g.mu.Unlock()
		return
	}
	g.active = enable

	if enable {
		g.stopCh = make(chan struct{})
		g.lastPing = 0
		g.jitter = 0
		g.packetLoss = 0
		g.historyRTT = g.historyRTT[:0]
		g.historyLost = 0
		g.mu.Unlock()

		// Optimize MTU on Wintun to 1420 to prevent fragmentation
		_ = newSilentCmd("netsh", "interface", "ipv4", "set", "subinterface", fmt.Sprintf("name=%s", tunDeviceName), "mtu=1420", "store=persistent").Run()
		LogInfo("Gaming Mode Activated: Wintun MTU optimized to 1420 (Zero-Fragmentation)")

		// Start background UDP jitter/packet-loss telemetry loop
		go g.telemetryLoop(g.stopCh)
	} else {
		if g.stopCh != nil {
			close(g.stopCh)
			g.stopCh = nil
		}
		g.mu.Unlock()

		// Restore standard MTU 1500
		_ = newSilentCmd("netsh", "interface", "ipv4", "set", "subinterface", fmt.Sprintf("name=%s", tunDeviceName), "mtu=1500", "store=persistent").Run()
		LogInfo("Gaming Mode Deactivated: Wintun MTU restored to 1500")
	}
}

func (g *GamingAccelerator) telemetryLoop(stopCh chan struct{}) {
	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	// Probing targets: Cloudflare & Google primary DNS UDP ports
	probes := []string{"1.1.1.1:53", "8.8.8.8:53"}
	probeIdx := 0

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			target := probes[probeIdx%len(probes)]
			probeIdx++

			start := time.Now()
			conn, err := net.DialTimeout("udp", target, 800*time.Millisecond)
			if err != nil {
				g.recordProbe(-1)
				continue
			}

			// Send simple DNS query packet to trigger response
			dnsQuery := []byte{
				0xAA, 0xBB, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x07, 'e', 'x', 'a',
				'm', 'p', 'l', 'e', 0x03, 'c', 'o', 'm',
				0x00, 0x00, 0x01, 0x00, 0x01,
			}
			_ = conn.SetDeadline(time.Now().Add(800 * time.Millisecond))
			_, err = conn.Write(dnsQuery)
			if err != nil {
				_ = conn.Close()
				g.recordProbe(-1)
				continue
			}

			resp := make([]byte, 512)
			_, err = conn.Read(resp)
			_ = conn.Close()

			if err != nil {
				g.recordProbe(-1)
			} else {
				rtt := int(time.Since(start).Milliseconds())
				g.recordProbe(rtt)
			}
		}
	}
}

func (g *GamingAccelerator) recordProbe(rtt int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	const windowSize = 15

	if rtt < 0 {
		g.historyLost++
		if len(g.historyRTT) >= windowSize {
			g.historyRTT = g.historyRTT[1:]
		}
		g.historyRTT = append(g.historyRTT, -1)
	} else {
		if g.lastPing > 0 {
			diff := int(math.Abs(float64(rtt - g.lastPing)))
			if g.jitter == 0 {
				g.jitter = diff
			} else {
				// Smooth moving average for jitter
				g.jitter = (g.jitter*3 + diff) / 4
			}
		}
		g.lastPing = rtt

		if len(g.historyRTT) >= windowSize {
			g.historyRTT = g.historyRTT[1:]
		}
		g.historyRTT = append(g.historyRTT, rtt)
	}

	// Calculate packet loss over rolling window
	lostCount := 0
	totalCount := len(g.historyRTT)
	for _, val := range g.historyRTT {
		if val < 0 {
			lostCount++
		}
	}
	if totalCount > 0 {
		g.packetLoss = (float64(lostCount) / float64(totalCount)) * 100.0
	}
}

// GetStats returns the current gaming telemetry metrics
func (g *GamingAccelerator) GetStats() GamingStats {
	g.mu.RLock()
	defer g.mu.RUnlock()

	mtu := 1500
	if g.active {
		mtu = 1420
	}

	return GamingStats{
		Active:     g.active,
		PingMs:     g.lastPing,
		JitterMs:   g.jitter,
		PacketLoss: math.Round(g.packetLoss*10) / 10,
		Mtu:        mtu,
	}
}
