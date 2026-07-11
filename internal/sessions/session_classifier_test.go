package sessions

import (
	"os"
	"testing"
	"time"
)

func TestDetectSupersession_NewerOnSameRepo(t *testing.T) {
	now := time.Now()
	sessions := []SessionMetadata{
		{ID: "old-session", RepoName: "myrepo", LastActivity: now.Add(-2 * time.Hour)},
		{ID: "new-session", RepoName: "myrepo", LastActivity: now},
	}

	superseded := DetectSupersession(sessions)

	if !superseded["old-session"] {
		t.Error("expected old-session to be superseded")
	}
	if superseded["new-session"] {
		t.Error("expected new-session to NOT be superseded")
	}
}

func TestDetectSupersession_OnlySessionForRepo(t *testing.T) {
	sessions := []SessionMetadata{
		{ID: "only-session", RepoName: "myrepo", LastActivity: time.Now()},
	}

	superseded := DetectSupersession(sessions)

	if superseded["only-session"] {
		t.Error("expected only-session to NOT be superseded")
	}
}

func TestDetectSupersession_DifferentRepos(t *testing.T) {
	now := time.Now()
	sessions := []SessionMetadata{
		{ID: "session-a", RepoName: "repo-a", LastActivity: now.Add(-1 * time.Hour)},
		{ID: "session-b", RepoName: "repo-b", LastActivity: now},
	}

	superseded := DetectSupersession(sessions)

	if superseded["session-a"] {
		t.Error("expected session-a to NOT be superseded (different repo)")
	}
	if superseded["session-b"] {
		t.Error("expected session-b to NOT be superseded (different repo)")
	}
}

func TestDetectSupersession_EmptyRepoName(t *testing.T) {
	now := time.Now()
	sessions := []SessionMetadata{
		{ID: "s1", RepoName: "", LastActivity: now.Add(-1 * time.Hour)},
		{ID: "s2", RepoName: "", LastActivity: now},
	}

	superseded := DetectSupersession(sessions)

	if superseded["s1"] || superseded["s2"] {
		t.Error("sessions with empty repo names should never be superseded")
	}
}

func TestDetectSupersession_SystemRepoIgnored(t *testing.T) {
	now := time.Now()
	sessions := []SessionMetadata{
		{ID: "s1", RepoName: "(system)", LastActivity: now.Add(-1 * time.Hour)},
		{ID: "s2", RepoName: "(system)", LastActivity: now},
	}

	superseded := DetectSupersession(sessions)

	if superseded["s1"] || superseded["s2"] {
		t.Error("sessions with (system) repo should never be superseded")
	}
}

func TestDetectSupersession_ThreeSessionsSameRepo(t *testing.T) {
	now := time.Now()
	sessions := []SessionMetadata{
		{ID: "oldest", RepoName: "myrepo", LastActivity: now.Add(-3 * time.Hour)},
		{ID: "middle", RepoName: "myrepo", LastActivity: now.Add(-1 * time.Hour)},
		{ID: "newest", RepoName: "myrepo", LastActivity: now},
	}

	superseded := DetectSupersession(sessions)

	if !superseded["oldest"] {
		t.Error("expected oldest to be superseded")
	}
	if !superseded["middle"] {
		t.Error("expected middle to be superseded")
	}
	if superseded["newest"] {
		t.Error("expected newest to NOT be superseded")
	}
}

// --- ClassifySession tests ---

func TestClassifySession_ExplicitQuitAlwaysFinished(t *testing.T) {
	now := time.Now()
	meta := SessionMetadata{
		ID:                        "quit-session",
		LastActivity:              now.Add(-1 * time.Hour), // recent
		IsExplicitQuit:            true,
		IsStructurallyInterrupted: true, // even if interrupted, quit wins
	}

	result := ClassifySession(meta, false, now)
	if result != ClassFinished {
		t.Errorf("expected ClassFinished, got %s", result)
	}
}

func TestClassifySession_RecentNotSupersededNeedsResume(t *testing.T) {
	now := time.Now()
	meta := SessionMetadata{
		ID:           "active-session",
		LastActivity: now.Add(-1 * time.Hour),
	}

	result := ClassifySession(meta, false, now)
	if result != ClassNeedsResume {
		t.Errorf("expected ClassNeedsResume, got %s", result)
	}
}

func TestClassifySession_RecentSupersededObsolete(t *testing.T) {
	now := time.Now()
	meta := SessionMetadata{
		ID:           "superseded-session",
		LastActivity: now.Add(-1 * time.Hour),
	}

	result := ClassifySession(meta, true, now)
	if result != ClassObsolete {
		t.Errorf("expected ClassObsolete, got %s", result)
	}
}

