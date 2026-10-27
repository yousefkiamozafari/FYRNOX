package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

type udpSample struct {
	timestamp time.Time
	success   bool
	latencyMs int
}

// FallbackManager provides dynamic, self-healing failover from UDP (H3) to TCP (H2)
// with zero adapter teardown and automated 60s recovery detection.
type FallbackManager struct {
	mu                   sync.Mutex
	enabled              bool
	isFallbackActive     bool
	windowDuration       time.Duration
	samples              []udpSample
	proxyURL             *url.URL
	apiProxy             string
	ipProtocol           string
	switchCallback       func(useH3 bool) error
	stopCh               chan struct{}
	lastFallbackTime     time.Time
	udpRecoveryThreshold int
	udpSuccessStreak     int
}

func NewFallbackManager(proxyURL *url.URL, apiProxy string, ipProtocol string, useH3 bool, switchCallback func(useH3 bool) error) *FallbackManager {
	fm := &FallbackManager{
		enabled:              useH3, // Only active if UDP/H3 was originally chosen
		isFallbackActive:     false,
		windowDuration:       10 * time.Second,
		samples:              make([]udpSample, 0, 32),
		proxyURL:             proxyURL,
		apiProxy:             apiProxy,
		ipProtocol:           ipProtocol,
		switchCallback:       switchCallback,
		stopCh:               make(chan struct{}),
		udpRecoveryThreshold: 2, // 2 consecutive successful UDP probes to switch back
	}

	if useH3 {
		go fm.recoveryLoop()
	}
	return fm
}

// RecordUDPResult adds a sample to the sliding 10s health window
func (fm *FallbackManager) RecordUDPResult(success bool, latencyMs int) {
	if !fm.enabled {
		return
	}

	fm.mu.Lock()
	now := time.Now()
	fm.samples = append(fm.samples, udpSample{
		timestamp: now,
		success:   success,
		latencyMs: latencyMs,
	})

	// Prune samples older than windowDuration (10s)
	cutoff := now.Add(-fm.windowDuration)
	validIdx := 0
	for i, s := range fm.samples {
		if s.timestamp.After(cutoff) {
			validIdx = i
			break
		}
	}
	if validIdx > 0 {
		fm.samples = fm.samples[validIdx:]
	}

	// Calculate loss rate in window
	total := len(fm.samples)
	failed := 0
	for _, s := range fm.samples {
		if !s.success {
			failed++
		}
	}
	lossRate := 0.0
	if total > 0 {
		lossRate = float64(failed) / float64(total)
	}

	shouldTrigger := !fm.isFallbackActive && total >= 3 && (lossRate >= 0.30 || failed >= 3)
	fm.mu.Unlock()

	if shouldTrigger {
		fm.TriggerFallback(fmt.Sprintf("UDP packet loss/timeout (%.0f%% in 10s window)", lossRate*100))
	}
}

// TriggerFallback immediately switches session to TLS-fragmented TCP/H2
func (fm *FallbackManager) TriggerFallback(reason string) {
	fm.mu.Lock()
	if !fm.enabled || fm.isFallbackActive {
		fm.mu.Unlock()
		return
	}
	fm.isFallbackActive = true
	fm.lastFallbackTime = time.Now()
	fm.udpSuccessStreak = 0
	cb := fm.switchCallback
	fm.mu.Unlock()

	LogWarn("[FALLBACK] %s -> Initiating seamless in-memory switch to TLS-fragmented TCP/H2", reason)
	if cb != nil {
		if err := cb(false); err != nil {
			LogError("[FALLBACK] Failed to switch to TCP/H2: %v", err)
		} else {
			LogInfo("[FALLBACK] Successfully switched transport to Protected Fallback (TCP/H2)")
		}
	}
}

// TriggerRestoreUDP seamlessly switches back to low-latency UDP
func (fm *FallbackManager) TriggerRestoreUDP() {
	fm.mu.Lock()
	if !fm.isFallbackActive {
		fm.mu.Unlock()
		return
	}
	fm.isFallbackActive = false
	fm.samples = fm.samples[:0] // Reset sample window
	cb := fm.switchCallback
	fm.mu.Unlock()

	LogInfo("[FALLBACK] UDP link recovery confirmed stable. Switching back to Optimized UDP (H3)")
	if cb != nil {
		if err := cb(true); err != nil {
			LogWarn("[FALLBACK] Restore to UDP/H3 failed: %v", err)
		} else {
			LogInfo("[FALLBACK] Successfully restored transport to Optimized UDP (H3)")
		}
	}
}

// GetStatus returns current protocol status and fallback metrics
func (fm *FallbackManager) GetStatus() (fallbackActive bool, protocolState string, lossRate float64) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if !fm.enabled {
		return false, "standard_tcp", 0.0
	}

	now := time.Now()
	cutoff := now.Add(-fm.windowDuration)
	total := 0
	failed := 0
	for _, s := range fm.samples {
		if s.timestamp.After(cutoff) {
			total++
			if !s.success {
				failed++
			}
		}
	}
	rate := 0.0
	if total > 0 {
		rate = float64(failed) / float64(total)
	}

	if fm.isFallbackActive {
		return true, "protected_tcp", rate
	}
	return false, "optimized_udp", rate
}

// probeUDP attempts a quick UDP QUIC handshake to check if UDP has stabilized
func (fm *FallbackManager) probeUDP(timeout time.Duration) bool {
	if fm.proxyURL == nil {
		return false
	}
	host := fm.proxyURL.Hostname()
	portStr := fm.proxyURL.Port()
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
			return false
		}
	}

	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return false
	}
	defer udpConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tlsCfg := createProxyTLSConfig(host, []string{"h3"})
	tlsCfg.MinVersion = tls.VersionTLS13

	qConn, err := quic.Dial(ctx, udpConn, udpAddr, tlsCfg, &quic.Config{
		HandshakeIdleTimeout: timeout,
	})
	if err != nil {
		return false
	}
	_ = qConn.CloseWithError(0, "probe ok")
	return true
}

// recoveryLoop probes UDP in background every 60s when fallback is active
func (fm *FallbackManager) recoveryLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-fm.stopCh:
			return
		case <-ticker.C:
			fm.mu.Lock()
			active := fm.isFallbackActive
			sinceFallback := time.Since(fm.lastFallbackTime)
			fm.mu.Unlock()

			// Check after at least 45 seconds in fallback
			if active && sinceFallback >= 45*time.Second {
				LogInfo("[FALLBACK] Probing UDP recovery...")
				ok := fm.probeUDP(4 * time.Second)
				fm.mu.Lock()
				if ok {
					fm.udpSuccessStreak++
					LogInfo("[FALLBACK] UDP probe succeeded (%d/%d)", fm.udpSuccessStreak, fm.udpRecoveryThreshold)
					if fm.udpSuccessStreak >= fm.udpRecoveryThreshold {
						fm.mu.Unlock()
						fm.TriggerRestoreUDP()
						continue
					}
				} else {
					fm.udpSuccessStreak = 0
					LogWarn("[FALLBACK] UDP probe still failing; staying on Protected TCP")
				}
				fm.mu.Unlock()
			}
		}
	}
}

// Stop shuts down the recovery loop
func (fm *FallbackManager) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	select {
	case <-fm.stopCh:
	default:
		close(fm.stopCh)
	}
}
