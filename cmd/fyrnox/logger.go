package main

import (
	"fmt"
	"sync"
	"time"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
}

type AppLogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
}

var globalLogBuffer = &AppLogBuffer{
	entries: make([]LogEntry, 0, 500),
	maxSize: 500,
}

func Log(level LogLevel, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp: time.Now().Format("15:04:05"),
		Level:     level,
		Message:   msg,
	}

	globalLogBuffer.mu.Lock()
	if len(globalLogBuffer.entries) >= globalLogBuffer.maxSize {
		globalLogBuffer.entries = globalLogBuffer.entries[1:]
	}
	globalLogBuffer.entries = append(globalLogBuffer.entries, entry)
	globalLogBuffer.mu.Unlock()

	fmt.Printf("[%s] [%s] %s\n", entry.Timestamp, level, msg)
}

func LogInfo(format string, args ...any) {
	Log(LogLevelInfo, format, args...)
}

func LogWarn(format string, args ...any) {
	Log(LogLevelWarn, format, args...)
}

func LogError(format string, args ...any) {
	Log(LogLevelError, format, args...)
}

func GetRecentLogs() []LogEntry {
	globalLogBuffer.mu.RLock()
	defer globalLogBuffer.mu.RUnlock()
	res := make([]LogEntry, len(globalLogBuffer.entries))
	copy(res, globalLogBuffer.entries)
	return res
}

func ClearLogs() {
	globalLogBuffer.mu.Lock()
	globalLogBuffer.entries = globalLogBuffer.entries[:0]
	globalLogBuffer.mu.Unlock()
}