func TestClassifySession_OldStructurallyInterruptedObsolete(t *testing.T) {
	now := time.Now()
	meta := SessionMetadata{
		ID:                        "old-interrupted",
		LastActivity:              now.Add(-10 * 24 * time.Hour), // 10 days ago
		IsStructurallyInterrupted: true,
	}

	result := ClassifySession(meta, false, now)
	if result != ClassObsolete {
		t.Errorf("expected ClassObsolete, got %s", result)
	}
}

func TestClassifySession_OldCleanFinished(t *testing.T) {
	now := time.Now()
	meta := SessionMetadata{
		ID:           "old-clean",
		LastActivity: now.Add(-10 * 24 * time.Hour), // 10 days ago
	}

	result := ClassifySession(meta, false, now)
	if result != ClassFinished {
		t.Errorf("expected ClassFinished, got %s", result)
	}
}

func TestClassifySession_ExactRecencyBoundary(t *testing.T) {
	now := time.Now()
	// Exactly 7 days ago should still be in the window
	meta := SessionMetadata{
		ID:           "boundary",
		LastActivity: now.Add(-RecencyWindow),
	}

	result := ClassifySession(meta, false, now)
	if result != ClassNeedsResume {
		t.Errorf("expected ClassNeedsResume at boundary, got %s", result)
	}
}

func TestClassifySession_JustOutsideRecencyWindow(t *testing.T) {
	now := time.Now()
	meta := SessionMetadata{
		ID:           "just-outside",
		LastActivity: now.Add(-RecencyWindow - 1*time.Second),
	}

	result := ClassifySession(meta, false, now)
	if result != ClassFinished {
		t.Errorf("expected ClassFinished just outside window, got %s", result)
	}
}

// --- BuildSessionMetadata test ---

func TestBuildSessionMetadata(t *testing.T) {
	// Create a temp transcript file
	tmpDir := t.TempDir()
	homeDir := tmpDir

	// Set up mock osStat to fail (no transcript files exist)
	origStat := osStat
	osStat = func(name string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}
	defer func() { osStat = origStat }()

	s := Session{
		ID:           "test-uuid",
		LastActivity: time.Now(),
	}

	meta := BuildSessionMetadata(s, homeDir)

	if meta.ID != "test-uuid" {
		t.Errorf("expected ID 'test-uuid', got '%s'", meta.ID)
	}
	if meta.IsStructurallyInterrupted {
		t.Error("expected no structural interruption for missing transcript")
	}
	if meta.IsExplicitQuit {
		t.Error("expected no explicit quit for missing transcript")
	}
}

// --- ClassifySessionFromTranscript test ---

func TestClassifySessionFromTranscript_WithQuitTranscript(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptFile := tmpDir + "/transcript.jsonl"

	// Write a transcript where last user input is /quit
	content := `{"step_index": 0, "type": "USER_INPUT", "content": "fix the bug", "tool_calls": []}
{"step_index": 1, "type": "PLANNER_RESPONSE", "content": "Done!", "tool_calls": []}
{"step_index": 2, "type": "USER_INPUT", "content": "/quit", "tool_calls": []}
`
	os.WriteFile(transcriptFile, []byte(content), 0644)

	now := time.Now()
	result := ClassifySessionFromTranscript("test-id", transcriptFile, now.Add(-1*time.Hour), false, now)

	if result != ClassFinished {
		t.Errorf("expected ClassFinished for /quit transcript, got %s", result)
	}
}

func TestClassifySessionFromTranscript_WithInterruptedTranscript(t *testing.T) {
	tmpDir := t.TempDir()
	transcriptFile := tmpDir + "/transcript.jsonl"

	// Write a transcript where last step is PLANNER_RESPONSE with tool_calls
	content := `{"step_index": 0, "type": "USER_INPUT", "content": "fix the bug", "tool_calls": []}
{"step_index": 1, "type": "PLANNER_RESPONSE", "content": "Let me check", "tool_calls": [{"name": "view_file", "args": {"path": "/tmp/foo"}}]}
`
	os.WriteFile(transcriptFile, []byte(content), 0644)

	now := time.Now()
	result := ClassifySessionFromTranscript("test-id", transcriptFile, now.Add(-1*time.Hour), false, now)

	if result != ClassNeedsResume {
		t.Errorf("expected ClassNeedsResume for interrupted transcript, got %s", result)
	}
}
