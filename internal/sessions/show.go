package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type ShowOptions struct {
	Classify bool `json:"classify"`
}

type LLMResult struct {
	About            string `json:"about"`
	UserInputPending bool   `json:"user_input_pending"`
	WorthResuscitate bool   `json:"worth_resuscitate"`
}

type LLMClassifier interface {
	Classify(sessionID string) (*LLMResult, error)
}

// isPIDActive checks if the process with the given PID is running.
// Can be overridden in unit tests to mock active processes.
var isPIDActive = func(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}
	// On UNIX-like systems, a signal error "operation not permitted" means the process is active.
	if err.Error() == "operation not permitted" {
		return true
	}
	return false
}

// ShowSession prints detailed information about a single session.
func ShowSession(w io.Writer, engine *ClassificationEngine, sessionID string, opts ShowOptions, classifier LLMClassifier) error {
	sessions, err := engine.Classify(nil)
	if err != nil {
		return err
	}

	var target *Session
	for i := range sessions {
		if sessions[i].ID == sessionID {
			target = &sessions[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("session %q not found", sessionID)
	}

	fmt.Fprintf(w, "SESSION ID:     %s\n", target.ID)
	fmt.Fprintf(w, "HARNESS:        %s\n", target.Harness)
	fmt.Fprintf(w, "STATE:          %s\n", target.State)
	folder := strings.ReplaceAll(target.Folder, "/usr/local/google/home/ricc", "~")
	fmt.Fprintf(w, "DIRECTORY:      %s\n", folder)
	fmt.Fprintf(w, "PROCESS/WINDOW: %d\n", target.ProcessCount)
	fmt.Fprintf(w, "RESUME CMD:     %s\n", target.ResumeCommand)

	printLastTranscriptLines(w, engine.fs, engine.homeDir, target.ID, DefaultLastLinesCount)

	if !opts.Classify {
		return nil
	}

	if classifier == nil {
		return fmt.Errorf("classifier interface not provided for classification request")
	}

	// 1. Concurrency Locking
	lockDir := filepath.Join(engine.homeDir, ".emorr-agy/locks")
	lockPath := filepath.Join(lockDir, sessionID+".lock")

	if lockData, err := engine.fs.ReadFile(lockPath); err == nil {
		if pid, err := strconv.Atoi(string(lockData)); err == nil {
			if isPIDActive(pid) {
				return fmt.Errorf("another script PID %d is doing this, try again later", pid)
			}
		}
	}

	// Create locks directory
	_ = engine.fs.MkdirAll(lockDir, 0755)

	// Write lock file
	myPID := os.Getpid()
	if err := engine.fs.WriteFile(lockPath, []byte(strconv.Itoa(myPID)), 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}
	defer func() {
		_ = engine.fs.Remove(lockPath)
	}()

	// 2. Cache Checking
	cacheDir := filepath.Join(engine.homeDir, ".emorr-agy/cache")
	cachePath := filepath.Join(cacheDir, sessionID+".json")

	var result *LLMResult
	if cacheData, err := engine.fs.ReadFile(cachePath); err == nil {
		var cachedResult LLMResult
		if err := json.Unmarshal(cacheData, &cachedResult); err == nil {
			result = &cachedResult
		}
	}

	// 3. LLM Classification (if not cached)
	if result == nil {
		res, err := classifier.Classify(sessionID)
		if err != nil {
			return fmt.Errorf("classification failed: %w", err)
		}
		result = res

		// Save to cache
		_ = engine.fs.MkdirAll(cacheDir, 0755)
		if data, err := json.Marshal(result); err == nil {
			_ = engine.fs.WriteFile(cachePath, data, 0644)
		}
	}

	// 4. Output classification results
	fmt.Fprintf(w, "\n--- LLM CLASSIFICATION ---\n")
	fmt.Fprintf(w, "About:               %s\n", result.About)
	fmt.Fprintf(w, "User Input Pending:  %t\n", result.UserInputPending)
	fmt.Fprintf(w, "Worth Resuscitating: %t\n", result.WorthResuscitate)

	return nil
}

func printLastTranscriptLines(w io.Writer, fs FileSystem, homeDir, sessionID string, count int) {
	transcriptPath := filepath.Join(homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript.jsonl")
	var data []byte
	var err error
	if _, statErr := fs.Stat(transcriptPath); statErr == nil {
		data, err = fs.ReadFile(transcriptPath)
	} else {
		transcriptPath = filepath.Join(homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript_full.jsonl")
		if _, statErr := fs.Stat(transcriptPath); statErr == nil {
			data, err = fs.ReadFile(transcriptPath)
		} else {
			fmt.Fprintf(w, "\nLast %d lines:\n", count)
			fmt.Fprintln(w, "------------")
			fmt.Fprintln(w, "(no transcript logs found for this session)")
			return
		}
	}

	if err != nil {
		fmt.Fprintf(w, "\nLast %d lines:\n", count)
		fmt.Fprintln(w, "------------")
		fmt.Fprintf(w, "(failed to read transcript: %v)\n", err)
		return
	}

	// Split lines
	rawLines := strings.Split(string(data), "\n")
	// Clean up empty trailing line if present
	if len(rawLines) > 0 && rawLines[len(rawLines)-1] == "" {
		rawLines = rawLines[:len(rawLines)-1]
	}

	// Get last N lines
	if len(rawLines) > count {
		rawLines = rawLines[len(rawLines)-count:]
	}

	fmt.Fprintf(w, "\nLast %d lines:\n", count)
	fmt.Fprintln(w, "------------")

	if len(rawLines) == 0 {
		fmt.Fprintln(w, "(transcript is empty)")
		return
	}

	for _, line := range rawLines {
		fmt.Fprintln(w, formatTranscriptLine(line))
	}
}

func formatTranscriptLine(raw string) string {
	var line struct {
		Source  string `json:"source"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(raw), &line); err != nil {
		return raw
	}
	content := strings.TrimSpace(line.Content)
	if len(content) > 150 {
		content = content[:147] + "..."
	}
	// Sanitize newlines and multiple spaces
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", "")
	content = strings.ReplaceAll(content, "\t", " ")
	content = strings.Join(strings.Fields(content), " ")

	return fmt.Sprintf("[%s:%s] %s", line.Source, line.Type, content)
}
