package remoteagy

import (
	"testing"
	"time"
)

func TestDiffContent_NoPrevious(t *testing.T) {
	result := diffContent("", "Hello\nWorld")
	if result != "Hello\nWorld" {
		t.Errorf("expected 'Hello\\nWorld', got %q", result)
	}
}

func TestDiffContent_SameContent(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3"
	result := diffContent(content, content)
	if result != "" {
		t.Errorf("expected empty string for same content, got %q", result)
	}
}

func TestDiffContent_NewLinesAppended(t *testing.T) {
	prev := "Line 1\nLine 2"
	curr := "Line 1\nLine 2\nLine 3\nLine 4"
	result := diffContent(prev, curr)
	if result != "Line 3\nLine 4" {
		t.Errorf("expected 'Line 3\\nLine 4', got %q", result)
	}
}

func TestDiffContent_CompletelyDifferent(t *testing.T) {
	prev := "Old content"
	curr := "Totally new content"
	result := diffContent(prev, curr)
	if result != "Totally new content" {
		t.Errorf("expected 'Totally new content', got %q", result)
	}
}

func TestDiffContent_ScrolledOutput(t *testing.T) {
	// Simulate tmux scroll: old lines disappear, new lines appear
	prev := "Line 1\nLine 2\nLine 3"
	curr := "Line 2\nLine 3\nLine 4\nLine 5"
	result := diffContent(prev, curr)
	if result != "Line 4\nLine 5" {
		t.Errorf("expected 'Line 4\\nLine 5', got %q", result)
	}
}

func TestNewPoller(t *testing.T) {
	outputReceived := false
	p := NewPoller("test-session", 100*time.Millisecond, func(s string) {
		outputReceived = true
	}, nil)

	if p.sessionName != "test-session" {
		t.Errorf("expected session name 'test-session', got %q", p.sessionName)
	}
	if p.running {
		t.Error("poller should not be running before Start()")
	}
	_ = outputReceived // suppress unused
}

func TestPollerPauseResume(t *testing.T) {
	p := NewPoller("test", 100*time.Millisecond, nil, nil)

	if p.IsPaused() {
		t.Error("poller should not be paused initially")
	}

	p.Pause()
	if !p.IsPaused() {
		t.Error("poller should be paused after Pause()")
	}

	p.Resume()
	if p.IsPaused() {
		t.Error("poller should not be paused after Resume()")
	}
}

func TestSessionInfo(t *testing.T) {
	trueVal := true
	session := &RemoteSession{
		Name: "remote-agy-gic",
		Folder: FolderConfig{
			Name:     "GIC",
			Path:     "/home/ricc/git/gic",
			AgyFlags: "",
			Sandbox:  &trueVal,
		},
		Marker:    "emorr-agy-marker-abc123",
		State:     SessionRunning,
		StartedAt: time.Now().Add(-5 * time.Minute),
	}

	info := SessionInfo(session)
	if info == "" {
		t.Error("expected non-empty session info")
	}
	if !contains(info, "remote-agy-gic") {
		t.Error("session info should contain session name")
	}
	if !contains(info, "Sandboxed") {
		t.Error("session info should show sandbox status")
	}
}

func TestSessionInfo_Unsandboxed(t *testing.T) {
	falseVal := false
	session := &RemoteSession{
		Name: "remote-agy-unsafe",
		Folder: FolderConfig{
			Name:    "Unsafe",
			Path:    "/tmp/test",
			Sandbox: &falseVal,
		},
		StartedAt: time.Now(),
	}

	info := SessionInfo(session)
	if !contains(info, "UNSANDBOXED") {
		t.Error("session info should show UNSANDBOXED for sandbox=false")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
