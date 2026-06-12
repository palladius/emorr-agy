package sessions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDetailedStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "transcripts-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test transcript directories
	sessionID := "test-session-123"
	logsDir := filepath.Join(tempDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create logs dir: %v", err)
	}

	transcriptPath := filepath.Join(logsDir, "transcript.jsonl")

	t.Run("Waiting on User Question", func(t *testing.T) {
		content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","content":"Go!"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","tool_calls":[{"name":"ask_question","args":{"questions":[{"question":"What is your color preference?"}]}}]}`
		err = os.WriteFile(transcriptPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write transcript: %v", err)
		}

		status, emoji, detail := GetDetailedStatus(tempDir, sessionID)
		if status != "Waiting (Question)" {
			t.Errorf("expected status 'Waiting (Question)', got %q", status)
		}
		if emoji != "💬" {
			t.Errorf("expected emoji '💬', got %q", emoji)
		}
		if detail != "What is your color preference?" {
			t.Errorf("expected detail 'What is your color preference?', got %q", detail)
		}
	})

	t.Run("Waiting on Permission", func(t *testing.T) {
		content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","content":"Go!"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","tool_calls":[{"name":"ask_permission","args":{"Action":"command","Target":"rm -rf /tmp"}}]}`
		err = os.WriteFile(transcriptPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write transcript: %v", err)
		}

		status, emoji, detail := GetDetailedStatus(tempDir, sessionID)
		if status != "Waiting (Permission)" {
			t.Errorf("expected status 'Waiting (Permission)', got %q", status)
		}
		if emoji != "🔒" {
			t.Errorf("expected emoji '🔒', got %q", emoji)
		}
		if detail != "Approve command: rm -rf /tmp" {
			t.Errorf("expected detail 'Approve command: rm -rf /tmp', got %q", detail)
		}
	})

	t.Run("Running Command", func(t *testing.T) {
		content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","content":"Go!"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","tool_calls":[{"name":"run_command","args":{"CommandLine":"ls -l"}}]}`
		err = os.WriteFile(transcriptPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write transcript: %v", err)
		}

		status, emoji, detail := GetDetailedStatus(tempDir, sessionID)
		if status != "Running Command" {
			t.Errorf("expected status 'Running Command', got %q", status)
		}
		if emoji != "⚙️" {
			t.Errorf("expected emoji '⚙️', got %q", emoji)
		}
		if detail != "ls -l" {
			t.Errorf("expected detail 'ls -l', got %q", detail)
		}
	})
}
