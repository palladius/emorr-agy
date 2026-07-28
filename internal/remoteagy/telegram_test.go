package remoteagy

import (
	"strings"
	"testing"
)

// --- Keyboard Tests ---

func TestBuildConciergeKeyboard_NewFolders(t *testing.T) {
	trueVal := true
	cfg := &Config{
		DefaultSandbox: &trueVal,
		Folders: []FolderConfig{
			{Name: "GIC", Path: "/tmp/gic"},
			{Name: "Test", Path: "/tmp/test"},
		},
	}

	result, err := BuildConciergeKeyboard(cfg, make(map[string]*RemoteSession))
	if err != nil {
		t.Fatalf("BuildConciergeKeyboard error: %v", err)
	}

	if !strings.Contains(result, "ragy:new:GIC") {
		t.Error("expected 'ragy:new:GIC' callback data")
	}
	if !strings.Contains(result, "🚀 GIC") {
		t.Error("expected rocket emoji + folder name")
	}
}

func TestBuildConciergeKeyboard_WithActiveSessions(t *testing.T) {
	cfg := &Config{
		Folders: []FolderConfig{
			{Name: "GIC", Path: "/tmp/gic"},
			{Name: "Test", Path: "/tmp/test"},
		},
	}

	sessions := map[string]*RemoteSession{
		"remote-agy-gic": {Name: "remote-agy-gic", Folder: cfg.Folders[0]},
	}

	result, err := BuildConciergeKeyboard(cfg, sessions)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// GIC should show as "Resume" since it has an active session
	if !strings.Contains(result, "ragy:attach:remote-agy-gic") {
		t.Error("expected attach callback for active GIC session")
	}
	if !strings.Contains(result, "Resume: GIC") {
		t.Error("expected Resume label for active session")
	}

	// Test should still show as "New"
	if !strings.Contains(result, "ragy:new:Test") {
		t.Error("expected 'ragy:new:Test' for inactive folder")
	}
}

func TestBuildAttachedKeyboard(t *testing.T) {
	result, err := BuildAttachedKeyboard("remote-agy-gic")
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// Check all control buttons exist
	expectedCallbacks := []string{
		"ragy:refresh:remote-agy-gic",
		"ragy:pause:remote-agy-gic",
		"ragy:info:remote-agy-gic",
		"ragy:menu",
		"ragy:kill:remote-agy-gic",
	}

	for _, cb := range expectedCallbacks {
		if !strings.Contains(result, cb) {
			t.Errorf("expected callback %q in keyboard", cb)
		}
	}
}

// --- ask_question Detection Tests ---

func TestDetectAskQuestion_NumberedOptions(t *testing.T) {
	paneContent := `Some context here

What would you like to do?

1) Fix the bug
2) Add a feature
3) Refactor the code
`

	result := DetectAskQuestion(paneContent)
	if result == nil {
		t.Fatal("expected ask_question to be detected")
	}

	if len(result.Choices) != 3 {
		t.Fatalf("expected 3 choices, got %d", len(result.Choices))
	}

	if result.Choices[0].Text != "Fix the bug" {
		t.Errorf("expected first choice 'Fix the bug', got %q", result.Choices[0].Text)
	}
}

func TestDetectAskQuestion_DotNotation(t *testing.T) {
	paneContent := `
1. Yes, proceed
2. No, stop
`

	result := DetectAskQuestion(paneContent)
	if result == nil {
		t.Fatal("expected ask_question to be detected")
	}

	if len(result.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(result.Choices))
	}
}

func TestDetectAskQuestion_NoOptions(t *testing.T) {
	paneContent := "Just some normal output\nNothing special here"

	result := DetectAskQuestion(paneContent)
	if result != nil {
		t.Errorf("expected nil for no options, got %+v", result)
	}
}

func TestDetectAskQuestion_SingleOption(t *testing.T) {
	paneContent := "1) Only one option"

	result := DetectAskQuestion(paneContent)
	if result != nil {
		t.Error("expected nil for single option (need at least 2)")
	}
}

func TestBuildAskQuestionKeyboard(t *testing.T) {
	aq := &DetectedAskQuestion{
		Question: "What to do?",
		Choices: []AskQuestionChoice{
			{Index: 1, Text: "Option A"},
			{Index: 2, Text: "Option B"},
		},
	}

	result, err := BuildAskQuestionKeyboard("remote-agy-gic", aq)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(result, "ragy:answer:remote-agy-gic:1") {
		t.Error("expected answer callback for option 1")
	}
	if !strings.Contains(result, "ragy:answer:remote-agy-gic:2") {
		t.Error("expected answer callback for option 2")
	}
	if !strings.Contains(result, "ragy:writein:remote-agy-gic") {
		t.Error("expected write-in button")
	}
}

// --- Number Parsing Tests ---

func TestIsNumberedOption(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"1) Fix bug", true},
		{"2. Add feature", true},
		{"3: Refactor", true},
		{"12) Long number", true},
		{"Not a number", false},
		{"1", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isNumberedOption(tt.line)
		if result != tt.expected {
			t.Errorf("isNumberedOption(%q) = %v, want %v", tt.line, result, tt.expected)
		}
	}
}

func TestParseNumberedOption(t *testing.T) {
	tests := []struct {
		line         string
		expectedIdx  int
		expectedText string
	}{
		{"1) Fix bug", 1, "Fix bug"},
		{"2. Add feature", 2, "Add feature"},
		{"3: Refactor", 3, "Refactor"},
		{"12) Long number option", 12, "Long number option"},
	}

	for _, tt := range tests {
		idx, text := parseNumberedOption(tt.line)
		if idx != tt.expectedIdx || text != tt.expectedText {
			t.Errorf("parseNumberedOption(%q) = (%d, %q), want (%d, %q)",
				tt.line, idx, text, tt.expectedIdx, tt.expectedText)
		}
	}
}
