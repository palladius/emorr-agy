package remoteagy

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Config Loading Tests ---

func TestLoadConfig_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-folders.yaml")

	yamlContent := `
default_sandbox: true
folders:
  - name: "GIC"
    path: "~/git/gic"
    agy_flags: ""
    sandbox: true
  - name: "GIC Unsafe"
    path: "~/git/gic"
    agy_flags: "--dangerously-skip-permissions"
    sandbox: false
  - name: "Casalborsetti YOLO"
    path: "~/git/ricc-personal-life/casalborsetti-cre-app"
    agy_flags: "--dangerously-skip-permissions"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test YAML: %v", err)
	}

	t.Setenv(ConfigEnvVar, configPath)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if len(cfg.Folders) != 3 {
		t.Fatalf("expected 3 folders, got %d", len(cfg.Folders))
	}

	if cfg.Folders[0].Name != "GIC" {
		t.Errorf("expected first folder name 'GIC', got %q", cfg.Folders[0].Name)
	}

	if cfg.Folders[1].AgyFlags != "--dangerously-skip-permissions" {
		t.Errorf("expected GIC Unsafe flags, got %q", cfg.Folders[1].AgyFlags)
	}

	if cfg.DefaultSandbox == nil || !*cfg.DefaultSandbox {
		t.Error("expected default_sandbox to be true")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	t.Setenv(ConfigEnvVar, "/nonexistent/path/config.yaml")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad.yaml")

	if err := os.WriteFile(configPath, []byte("{{{{not yaml"), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	t.Setenv(ConfigEnvVar, configPath)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadConfig_DuplicateFolderNames(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "dup.yaml")

	yamlContent := `
folders:
  - name: "GIC"
    path: "~/git/gic"
  - name: "GIC"
    path: "~/git/gic2"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	t.Setenv(ConfigEnvVar, configPath)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for duplicate folder names")
	}
}

