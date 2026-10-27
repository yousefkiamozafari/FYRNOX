package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestMonthlyQuotaTrackerCalculationsAndPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_monthly_usage.json")

	tracker := NewMonthlyQuotaTracker(filePath)

	// Initially 0 bytes used
	used, remaining, resetDate, pct := tracker.GetMonthlyUsage()
	if used != 0 {
		t.Errorf("Expected initial used 0, got %d", used)
	}
	if remaining != MonthlyQuotaAllowanceBytes {
		t.Errorf("Expected initial remaining %d, got %d", MonthlyQuotaAllowanceBytes, remaining)
	}
	if pct != 0 {
		t.Errorf("Expected 0%% used, got %f", pct)
	}
	if resetDate == "" {
		t.Errorf("Expected non-empty resetDate")
	}

	// Add 5 GB (5 * 1024 * 1024 * 1024)
	fiveGB := int64(5 * 1024 * 1024 * 1024)
	tracker.AddBytes(fiveGB)

	used, remaining, _, pct = tracker.GetMonthlyUsage()
	if used != fiveGB {
		t.Errorf("Expected used %d, got %d", fiveGB, used)
	}
	expectedRemaining := MonthlyQuotaAllowanceBytes - fiveGB
	if remaining != expectedRemaining {
		t.Errorf("Expected remaining %d, got %d", expectedRemaining, remaining)
	}
	if pct < 9.9 || pct > 10.1 {
		t.Errorf("Expected ~10%% used, got %f", pct)
	}

	// Verify persistence: create new tracker from same file
	tracker2 := NewMonthlyQuotaTracker(filePath)
	used2, remaining2, _, pct2 := tracker2.GetMonthlyUsage()
	if used2 != fiveGB {
		t.Errorf("Persisted tracker expected used %d, got %d", fiveGB, used2)
	}
	if remaining2 != expectedRemaining {
		t.Errorf("Persisted tracker expected remaining %d, got %d", expectedRemaining, remaining2)
	}
	if pct2 < 9.9 || pct2 > 10.1 {
		t.Errorf("Persisted tracker expected ~10%%, got %f", pct2)
	}

	// Test month rollover simulation
	tracker.mu.Lock()
	tracker.record.Month = int(time.Now().Month()) - 1
	if tracker.record.Month == 0 {
		tracker.record.Month = 12
		tracker.record.Year--
	}
	tracker.saveLocked()
	tracker.mu.Unlock()

	// New tracker loading old month data must rollover to 0 used bytes
	tracker3 := NewMonthlyQuotaTracker(filePath)
	used3, _, _, _ := tracker3.GetMonthlyUsage()
	if used3 != 0 {
		t.Errorf("Rollover expected used 0, got %d", used3)
	}
}

func TestGetNextMonthlyResetDate(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	reset, formatted := GetNextMonthlyResetDate(now)

	if reset.Year() != 2026 || reset.Month() != 11 || reset.Day() != 1 {
		t.Errorf("Expected 2026-11-01, got %v", reset)
	}
	if formatted != "2026-11-01T00:00:00Z" {
		t.Errorf("Expected 2026-11-01T00:00:00Z, got %s", formatted)
	}

	// Test year boundary
	december := time.Date(2026, 12, 15, 10, 0, 0, 0, time.UTC)
	resetDec, _ := GetNextMonthlyResetDate(december)
	if resetDec.Year() != 2027 || resetDec.Month() != 1 || resetDec.Day() != 1 {
		t.Errorf("Expected 2027-01-01, got %v", resetDec)
	}
}
