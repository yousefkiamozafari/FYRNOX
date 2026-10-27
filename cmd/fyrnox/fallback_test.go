package main

import (
	"net/url"
	"testing"
)

func TestFallbackManagerTriggerAndRestore(t *testing.T) {
	fakeURL, _ := url.Parse("https://127.0.0.1:443")
	switchedTo := ""

	fm := NewFallbackManager(fakeURL, "", "dual", true, func(useH3 bool) error {
		if useH3 {
			switchedTo = "h3"
		} else {
			switchedTo = "h2"
		}
		return nil
	})
	defer fm.Stop()

	// Initial state
	active, proto, rate := fm.GetStatus()
	if active || proto != "optimized_udp" || rate != 0.0 {
		t.Fatalf("Unexpected initial state: active=%v, proto=%s, rate=%f", active, proto, rate)
	}

	// Record 2 failures (< 3 samples, shouldn't trigger yet)
	fm.RecordUDPResult(false, 100)
	fm.RecordUDPResult(false, 120)

	active, _, _ = fm.GetStatus()
	if active {
		t.Fatalf("Should not trigger fallback before 3 samples")
	}

	// 3rd failure -> should immediately trigger fallback!
	fm.RecordUDPResult(false, 150)

	active, proto, _ = fm.GetStatus()
	if !active || proto != "protected_tcp" {
		t.Fatalf("Fallback should be active now: active=%v, proto=%s", active, proto)
	}
	if switchedTo != "h2" {
		t.Fatalf("Switch callback should have been called with h2, got %s", switchedTo)
	}

	// Trigger manual restore
	fm.TriggerRestoreUDP()

	active, proto, _ = fm.GetStatus()
	if active || proto != "optimized_udp" {
		t.Fatalf("Expected restored to UDP: active=%v, proto=%s", active, proto)
	}
	if switchedTo != "h3" {
		t.Fatalf("Switch callback should have been called with h3, got %s", switchedTo)
	}
}
