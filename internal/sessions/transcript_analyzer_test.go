package sessions

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTranscript is a test helper that writes JSONL content to a temp file and returns the path.
func writeTranscript(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}
	return path
}

// --- IsStructurallyInterrupted tests ---

func TestIsStructurallyInterrupted_LastStepPlannerWithToolCalls(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix the bug","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"I'll fix it","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/home/riccardo/git/myrepo/main.go"}}]}`
	path := writeTranscript(t, dir, content)

	if !IsStructurallyInterrupted(path) {
		t.Error("expected IsStructurallyInterrupted=true when last step is PLANNER_RESPONSE with tool_calls")
	}
}

func TestIsStructurallyInterrupted_LastStepPlannerWithoutToolCalls(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix the bug","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"Done! I fixed it.","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if IsStructurallyInterrupted(path) {
		t.Error("expected IsStructurallyInterrupted=false when last step is PLANNER_RESPONSE without tool_calls")
	}
}

func TestIsStructurallyInterrupted_LastStepUserInput(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"hello","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if IsStructurallyInterrupted(path) {
		t.Error("expected IsStructurallyInterrupted=false when last step is USER_INPUT")
	}
}

func TestIsStructurallyInterrupted_EmptyTranscript(t *testing.T) {
	dir := t.TempDir()
	path := writeTranscript(t, dir, "")

	if IsStructurallyInterrupted(path) {
		t.Error("expected IsStructurallyInterrupted=false for empty transcript")
	}
}

func TestIsStructurallyInterrupted_FileDoesNotExist(t *testing.T) {
	if IsStructurallyInterrupted("/nonexistent/path/transcript.jsonl") {
		t.Error("expected IsStructurallyInterrupted=false when file doesn't exist")
	}
}

// --- IsExplicitQuit tests ---

func TestIsExplicitQuit_SlashQuit(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"/quit","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true for /quit")
	}
}

func TestIsExplicitQuit_SlashExit(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"/exit","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true for /exit")
	}
}

func TestIsExplicitQuit_Exit(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"exit","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true for exit")
	}
}

func TestIsExplicitQuit_Quit(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"quit","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true for quit")
	}
}

func TestIsExplicitQuit_SlashQ(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"/q","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true for /q")
	}
}

func TestIsExplicitQuit_NormalPrompt(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix the bug","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=false for normal prompt 'fix the bug'")
	}
}

func TestIsExplicitQuit_XMLWrapped(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"<USER_REQUEST>exit</USER_REQUEST>\n<ADDITIONAL_METADATA>\nsome metadata\n</ADDITIONAL_METADATA>","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true for XML-wrapped exit command")
	}
}

func TestIsExplicitQuit_NoUserInput(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"Hello!","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=false when no USER_INPUT steps exist")
	}
}

func TestIsExplicitQuit_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := writeTranscript(t, dir, "")

	if IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=false for empty file")
	}
}

func TestIsExplicitQuit_LastUserInputIsQuit(t *testing.T) {
	// The LAST USER_INPUT should be checked, not the first
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix the bug","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"Done!","tool_calls":[]}
{"step_index":2,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"/quit","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if !IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=true when the last USER_INPUT is /quit")
	}
}

func TestIsExplicitQuit_LastUserInputIsNotQuit(t *testing.T) {
	// First USER_INPUT is quit but last is not → should be false
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"/quit","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"Bye!","tool_calls":[]}
{"step_index":2,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"actually keep going","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	if IsExplicitQuit(path) {
		t.Error("expected IsExplicitQuit=false when the last USER_INPUT is not a quit command")
	}
}

// --- ExtractRepoName tests ---

func TestExtractRepoName_FromToolCallArgs(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix it","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"fixing","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/home/riccardo/git/myrepo/src/main.go"}}]}`
	path := writeTranscript(t, dir, content)

	got := ExtractRepoName(path)
	if got != "myrepo" {
		t.Errorf("expected repo name 'myrepo', got %q", got)
	}
}

func TestExtractRepoName_MostCommonWins(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix it","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"fixing","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/home/riccardo/git/repo-a/main.go"}},{"name":"view_file","args":{"AbsolutePath":"/home/riccardo/git/repo-b/main.go"}}]}
{"step_index":2,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"more","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/home/riccardo/git/repo-b/lib.go"}},{"name":"view_file","args":{"AbsolutePath":"/home/riccardo/git/repo-b/test.go"}}]}`
	path := writeTranscript(t, dir, content)

	got := ExtractRepoName(path)
	if got != "repo-b" {
		t.Errorf("expected most common repo 'repo-b', got %q", got)
	}
}

func TestExtractRepoName_FromContent(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix it","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"Looking at /home/riccardo/git/emorr-agy/main.go","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	got := ExtractRepoName(path)
	if got != "emorr-agy" {
		t.Errorf("expected repo name 'emorr-agy', got %q", got)
	}
}

func TestExtractRepoName_NoPaths(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"hello world","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"Hi there!","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	got := ExtractRepoName(path)
	if got != "" {
		t.Errorf("expected empty repo name, got %q", got)
	}
}

func TestExtractRepoName_FilterSystemNames(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix it","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"checking /home/riccardo/git/bin/something","tool_calls":[]}`
	path := writeTranscript(t, dir, content)

	got := ExtractRepoName(path)
	if got != "" {
		t.Errorf("expected empty repo name for system path 'bin', got %q", got)
	}
}

func TestExtractRepoName_MacOSPaths(t *testing.T) {
	dir := t.TempDir()
	content := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"fix it","tool_calls":[]}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","content":"fixing","tool_calls":[{"name":"view_file","args":{"AbsolutePath":"/Users/riccardo/git/coolproject/main.go"}}]}`
	path := writeTranscript(t, dir, content)

	got := ExtractRepoName(path)
	if got != "coolproject" {
		t.Errorf("expected repo name 'coolproject', got %q", got)
	}
}

func TestExtractRepoName_FileDoesNotExist(t *testing.T) {
	got := ExtractRepoName("/nonexistent/path/transcript.jsonl")
	if got != "" {
		t.Errorf("expected empty repo name for nonexistent file, got %q", got)
	}
}
