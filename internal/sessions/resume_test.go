package sessions

import (
	"encoding/json"
	"os/exec"
	"testing"
)

func TestResumeSession_NotFound(t *testing.T) {
	mockTmux := &MockTmuxRunner{}
	mockFS := &MockFileSystem{}
	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	err := ResumeSession(engine, "nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != `session "nonexistent-id" not found` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResumeSession_AlreadyRunning(t *testing.T) {
	mockTmux := &MockTmuxRunner{
		sessions: []TmuxSession{
			{Name: "emagy-session-123", Path: "/workspace/proj", Attached: true, Windows: 2},
		},
	}
	mockFS := &MockFileSystem{}
	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	var attachedSession string
	var syscallExecCalled bool
	// Save & restore
	origSyscall := syscallExec
	defer func() { syscallExec = origSyscall }()

	syscallExec = func(argv0 string, argv []string, envv []string) error {
		syscallExecCalled = true
		if len(argv) >= 4 && argv[1] == "attach-session" && argv[2] == "-t" {
			attachedSession = argv[3]
		}
		return nil
	}

	err := ResumeSession(engine, "session-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !syscallExecCalled {
		t.Error("expected syscallExec to be called")
	}
	if attachedSession != "emagy-session-123" {
		t.Errorf("expected attached session to be emagy-session-123, got %q", attachedSession)
	}
}

func TestResumeSession_DeadResuscitate(t *testing.T) {
	// Let's mock classification engine to return a dead session
	mockTmux := &MockTmuxRunner{}
	mockFS := &MockFileSystem{
		files: map[string][]byte{
			"/home/ricc/.gemini/antigravity-cli/cache/last_conversations.json": []byte(`{"/workspace/sre-extension": "f0a30f63-d1b1-4ff3-9220-98c410fb17a9"}`),
		},
	}
	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	var newSessionArgs []string
	var execCalled bool

	origExec := execCommand
	origSyscall := syscallExec
	defer func() {
		execCommand = origExec
		syscallExec = origSyscall
	}()

	// Mock execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		execCalled = true
		newSessionArgs = arg
		// Return a command that succeeds immediately when Run() is called
		return exec.Command("true")
	}

	var attachSessionName string
	syscallExec = func(argv0 string, argv []string, envv []string) error {
		if len(argv) >= 4 && argv[1] == "attach-session" && argv[2] == "-t" {
			attachSessionName = argv[3]
		}
		return nil
	}

	err := ResumeSession(engine, "f0a30f63-d1b1-4ff3-9220-98c410fb17a9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !execCalled {
		t.Error("expected execCommand to be called to start tmux session")
	}
	// Expected args: "new-session", "-d", "-s", "emagy-f0a30f63-d1b1-4ff3-9220-98c410fb17a9", "-c", "/workspace/sre-extension", "agy --conversation f0a30f63-d1b1-4ff3-9220-98c410fb17a9"
	if len(newSessionArgs) < 7 {
		t.Fatalf("expected at least 7 arguments for tmux new-session, got %v", newSessionArgs)
	}
	if newSessionArgs[0] != "new-session" || newSessionArgs[1] != "-d" || newSessionArgs[2] != "-s" ||
		newSessionArgs[3] != "emagy-f0a30f63-d1b1-4ff3-9220-98c410fb17a9" || newSessionArgs[4] != "-c" ||
		newSessionArgs[5] != "/workspace/sre-extension" || newSessionArgs[6] != "agy --conversation f0a30f63-d1b1-4ff3-9220-98c410fb17a9" {
		t.Errorf("unexpected tmux arguments: %v", newSessionArgs)
	}

	if attachSessionName != "emagy-f0a30f63-d1b1-4ff3-9220-98c410fb17a9" {
		t.Errorf("expected attach to emagy-f0a30f63-d1b1-4ff3-9220-98c410fb17a9, got %q", attachSessionName)
	}
}

func TestResuscitateAndArchiveSession(t *testing.T) {
	mockTmux := &MockTmuxRunner{}
	mockFS := &MockFileSystem{
		files: map[string][]byte{
			"/home/ricc/.gemini/antigravity-cli/cache/last_conversations.json": []byte(`{"/workspace/sre-extension": "f0a30f63-d1b1-4ff3-9220-98c410fb17a9"}`),
		},
	}
	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	var newSessionArgs []string
	var execCalled bool

	origExec := execCommand
	origSyscall := syscallExec
	defer func() {
		execCommand = origExec
		syscallExec = origSyscall
	}()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		execCalled = true
		newSessionArgs = arg
		return exec.Command("true")
	}

	var syscallExecCalled bool
	syscallExec = func(argv0 string, argv []string, envv []string) error {
		syscallExecCalled = true
		return nil
	}

	// 1. Test ResuscitateSession (should start in background but NOT attach)
	err := ResuscitateSession(engine, "f0a30f63-d1b1-4ff3-9220-98c410fb17a9")
	if err != nil {
		t.Fatalf("unexpected error during resuscitation: %v", err)
	}

	if !execCalled {
		t.Error("expected execCommand to be called")
	}
	if syscallExecCalled {
		t.Error("resuscitate should not attach (syscallExec should not be called)")
	}
	if len(newSessionArgs) < 7 || newSessionArgs[3] != "emagy-f0a30f63-d1b1-4ff3-9220-98c410fb17a9" {
		t.Errorf("unexpected resuscitation arguments: %v", newSessionArgs)
	}

	// 2. Test ArchiveSession
	err = ArchiveSession(engine, "f0a30f63-d1b1-4ff3-9220-98c410fb17a9")
	if err != nil {
		t.Fatalf("unexpected error during archiving: %v", err)
	}

	// Should have written false worth_resuscitate to cache
	cachePath := "/home/ricc/.emorr-agy/cache/f0a30f63-d1b1-4ff3-9220-98c410fb17a9.json"
	writtenData, ok := mockFS.files[cachePath]
	if !ok {
		t.Fatal("expected cache file to be written")
	}
	var cacheRes struct {
		WorthResuscitate bool `json:"worth_resuscitate"`
	}
	if err := json.Unmarshal(writtenData, &cacheRes); err != nil {
		t.Fatalf("failed to decode written cache: %v", err)
	}
	if cacheRes.WorthResuscitate {
		t.Error("expected worth_resuscitate to be false")
	}
}

