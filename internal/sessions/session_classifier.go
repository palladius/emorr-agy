package sessions

import (
	"os"
	"path/filepath"
	"time"
)

// RecencyWindow is the duration after which a session is considered stale.
// Sessions active within this window are still "in scope".
const RecencyWindow = 7 * 24 * time.Hour // 7 days

// SessionMetadata holds the metadata needed for classification.
// This is a lightweight struct that can be built from Session + transcript analysis.
type SessionMetadata struct {
	ID                       string
	TranscriptPath           string
	RepoName                 string
	LastActivity             time.Time
	IsStructurallyInterrupted bool
	IsExplicitQuit           bool
}

// DetectSupersession takes a list of session metadata and returns a map
// of session ID → whether that session has been superseded by a newer
// session on the same repository.
func DetectSupersession(sessions []SessionMetadata) map[string]bool {
	superseded := make(map[string]bool)

	// Group sessions by repo name, tracking the latest per repo
	latestByRepo := make(map[string]time.Time)
	latestIDByRepo := make(map[string]string)

	for _, s := range sessions {
		if s.RepoName == "" || s.RepoName == "(system)" {
			continue
		}
		if existing, ok := latestByRepo[s.RepoName]; !ok || s.LastActivity.After(existing) {
			latestByRepo[s.RepoName] = s.LastActivity
			latestIDByRepo[s.RepoName] = s.ID
		}
	}

	// Mark older sessions on the same repo as superseded
	for _, s := range sessions {
		if s.RepoName == "" || s.RepoName == "(system)" {
			continue
		}
		if latestID, ok := latestIDByRepo[s.RepoName]; ok && latestID != s.ID {
			superseded[s.ID] = true
		}
	}

	return superseded
}

// ClassifySession determines the 3-state classification for a single session.
// It combines structural interruption, explicit quit, recency window, and
// supersession status to produce the final classification.
func ClassifySession(meta SessionMetadata, isSuperseded bool, now time.Time) SessionClassification {
	// Rule 1: Explicit quit always means FINISHED
	if meta.IsExplicitQuit {
		return ClassFinished
	}

	age := now.Sub(meta.LastActivity)
	isRecent := age <= RecencyWindow

	// Rule 2: Recent + superseded → OBSOLETE
	if isRecent && isSuperseded {
		return ClassObsolete
	}

	// Rule 3: Recent + not superseded → NEEDS_RESUME (if interrupted) or NEEDS_RESUME (default for recent)
	if isRecent {
		// Any recent session that wasn't explicitly quit needs resume
		return ClassNeedsResume
	}

	// Rule 4: Old + structurally interrupted → OBSOLETE (was interrupted, now stale)
	if meta.IsStructurallyInterrupted {
		return ClassObsolete
	}

	// Rule 5: Old + clean → FINISHED
	return ClassFinished
}

// ClassifySessionFromTranscript is a convenience function that performs
// transcript analysis and classification in one step.
func ClassifySessionFromTranscript(id string, transcriptPath string, lastActivity time.Time, isSuperseded bool, now time.Time) SessionClassification {
	meta := SessionMetadata{
		ID:                        id,
		TranscriptPath:            transcriptPath,
		LastActivity:              lastActivity,
		IsStructurallyInterrupted: IsStructurallyInterrupted(transcriptPath),
		IsExplicitQuit:            IsExplicitQuit(transcriptPath),
	}
	return ClassifySession(meta, isSuperseded, now)
}

// BuildSessionMetadata creates a SessionMetadata struct from a Session by
// analyzing its transcript file.
func BuildSessionMetadata(s Session, homeDir string) SessionMetadata {
	transcriptPath := findTranscriptPath(s.ID, homeDir)
	return SessionMetadata{
		ID:                        s.ID,
		TranscriptPath:            transcriptPath,
		RepoName:                  ExtractRepoName(transcriptPath),
		LastActivity:              s.LastActivity,
		IsStructurallyInterrupted: IsStructurallyInterrupted(transcriptPath),
		IsExplicitQuit:            IsExplicitQuit(transcriptPath),
	}
}

// findTranscriptPath returns the path to the transcript.jsonl for a given session ID.
func findTranscriptPath(sessionID string, homeDir string) string {
	// Try Antigravity brain path first
	brainPath := filepath.Join(homeDir, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript.jsonl")
	if _, err := osStat(brainPath); err == nil {
		return brainPath
	}
	// Try full transcript
	fullPath := filepath.Join(homeDir, ".gemini", "antigravity-cli", "brain", sessionID, ".system_generated", "logs", "transcript_full.jsonl")
	if _, err := osStat(fullPath); err == nil {
		return fullPath
	}
	return brainPath // return default even if not found
}

// osStat is a package-level variable for testing.
var osStat = os.Stat
