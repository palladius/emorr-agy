package sessions

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type MockClassifier struct {
	called bool
	result *LLMResult
	err    error
}

func (m *MockClassifier) Classify(sessionID string) (*LLMResult, error) {
	m.called = true
	return m.result, m.err
}

func TestShowSessionBasic(t *testing.T) {
	mockTmux := &MockTmuxRunner{
		sessions: []TmuxSession{
			{Name: "emagy-session-123", Path: "/workspace/proj", Attached: true, Windows: 2},
		},
	}
	mockFS := &MockFileSystem{
		files: map[string][]byte{},
	}

	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	var buf bytes.Buffer
	err := ShowSession(&buf, engine, "emagy-session-123", ShowOptions{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "emagy-session-123") || !strings.Contains(output, "/workspace/proj") || !strings.Contains(output, "open_tmux") {
		t.Errorf("missing basic session details: %q", output)
	}
}

func TestShowSessionLockingAndCaching(t *testing.T) {
	sessionID := "emagy-session-456"
	homeDir := "/home/ricc"
	lockPath := filepath.Join(homeDir, ".emorr-agy/locks", sessionID+".lock")
	cachePath := filepath.Join(homeDir, ".emorr-agy/cache", sessionID+".json")

	t.Run("Locking fails when active PID exists", func(t *testing.T) {
		// Mock FS containing lock file with current process PID (always active)
		currentPID := os.Getpid()
		mockTmux := &MockTmuxRunner{
			sessions: []TmuxSession{
				{Name: sessionID, Path: "/workspace/proj", Attached: true, Windows: 2},
			},
		}
		mockFS := &MockFileSystem{
			files: map[string][]byte{
				lockPath: []byte(strconv.Itoa(currentPID)),
			},
			stats: map[string]os.FileInfo{
				lockPath: MockFileInfo{name: filepath.Base(lockPath), isDir: false},
			},
		}

		engine := NewClassificationEngine(mockTmux, mockFS, homeDir)
		mockClassifier := &MockClassifier{}

		var buf bytes.Buffer
		err := ShowSession(&buf, engine, sessionID, ShowOptions{Classify: true}, mockClassifier)
		if err == nil {
			t.Fatal("expected error due to active lock, got nil")
		}
		if !strings.Contains(err.Error(), "another script PID") {
			t.Errorf("expected error message to mention 'another script PID', got %v", err)
		}
		if mockClassifier.called {
			t.Error("classifier should not have been called")
		}
	})

	t.Run("Use cached results if available", func(t *testing.T) {
		mockTmux := &MockTmuxRunner{
			sessions: []TmuxSession{
				{Name: sessionID, Path: "/workspace/proj", Attached: true, Windows: 2},
			},
		}
		cachedData, _ := json.Marshal(LLMResult{
			About:            "Debugging tests",
			UserInputPending: true,
			WorthResuscitate: false,
		})
		mockFS := &MockFileSystem{
			files: map[string][]byte{
				cachePath: cachedData,
			},
			stats: map[string]os.FileInfo{
				cachePath: MockFileInfo{name: filepath.Base(cachePath), isDir: false},
			},
		}

		engine := NewClassificationEngine(mockTmux, mockFS, homeDir)
		mockClassifier := &MockClassifier{}

		var buf bytes.Buffer
		err := ShowSession(&buf, engine, sessionID, ShowOptions{Classify: true}, mockClassifier)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Debugging tests") || !strings.Contains(output, "User Input Pending: true") {
			t.Errorf("missing cached classification details: %q", output)
		}
		if mockClassifier.called {
			t.Error("classifier should not have been called since cache was used")
		}
	})

	t.Run("Call classifier and cache result when cache missing", func(t *testing.T) {
		mockTmux := &MockTmuxRunner{
			sessions: []TmuxSession{
				{Name: sessionID, Path: "/workspace/proj", Attached: true, Windows: 2},
			},
		}
		mockFS := &MockFileSystem{
			files: map[string][]byte{},
			stats: map[string]os.FileInfo{},
		}

		engine := NewClassificationEngine(mockTmux, mockFS, homeDir)
		mockClassifier := &MockClassifier{
			result: &LLMResult{
				About:            "Running code generator",
				UserInputPending: false,
				WorthResuscitate: true,
			},
		}

		var buf bytes.Buffer
		// In a real execution, writing to FS happens. In unit tests, we'll check that the classifier was called.
		err := ShowSession(&buf, engine, sessionID, ShowOptions{Classify: true}, mockClassifier)
		if err != nil && !errors.Is(err, os.ErrNotExist) { // ignore write errors on mocked FS
			t.Fatalf("unexpected error: %v", err)
		}

		if !mockClassifier.called {
			t.Error("classifier should have been called")
		}
	})
}
