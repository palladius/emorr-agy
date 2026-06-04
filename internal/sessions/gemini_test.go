package sessions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGeminiClassifier(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gemini-test-db-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbDir := filepath.Join(tempDir, ".gemini/antigravity-cli/conversations")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		t.Fatalf("failed to create db dir: %v", err)
	}

	sessionID := "test-session-123"
	dbPath := filepath.Join(dbDir, sessionID+".db")

	// Create a dummy SQLite database with steps table and a step
	cmd := exec.Command("sqlite3", dbPath, `
		CREATE TABLE steps (
			idx integer primary key,
			step_type integer,
			status integer,
			has_subtrajectory numeric,
			step_payload blob
		);
		INSERT INTO steps (idx, step_type, status, has_subtrajectory, step_payload)
		VALUES (1, 15, 3, 0, 'Test payload text contents.');
	`)
	err = cmd.Run()
	if err != nil {
		t.Logf("Warning: sqlite3 not running or failed: %v", err)
		// We'll handle tests gracefully if sqlite3 is missing
	}

	// 1. Mock HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"candidates": [
				{
					"content": {
						"parts": [
							{
								"text": "{\"about\": \"mock testing\", \"user_input_pending\": false, \"worth_resuscitate\": true}"
							}
						]
					}
				}
			]
		}`)
	}))
	defer server.Close()

	// Override API URL
	originalURL := geminiAPIURL
	defer func() { geminiAPIURL = originalURL }()
	geminiAPIURL = server.URL + "/?key=%s"

	t.Run("Valid classification", func(t *testing.T) {
		classifier := NewGeminiClassifier("dummy-key", tempDir)
		res, err := classifier.Classify(sessionID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.About != "mock testing" || res.WorthResuscitate != true || res.UserInputPending != false {
			t.Errorf("unexpected result: %+v", res)
		}
	})

	t.Run("Empty API key error", func(t *testing.T) {
		classifier := NewGeminiClassifier("", tempDir)
		_, err := classifier.Classify(sessionID)
		if err == nil {
			t.Fatal("expected error due to empty API key, got nil")
		}
	})

	t.Run("Non-existent session database returns empty log result", func(t *testing.T) {
		classifier := NewGeminiClassifier("dummy-key", tempDir)
		res, err := classifier.Classify("non-existent-session")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.About != "No logs available in database" {
			t.Errorf("expected empty log summary, got %q", res.About)
		}
	})

	t.Run("Dynamic API URL with environment variables", func(t *testing.T) {
		// Restore default temporarily to test dynamic URL construction
		currentURL := geminiAPIURL
		geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s"
		defer func() { geminiAPIURL = currentURL }()

		classifier := NewGeminiClassifier("dummy-key", tempDir)

		// 1. Without env override (defaulting to gemini-3.1-flash-lite)
		t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "")
		t.Setenv("GEMINI_MODEL", "")
		url := classifier.getAPIURL()
		expected := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite:generateContent?key=dummy-key"
		if url != expected {
			t.Errorf("expected %q, got %q", expected, url)
		}

		// 2. With env override
		t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "gemini-override-2")
		url = classifier.getAPIURL()
		expected = "https://generativelanguage.googleapis.com/v1beta/models/gemini-override-2:generateContent?key=dummy-key"
		if url != expected {
			t.Errorf("expected %q, got %q", expected, url)
		}
	})
}

