package main

import (
	"testing"
)

func TestKillSwitchLifecycle(t *testing.T) {
	// 1. Initial state should be inactive
	if globalKillSwitch.IsActive() {
		t.Fatalf("expected KillSwitch initially inactive")
	}

	// 2. Disarm should be idempotent and safe
	err := globalKillSwitch.Disarm()
	if err != nil {
		t.Fatalf("disarm error: %v", err)
	}

	// 3. Cleanup should be safe
	CleanupStaleKillSwitchRules()
}
