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
}

func NewTrafficMeter() *TrafficMeter {
	m := &TrafficMeter{
		lastTick:     time.Now(),
		sessionStart: time.Now(),
	}
	go m.tickerLoop()
	return m
}

func (m *TrafficMeter) AddDown(n int64) {
	m.bytesDown.Add(n)
}

func (m *TrafficMeter) AddUp(n int64) {
	m.bytesUp.Add(n)
}

func (m *TrafficMeter) Reset() {
	m.bytesDown.Store(0)
	m.bytesUp.Store(0)
	m.speedDown.Store(0)
	m.speedUp.Store(0)
	m.lastBytesDown = 0
	m.lastBytesUp = 0
	m.lastTick = time.Now()
	m.sessionStart = time.Now()
}

func (m *TrafficMeter) tickerLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for now := range ticker.C {
		currentDown := m.bytesDown.Load()
		currentUp := m.bytesUp.Load()

		dt := now.Sub(m.lastTick).Seconds()
		if dt <= 0 {
			dt = 1
		}

		diffDown := currentDown - m.lastBytesDown
		diffUp := currentUp - m.lastBytesUp

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

	duration := int64(time.Since(m.sessionStart).Seconds())
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
	return FormatBytes(bps) + "/s"
}
