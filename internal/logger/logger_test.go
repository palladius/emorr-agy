package logger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalLogger(t *testing.T) {
	// Redirect logging directory to temporary directory
	tempDir := t.TempDir()
	localDir = tempDir

	err := Init("")
	if err != nil {
		t.Fatalf("logger Init failed: %v", err)
	}
	defer Close()

	Infof("This is an info message: %s", "hello")
	Warningf("This is a warning: %d", 42)
	Errorf("This is an error")

	logPath := filepath.Join(tempDir, "server.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[INFO]") || !strings.Contains(content, "This is an info message: hello") {
		t.Errorf("missing info log details in: %q", content)
	}
	if !strings.Contains(content, "[WARNING]") || !strings.Contains(content, "This is a warning: 42") {
		t.Errorf("missing warning log details in: %q", content)
	}
	if !strings.Contains(content, "[ERROR]") || !strings.Contains(content, "This is an error") {
		t.Errorf("missing error log details in: %q", content)
	}
}

func TestGCPLoggerBatchSend(t *testing.T) {
	// Mock auth token cache
	cachedToken = "mock-bearer-token"
	tokenExpiry = time.Now().Add(10 * time.Minute)

	var lastReq gcpLogRequest
	var lastAuth string
	var reqChan = make(chan bool, 1)

	gcpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastAuth = r.Header.Get("Authorization")
		if r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&lastReq)
			reqChan <- true
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer gcpServer.Close()

	// Redirect endpoints and parameters
	tempDir := t.TempDir()
	localDir = tempDir
	
	// Override the logging API endpoint URL
	originalURL := gcpLoggingURL
	gcpLoggingURL = gcpServer.URL + "/v2/entries:write"
	defer func() {
		gcpLoggingURL = originalURL
	}()

	err := Init("test-project-123")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer Close()

	Infof("GCP Test Log Msg")

	// Wait for the background worker to dispatch to mock server
	select {
	case <-reqChan:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for GCP logging dispatcher to send batch payload")
	}

	if lastAuth != "Bearer mock-bearer-token" {
		t.Errorf("expected Authorization header 'Bearer mock-bearer-token', got %q", lastAuth)
	}
	if len(lastReq.Entries) != 1 {
		t.Fatalf("expected 1 log entry in batch request, got %d", len(lastReq.Entries))
	}
	entry := lastReq.Entries[0]
	if entry.Severity != "INFO" {
		t.Errorf("expected severity INFO, got %q", entry.Severity)
	}
	if entry.JSONPayload.Message != "GCP Test Log Msg" {
		t.Errorf("expected message 'GCP Test Log Msg', got %q", entry.JSONPayload.Message)
	}
	if !strings.Contains(entry.LogName, "projects/test-project-123/logs/emorr-agy-server") {
		t.Errorf("expected LogName to contain project ID and log identifier, got %q", entry.LogName)
	}
}
