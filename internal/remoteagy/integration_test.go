package remoteagy

import (
	"os"
	"path/filepath"
	"testing"
)

func ptrBool(b bool) *bool {
	return &b
}

// TestIntegrationConfigLoad tests the full config loading pipeline.
func TestIntegrationConfigLoad(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test-folders.yaml")

	yamlContent := `
default_sandbox: true
folders:
  - name: "Test A"
    path: "~/git/test-a"
    sandbox: true
  - name: "Test B"
    path: "~/git/test-b"
    agy_flags: "--dangerously-skip-permissions"
    sandbox: false
  - name: "🧪 Emoji Folder"
    path: "/tmp/emoji-test"
    sandbox: true
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EMORR_AGY_FOLDERS_YAML", yamlPath)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if len(cfg.Folders) != 3 {
		t.Fatalf("Expected 3 folders, got %d", len(cfg.Folders))
	}

	if cfg.DefaultSandbox == nil || !*cfg.DefaultSandbox {
		t.Error("Expected default_sandbox to be true")
	}

	// Folder A: sandbox=true from config
	if !cfg.Folders[0].IsSandboxed(cfg.DefaultSandbox) {
		t.Error("Folder A should be sandboxed")
	}

	// Folder B: sandbox=false, overrides default
	if cfg.Folders[1].IsSandboxed(cfg.DefaultSandbox) {
		t.Error("Folder B should NOT be sandboxed (explicitly false)")
	}

	// Folder B: has custom flags
	if cfg.Folders[1].AgyFlags != "--dangerously-skip-permissions" {
		t.Errorf("Folder B flags = %q, want --dangerously-skip-permissions", cfg.Folders[1].AgyFlags)
	}

	// Emoji folder
	if cfg.Folders[2].Name != "🧪 Emoji Folder" {
		t.Errorf("Folder C name = %q, want '🧪 Emoji Folder'", cfg.Folders[2].Name)
	}
}

// TestIntegrationConciergeLifecycle tests the full concierge state machine.
func TestIntegrationConciergeLifecycle(t *testing.T) {
	cfg := &Config{
		DefaultSandbox: ptrBool(true),
		Folders: []FolderConfig{
			{Name: "TestA", Path: "/tmp/a", Sandbox: ptrBool(true)},
			{Name: "TestB", Path: "/tmp/b", Sandbox: ptrBool(false)},
		},
	}

	c := NewConcierge(cfg)

	// Initial state
	mode, attached := c.GetState()
	if mode != ModeConcierge {
		t.Errorf("Initial mode = %v, want ModeConcierge", mode)
	}
	if attached != "" {
		t.Errorf("Initial attached = %q, want empty", attached)
	}
	if c.ActiveSessionCount() != 0 {
		t.Errorf("Initial session count = %d, want 0", c.ActiveSessionCount())
	}

	// Add a session
	c.Sessions["test-session-1"] = &RemoteSession{
		Name:   "test-session-1",
		Folder: cfg.Folders[0],
		State:  SessionRunning,
	}

	// Session count after add
	if c.ActiveSessionCount() != 1 {
		t.Errorf("Session count after add = %d, want 1", c.ActiveSessionCount())
	}

	// Attach to non-existent session (not in tmux) should fail
	err := c.AttachSession("nonexistent")
	if err == nil {
		t.Error("Expected error attaching to nonexistent session")
	}

	// Directly test detach cycle (without tmux)
	c.mu.Lock()
	c.Mode = ModeAttached
	c.AttachedSession = "test-session-1"
	c.mu.Unlock()

	mode, attached = c.GetState()
	if mode != ModeAttached {
		t.Errorf("After manual attach, mode = %v, want ModeAttached", mode)
	}
	if attached != "test-session-1" {
		t.Errorf("After manual attach, attached = %q, want test-session-1", attached)
	}

	c.DetachSession()
	mode, attached = c.GetState()
	if mode != ModeConcierge {
		t.Errorf("After detach, mode = %v, want ModeConcierge", mode)
	}
	if attached != "" {
		t.Errorf("After detach, attached = %q, want empty", attached)
	}
}

// TestIntegrationKeyboardBuilding tests the full keyboard building pipeline.
func TestIntegrationKeyboardBuilding(t *testing.T) {
	cfg := &Config{
		DefaultSandbox: ptrBool(true),
		Folders: []FolderConfig{
			{Name: "TestA", Path: "/tmp/a", Sandbox: ptrBool(true)},
			{Name: "TestB", Path: "/tmp/b", Sandbox: ptrBool(false)},
		},
	}

	sessions := map[string]*RemoteSession{
		"remote-agy-testa": {
			Name:   "remote-agy-testa",
			Folder: cfg.Folders[0],
			State:  SessionRunning,
		},
	}

	// Build concierge keyboard
	markup, err := BuildConciergeKeyboard(cfg, sessions)
	if err != nil {
		t.Fatalf("BuildConciergeKeyboard error: %v", err)
	}
	if markup == "" {
		t.Fatal("Expected non-empty concierge markup")
	}

	// Build attached keyboard
	aMarkup, err := BuildAttachedKeyboard("remote-agy-testa")
	if err != nil {
		t.Fatalf("BuildAttachedKeyboard error: %v", err)
	}
	if aMarkup == "" {
		t.Fatal("Expected non-empty attached markup")
	}

	// Build ask_question keyboard
	aq := &DetectedAskQuestion{
		Question: "What color?",
		Choices: []AskQuestionChoice{
			{Index: 1, Text: "Red"},
			{Index: 2, Text: "Blue"},
			{Index: 3, Text: "Green"},
		},
	}
	aqMarkup, err := BuildAskQuestionKeyboard("remote-agy-testa", aq)
	if err != nil {
		t.Fatalf("BuildAskQuestionKeyboard error: %v", err)
	}
	if aqMarkup == "" {
		t.Fatal("Expected non-empty ask_question markup")
	}
}

// TestIntegrationSanitizeSessionName tests name sanitization for tmux.
func TestIntegrationSanitizeSessionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GIC", "remote-agy-gic"},
		{"🔓 GIC Unsafe", "remote-agy-gic-unsafe"},
		{"🏖️ Casalborsetti YOLO", "remote-agy-casalborsetti-yolo"},
		{"🧪 Test Repo", "remote-agy-test-repo"},
	}

	for _, tt := range tests {
		result := SanitizeSessionName(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeSessionName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestIntegrationSessionInfo tests session info formatting.
func TestIntegrationSessionInfo(t *testing.T) {
	session := &RemoteSession{
		Name:   "remote-agy-test",
		Folder: FolderConfig{Name: "Test", Path: "/tmp/test", Sandbox: ptrBool(true)},
		State:  SessionRunning,
	}

	info := SessionInfo(session)
	if info == "" {
		t.Error("Expected non-empty session info")
	}
}

// TestIntegrationPollerCreation tests poller creation without starting it.
func TestIntegrationPollerCreation(t *testing.T) {
	onOutput := func(s string) {}
	onEnd := func() {}

	p := NewPoller("test-session", 1000, onOutput, onEnd)
	if p == nil {
		t.Fatal("Expected non-nil poller")
	}

	// Test pause/resume without starting
	if p.IsPaused() {
		t.Error("New poller should not be paused")
	}
	p.Pause()
	if !p.IsPaused() {
		t.Error("Poller should be paused after Pause()")
	}
	p.Resume()
	if p.IsPaused() {
		t.Error("Poller should not be paused after Resume()")
	}
}
