package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// MonthlyQuotaAllowanceBytes is 50.0 GB (50 * 1024 * 1024 * 1024)
	MonthlyQuotaAllowanceBytes = int64(50 * 1024 * 1024 * 1024)
)

type MonthlyUsageRecord struct {
	Year        int       `json:"year"`
	Month       int       `json:"month"`
	UsedBytes   int64     `json:"used_bytes"`
	LastUpdated time.Time `json:"last_updated"`
}

type MonthlyQuotaTracker struct {
	mu       sync.Mutex
	filePath string
	record   MonthlyUsageRecord
}

func getAppDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		configDir = os.Getenv("APPDATA")
	}
	if configDir == "" {
		configDir = "."
	}
	dir := filepath.Join(configDir, "Fyrnox")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func NewMonthlyQuotaTracker(customPath string) *MonthlyQuotaTracker {
	p := customPath
	if p == "" {
		p = filepath.Join(getAppDataDir(), "monthly_usage.json")
	}

	tracker := &MonthlyQuotaTracker{
		filePath: p,
	}
	tracker.load()
	return tracker
}

func (t *MonthlyQuotaTracker) load() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	data, err := os.ReadFile(t.filePath)
	if err == nil {
		var rec MonthlyUsageRecord
		if jErr := json.Unmarshal(data, &rec); jErr == nil {
			// Check if month rolled over
			if rec.Year == now.Year() && rec.Month == int(now.Month()) {
				t.record = rec
				return
			}
		}
	}

	// New month or new file
	t.record = MonthlyUsageRecord{
		Year:        now.Year(),
		Month:       int(now.Month()),
		UsedBytes:   0,
		LastUpdated: now,
	}
	t.saveLocked()
}

func (t *MonthlyQuotaTracker) saveLocked() {
	data, err := json.MarshalIndent(t.record, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(t.filePath, data, 0644)
}

func (t *MonthlyQuotaTracker) AddBytes(n int64) {
	if n <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if t.record.Year != now.Year() || t.record.Month != int(now.Month()) {
		// Rollover to new month
		t.record.Year = now.Year()
		t.record.Month = int(now.Month())
		t.record.UsedBytes = 0
	}

	t.record.UsedBytes += n
	t.record.LastUpdated = now

	// Save periodically or on significant change
	t.saveLocked()
}

// GetNextMonthlyResetDate calculates the first day of the next calendar month
func GetNextMonthlyResetDate(now time.Time) (time.Time, string) {
	year := now.Year()
	month := now.Month() + 1
	if month > 12 {
		month = 1
		year++
	}
	reset := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return reset, reset.Format(time.RFC3339)
}

func (t *MonthlyQuotaTracker) GetMonthlyUsage() (usedBytes int64, remainingBytes int64, resetDate string, pct float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if t.record.Year != now.Year() || t.record.Month != int(now.Month()) {
		t.record.Year = now.Year()
		t.record.Month = int(now.Month())
		t.record.UsedBytes = 0
		t.saveLocked()
	}

	usedBytes = t.record.UsedBytes
	remainingBytes = MonthlyQuotaAllowanceBytes - usedBytes
	if remainingBytes < 0 {
		remainingBytes = 0
	}

	if MonthlyQuotaAllowanceBytes > 0 {
		pct = float64(usedBytes) / float64(MonthlyQuotaAllowanceBytes) * 100
		if pct > 100 {
			pct = 100
		}
	}

	_, resetDate = GetNextMonthlyResetDate(now)
	return usedBytes, remainingBytes, resetDate, pct
}

func (t *MonthlyQuotaTracker) SetUsedForTesting(bytes int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.record.UsedBytes = bytes
	t.saveLocked()
}
