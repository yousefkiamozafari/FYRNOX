package main

import (
	"testing"
)

func TestTrafficMeterStatsAndMapping(t *testing.T) {
	meter := NewTrafficMeter()
	meter.Reset()

	stats := meter.GetStats()
	if stats.BytesDown != 0 || stats.BytesUp != 0 {
		t.Fatalf("expected 0 bytes, got down=%d up=%d", stats.BytesDown, stats.BytesUp)
	}
	if stats.TotalDownHuman != "0 B" || stats.BytesReadHuman != "0 B" {
		t.Fatalf("expected '0 B' for initial human stats, got %q / %q", stats.TotalDownHuman, stats.BytesReadHuman)
	}

	// Simulate transferring 10.5 MB down and 2.5 MB up
	meter.AddDown(10 * 1024 * 1024 + 512*1024)
	meter.AddUp(2 * 1024 * 1024 + 512*1024)

	stats = meter.GetStats()
	if stats.BytesDown <= 0 || stats.BytesUp <= 0 {
		t.Fatalf("expected positive byte counters, got down=%d up=%d", stats.BytesDown, stats.BytesUp)
	}
	if stats.TotalDownHuman == "0 B" || stats.BytesReadHuman == "0 B" {
		t.Fatalf("expected formatted human strings, got %q / %q", stats.TotalDownHuman, stats.BytesReadHuman)
	}
	if stats.TotalDownHuman != stats.BytesReadHuman {
		t.Fatalf("expected TotalDownHuman == BytesReadHuman, got %q vs %q", stats.TotalDownHuman, stats.BytesReadHuman)
	}
	if stats.TotalUpHuman != stats.BytesWrittenHuman {
		t.Fatalf("expected TotalUpHuman == BytesWrittenHuman, got %q vs %q", stats.TotalUpHuman, stats.BytesWrittenHuman)
	}

	t.Logf("Traffic stats verified: Down=%s (%d bytes), Up=%s (%d bytes)", stats.TotalDownHuman, stats.BytesDown, stats.TotalUpHuman, stats.BytesUp)
}