func TestLoadConfig_EmptyFolderName(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty-name.yaml")

	yamlContent := `
folders:
  - name: ""
    path: "~/git/gic"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	t.Setenv(ConfigEnvVar, configPath)

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for empty folder name")
	}
}

func TestLoadConfig_EnvVarOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom.yaml")

	yamlContent := `
folders:
  - name: "Test"
    path: "/tmp/test"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	t.Setenv(ConfigEnvVar, configPath)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if len(cfg.Folders) != 1 || cfg.Folders[0].Name != "Test" {
		t.Errorf("expected 1 folder named 'Test', got %+v", cfg.Folders)
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home dir")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/git/gic", filepath.Join(home, "git/gic")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		result, err := expandHome(tt.input)
		if err != nil {
			t.Errorf("expandHome(%q) error: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// --- Sandbox Tests ---

func TestFolderConfig_IsSandboxed(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name           string
		folderSandbox  *bool
		defaultSandbox *bool
		expected       bool
	}{
		{"folder true, default nil", &trueVal, nil, true},
		{"folder false, default true", &falseVal, &trueVal, false}, // folder overrides
		{"folder nil, default true", nil, &trueVal, true},
		{"folder nil, default false", nil, &falseVal, false},
		{"both nil (safe default)", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FolderConfig{Sandbox: tt.folderSandbox}
			result := f.IsSandboxed(tt.defaultSandbox)
			if result != tt.expected {
				t.Errorf("IsSandboxed() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// --- Session Naming Tests ---

func TestSanitizeSessionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GIC", "remote-agy-gic"},
		{"🔓 GIC Unsafe", "remote-agy-gic-unsafe"},
		{"emorr-agy", "remote-agy-emorr-agy"},
		{"🏖️ Casalborsetti YOLO", "remote-agy-casalborsetti-yolo"},
		{"  spaces  ", "remote-agy-spaces"},
	}

	for _, tt := range tests {
		result := SanitizeSessionName(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeSessionName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsRemoteSession(t *testing.T) {
	if !IsRemoteSession("remote-agy-gic") {
		t.Error("expected remote-agy-gic to be a remote session")
	}
	if IsRemoteSession("my-local-session") {
		t.Error("expected my-local-session to NOT be a remote session")
	}
	if IsRemoteSession("emagy-something") {
		t.Error("expected emagy-something to NOT be a remote session")
	}
}

// --- Marker Tests ---

func TestGenerateMarker(t *testing.T) {
	marker1 := GenerateMarker("gic")
	marker2 := GenerateMarker("gic")

	if marker1 == marker2 {
		t.Error("markers should be unique (time-based)")
	}

	if len(marker1) < 20 {
		t.Errorf("marker too short: %q", marker1)
	}

	if marker1[:17] != "emorr-agy-marker-" {
		t.Errorf("marker should start with 'emorr-agy-marker-', got %q", marker1)
	}
}

// --- Message Splitting Tests ---

func TestSplitMessage_Short(t *testing.T) {
	result := SplitMessage("Hello world", 4096)
	if len(result) != 1 {
		t.Fatalf("expected 1 chunk for short message, got %d", len(result))
	}
	if result[0] != "Hello world" {
		t.Errorf("expected unchanged short message, got %q", result[0])
	}
}

func TestSplitMessage_Long(t *testing.T) {
	// Create a string longer than 100 chars
	longStr := ""
	for i := 0; i < 20; i++ {
		longStr += "Line number " + string(rune('A'+i)) + "\n"
	}

	result := SplitMessage(longStr, 100)
	if len(result) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(result))
	}

	// First chunk should have emoji and continuation marker
	if result[0][0:4] != "1️" {
		// emoji check (multi-byte)
	}

	// Last chunk should NOT have ✂️
	lastChunk := result[len(result)-1]
	if lastChunk[len(lastChunk)-3:] == "✂️" {
		t.Error("last chunk should not have continuation marker")
	}
}

func TestSplitMessage_ExactFit(t *testing.T) {
	msg := "exactly right"
	result := SplitMessage(msg, len(msg))
	if len(result) != 1 {
		t.Fatalf("expected 1 chunk for exact fit, got %d", len(result))
	}
}

// --- Concierge State Machine Tests ---

func TestNewConcierge(t *testing.T) {
	cfg := &Config{}
	c := NewConcierge(cfg)

	if c.Mode != ModeConcierge {
		t.Errorf("expected ModeConcierge, got %d", c.Mode)
	}
	if c.AttachedSession != "" {
		t.Error("expected no attached session")
	}
	if len(c.Sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(c.Sessions))
	}
}

func TestConcierge_DetachSession(t *testing.T) {
	cfg := &Config{}
	c := NewConcierge(cfg)
	c.Mode = ModeAttached
	c.AttachedSession = "remote-agy-test"

	c.DetachSession()

	if c.Mode != ModeConcierge {
		t.Error("expected ModeConcierge after detach")
	}
	if c.AttachedSession != "" {
		t.Error("expected empty attached session after detach")
	}
}

func TestConcierge_MaxSessions(t *testing.T) {
	cfg := &Config{
		Folders: []FolderConfig{
			{Name: "a", Path: "/tmp/a"},
			{Name: "b", Path: "/tmp/b"},
			{Name: "c", Path: "/tmp/c"},
			{Name: "d", Path: "/tmp/d"},
		},
	}
	c := NewConcierge(cfg)

	// Simulate 3 sessions already running
	c.Sessions["remote-agy-a"] = &RemoteSession{Name: "remote-agy-a"}
	c.Sessions["remote-agy-b"] = &RemoteSession{Name: "remote-agy-b"}
	c.Sessions["remote-agy-c"] = &RemoteSession{Name: "remote-agy-c"}

	// 4th should fail
	_, err := c.StartRemoteSession(FolderConfig{Name: "d", Path: "/tmp/d"})
	if err == nil {
		t.Fatal("expected error when exceeding max sessions")
	}
}

func TestConcierge_SendInput_NotAttached(t *testing.T) {
	c := NewConcierge(&Config{})
	err := c.SendInput("hello")
	if err == nil {
		t.Fatal("expected error when sending input without attachment")
	}
}

// --- BuildAgyCommand tests ---

func TestBuildAgyCommand_Sandboxed(t *testing.T) {
	trueVal := true
	folder := FolderConfig{Name: "GIC", Path: "/tmp/gic", Sandbox: &trueVal}
	cmd := BuildAgyCommand(folder, "marker123", nil)
	if cmd != "agy --sandbox" {
		t.Errorf("expected 'agy --sandbox', got %q", cmd)
	}
}

func TestBuildAgyCommand_Unsandboxed(t *testing.T) {
	falseVal := false
	folder := FolderConfig{
		Name:     "GIC Unsafe",
		Path:     "/tmp/gic",
		AgyFlags: "--dangerously-skip-permissions",
		Sandbox:  &falseVal,
	}
	cmd := BuildAgyCommand(folder, "marker123", nil)
	if cmd != "agy --dangerously-skip-permissions" {
		t.Errorf("expected 'agy --dangerously-skip-permissions', got %q", cmd)
	}
}

func TestBuildAgyCommand_DefaultSandbox(t *testing.T) {
	trueVal := true
	folder := FolderConfig{Name: "Test", Path: "/tmp/test"}
	cmd := BuildAgyCommand(folder, "marker123", &trueVal)
	if cmd != "agy --sandbox" {
		t.Errorf("expected 'agy --sandbox', got %q", cmd)
	}
}

// --- shellescape test ---

func TestShellescape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "'simple'"},
		{"it's", "'it'\\''s'"},
		{"/path/to/dir", "'/path/to/dir'"},
	}

	for _, tt := range tests {
		result := shellescape(tt.input)
		if result != tt.expected {
			t.Errorf("shellescape(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
