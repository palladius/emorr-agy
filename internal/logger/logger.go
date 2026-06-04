package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type LogLevel string

const (
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARNING"
	LevelError   LogLevel = "ERROR"
)

var (
	mu       sync.Mutex
	logFile  *os.File
	projID   string
	localDir = "log"
)

// Init initializes the logger package, opening/creating the local log file in the specified directory.
func Init(projectID string) error {
	mu.Lock()
	defer mu.Unlock()

	projID = projectID

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(localDir, "server.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	if logFile != nil {
		_ = logFile.Close()
	}
	logFile = file
	return nil
}

// Close closes the local log file descriptor.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// Logf formats and writes a log entry with the given level.
func Logf(level LogLevel, format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)

	// Determine caller context
	var caller string
	if _, file, line, ok := runtime.Caller(2); ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n", now, level, caller, msg)

	// Write to stderr/stdout
	if level == LevelError {
		_, _ = os.Stderr.WriteString(logLine)
	} else {
		_, _ = os.Stdout.WriteString(logLine)
	}

	// Write to local file
	if logFile != nil {
		_, _ = logFile.WriteString(logLine)
	}

	// Queue for GCP Cloud Logging if project is configured
	if projID != "" {
		queueGCPLog(level, msg, caller)
	}
}

// Infof writes an info log entry.
func Infof(format string, args ...interface{}) {
	Logf(LevelInfo, format, args...)
}

// Warningf writes a warning log entry.
func Warningf(format string, args ...interface{}) {
	Logf(LevelWarning, format, args...)
}

// Errorf writes an error log entry.
func Errorf(format string, args ...interface{}) {
	Logf(LevelError, format, args...)
}
