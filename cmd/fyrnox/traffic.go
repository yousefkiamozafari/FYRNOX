package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

type TrafficMeter struct {
	bytesDown atomic.Int64
	bytesUp   atomic.Int64

	lastBytesDown int64
	lastBytesUp   int64
	lastTick      time.Time

	speedDown atomic.Int64 // bytes per second
	speedUp   atomic.Int64 // bytes per second

	sessionStart time.Time
	sessionEnd   time.Time
	active       atomic.Bool
}

func NewTrafficMeter() *TrafficMeter {
	now := time.Now()
	m := &TrafficMeter{
		lastTick:     now,
		sessionStart: now,
	}
	m.active.Store(false)
	go m.tickerLoop()
	return m
}

func (m *TrafficMeter) AddDown(n int64) {
	if n > 0 {
		m.bytesDown.Add(n)
	}
}

func (m *TrafficMeter) AddUp(n int64) {
	if n > 0 {
		m.bytesUp.Add(n)
	}
}

// StartSession begins or resets the cumulative VPN session counters
func (m *TrafficMeter) StartSession() {
	m.bytesDown.Store(0)
	m.bytesUp.Store(0)
	m.speedDown.Store(0)
	m.speedUp.Store(0)
	m.lastBytesDown = 0
	m.lastBytesUp = 0
	now := time.Now()
	m.lastTick = now
	m.sessionStart = now
	m.sessionEnd = time.Time{}
	m.active.Store(true)
}

// StopSession halts active speed calculation and freezes session duration,
// preserving the cumulative bytes downloaded and uploaded for display.
func (m *TrafficMeter) StopSession() {
	if m.active.Swap(false) {
		m.sessionEnd = time.Now()
	}
	m.speedDown.Store(0)
	m.speedUp.Store(0)
}

// Reset resets the session to zero (used when starting fresh connection)
func (m *TrafficMeter) Reset() {
	m.StartSession()
}

func (m *TrafficMeter) tickerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for now := range ticker.C {
		if !m.active.Load() {
			m.speedDown.Store(0)
			m.speedUp.Store(0)
			continue
		}

		currentDown := m.bytesDown.Load()
		currentUp := m.bytesUp.Load()

		dt := now.Sub(m.lastTick).Seconds()
		if dt <= 0 {
			dt = 1
		}

		diffDown := currentDown - m.lastBytesDown
		diffUp := currentUp - m.lastBytesUp
		if diffDown < 0 {
			diffDown = 0
		}
		if diffUp < 0 {
			diffUp = 0
		}

		m.speedDown.Store(int64(float64(diffDown) / dt))
		m.speedUp.Store(int64(float64(diffUp) / dt))

		m.lastBytesDown = currentDown
		m.lastBytesUp = currentUp
		m.lastTick = now
	}
}

type TrafficStats struct {
	BytesDown         int64  `json:"bytes_down"`
	BytesUp           int64  `json:"bytes_up"`
	SpeedDown         int64  `json:"speed_down"`
	SpeedUp           int64  `json:"speed_up"`
	SpeedDownHuman    string `json:"speed_down_human"`
	SpeedUpHuman      string `json:"speed_up_human"`
	TotalDownHuman    string `json:"total_down_human"`
	TotalUpHuman      string `json:"total_up_human"`
	BytesReadHuman    string `json:"bytes_read_human"`
	BytesWrittenHuman string `json:"bytes_written_human"`
	DurationSec       int64  `json:"duration_sec"`
}

func (m *TrafficMeter) GetStats() TrafficStats {
	down := m.bytesDown.Load()
	up := m.bytesUp.Load()
	sDown := m.speedDown.Load()
	sUp := m.speedUp.Load()

	var duration int64
	if m.active.Load() {
		duration = int64(time.Since(m.sessionStart).Seconds())
	} else if !m.sessionEnd.IsZero() {
		duration = int64(m.sessionEnd.Sub(m.sessionStart).Seconds())
		sDown = 0
		sUp = 0
	} else {
		duration = 0
		sDown = 0
		sUp = 0
	}

	humanDown := FormatBytes(down)
	humanUp := FormatBytes(up)

	return TrafficStats{
		BytesDown:         down,
		BytesUp:           up,
		SpeedDown:         sDown,
		SpeedUp:           sUp,
		SpeedDownHuman:    FormatSpeed(sDown),
		SpeedUpHuman:      FormatSpeed(sUp),
		TotalDownHuman:    humanDown,
		TotalUpHuman:      humanUp,
		BytesReadHuman:    humanDown,
		BytesWrittenHuman: humanUp,
		DurationSec:       duration,
	}
}

func FormatBytes(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatSpeed(bps int64) string {
	if bps <= 0 {
		return "0 B/s"
	}
	return FormatBytes(bps) + "/s"
}
