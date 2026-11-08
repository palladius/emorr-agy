package sessions

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCrop(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"line1\nline2\r\nline3", 50, "line1 line2 line3"},
		{"untruncated", 0, "untruncated"},
	}

	for _, tc := range tests {
		got := crop(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("crop(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{45 * time.Second, "45s"},
		{75 * time.Second, "1m 15s"},
		{65 * time.Minute, "1h 5m"},
	}

	for _, tc := range tests {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestShowTimeline(t *testing.T) {
	tempDir := t.TempDir()
	transcriptPath := filepath.Join(tempDir, "transcript.jsonl")

	transcriptContent := `{"step_index":0,"type":"USER_INPUT","created_at":"2026-11-08T09:00:00Z","content":"<USER_REQUEST>Fix bug in carchess3d</USER_REQUEST><ADDITIONAL_METADATA>meta</ADDITIONAL_METADATA>"}
{"step_index":1,"type":"PLANNER_RESPONSE","created_at":"2026-11-08T09:01:00Z","content":"I will inspect the codebase","tool_calls":[{"name":"run_command","args":{"CommandLine":"git status","toolAction":"check git status"}}]}
{"step_index":2,"type":"PLANNER_RESPONSE","created_at":"2026-11-08T09:02:00Z","content":"","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/ricc/git/carchess3d/main.js"}}]}
{"step_index":3,"type":"PLANNER_RESPONSE","created_at":"2026-11-08T09:03:00Z","content":"","tool_calls":[{"name":"search_repo","args":{"query":"chess"}}]}
`
	if err := os.WriteFile(transcriptPath, []byte(transcriptContent), 0644); err != nil {
		t.Fatalf("failed to write test transcript: %v", err)
	}

	var buf bytes.Buffer
	err := ShowTimeline(&buf, transcriptPath, "test-uuid-12345", tempDir)
	if err != nil {
		t.Fatalf("ShowTimeline failed: %v", err)
	}

	output := buf.String()

	// Verify date separator
	if !strings.Contains(output, "2026-11-08") {
		t.Errorf("expected date separator 2026-11-08 in output, got:\n%s", output)
	}

	// Verify cleaned user content
	if !strings.Contains(output, "Fix bug in carchess3d") {
		t.Errorf("expected prompt in output, got:\n%s", output)
	}
	if strings.Contains(output, "<USER_REQUEST>") {
		t.Errorf("expected <USER_REQUEST> to be cleaned, got:\n%s", output)
	}

	// Verify run_command rendering
	if !strings.Contains(output, "git status # check git status") {
		t.Errorf("expected run_command display in output, got:\n%s", output)
	}

	// Verify view_file rendering
	if !strings.Contains(output, "view_file") || !strings.Contains(output, "carchess3d/main.js") {
		t.Errorf("expected view_file in output, got:\n%s", output)
	}

	// Verify stats
	if !strings.Contains(output, "Stats della Sessione") || !strings.Contains(output, "Steps:     4") {
		t.Errorf("expected stats in output, got:\n%s", output)
	}

	// Verify resume hint
	if !strings.Contains(output, "cd ~/git/carchess3d && agy --conversation test-uuid-12345") {
		t.Errorf("expected resume hint in output, got:\n%s", output)
	}
}

func TestFindTranscriptPath(t *testing.T) {
	tempHome := t.TempDir()
	convUUID := "12345678-abcd-ef01-2345-6789abcdef01"
	logDir := filepath.Join(tempHome, ".gemini/antigravity-cli/brain", convUUID, ".system_generated/logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}
	tFile := filepath.Join(logDir, "transcript.jsonl")
	if err := os.WriteFile(tFile, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write dummy transcript: %v", err)
	}

	// Exact match
	path, uuid, err := FindTranscriptPath(tempHome, convUUID)
	if err != nil || path != tFile || uuid != convUUID {
		t.Errorf("FindTranscriptPath exact match failed: got (%q, %q, %v), want (%q, %q, nil)", path, uuid, err, tFile, convUUID)
	}

	// Prefix match
	path, uuid, err = FindTranscriptPath(tempHome, "12345678")
	if err != nil || path != tFile || uuid != convUUID {
		t.Errorf("FindTranscriptPath prefix match failed: got (%q, %q, %v), want (%q, %q, nil)", path, uuid, err, tFile, convUUID)
	}

	// Non-existent match
	_, _, err = FindTranscriptPath(tempHome, "non-existent-uuid")
	if err == nil {
		t.Error("expected error for non-existent session, got nil")
	}
}
