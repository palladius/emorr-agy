package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/palladius/emorr-agy/internal/color"
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

	trimmedID := target.ID
	if idx := strings.Index(target.ID, "-"); idx != -1 {
		if strings.HasPrefix(target.ID, "emagy-") {
			trimmedID = strings.TrimPrefix(target.ID, "emagy-")
		} else if strings.HasPrefix(target.ID, "emgem-") {
			trimmedID = strings.TrimPrefix(target.ID, "emgem-")
		} else if strings.HasPrefix(target.ID, "emcld-") {
			trimmedID = strings.TrimPrefix(target.ID, "emcld-")
		}
	}
	activeConvs := engine.FindActiveConvs()
	pid := activeConvs[trimmedID]
	if pid == 0 {
		pid = activeConvs[target.ID]
	}
	detailedState := InferDetailedState(engine.homeDir, target.ID, target.State, pid)

	fmt.Fprintf(w, "SESSION ID:     %s\n", target.ID)
	fmt.Fprintf(w, "HARNESS:        %s\n", target.Harness)
	fmt.Fprintf(w, "STATE:          %s\n", detailedState)
	folder := strings.ReplaceAll(target.Folder, "/usr/local/google/home/ricc", "~")
	fmt.Fprintf(w, "DIRECTORY:      %s\n", folder)
	fmt.Fprintf(w, "PROCESS/WINDOW: %d\n", target.ProcessCount)
	fmt.Fprintf(w, "RESUME CMD:     %s\n", target.ResumeCommand)

	printLastTranscriptLines(w, engine.fs, engine.homeDir, target.ID, DefaultLastLinesCount)

	// 1. Cache Checking
	cacheDir := filepath.Join(engine.homeDir, ".emorr-agy/cache")
	cachePath := filepath.Join(cacheDir, sessionID+".json")

	var result *LLMResult
	if cacheData, err := engine.fs.ReadFile(cachePath); err == nil {
		var cachedResult LLMResult
		if err := json.Unmarshal(cacheData, &cachedResult); err == nil {
			result = &cachedResult
		}
	}

	// 2. Compute LLM Classification (if not cached and requested)
	if result == nil && opts.Classify {
		if classifier == nil {
			return fmt.Errorf("classifier interface not provided for classification request")
		}

		// Concurrency Locking
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

	// 3. Output classification results if available
	if result != nil {
		fmt.Fprintf(w, "\n--- LLM CLASSIFICATION ---\n")
		fmt.Fprintf(w, "About:               %s\n", result.About)
		fmt.Fprintf(w, "User Input Pending:  %t\n", result.UserInputPending)
		fmt.Fprintf(w, "Worth Resuscitating: %t\n", result.WorthResuscitate)
	}

	return nil
}

func printLastTranscriptLines(w io.Writer, fs FileSystem, homeDir, sessionID string, count int) {
	// 1. Try to capture from tmux pane first if active
	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", sessionID)
	if output, err := cmd.Output(); err == nil {
		rawLines := strings.Split(string(output), "\n")
		var lines []string
		for _, line := range rawLines {
			lines = append(lines, line)
		}
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}

		fmt.Fprintf(w, "\nLast %d lines (captured from tmux pane):\n", count)
		fmt.Fprintln(w, "------------")
		start := len(lines) - count
		if start < 0 {
			start = 0
		}
		for i := start; i < len(lines); i++ {
			fmt.Fprintln(w, color.Colorize(lines[i], color.Cyan))
		}
		return
	}

	// 2. Fall back to transcript file
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
			// Try with trimmed prefix too
			trimmedID := sessionID
			if idx := strings.Index(sessionID, "-"); idx != -1 {
				if strings.HasPrefix(sessionID, "emagy-") {
					trimmedID = strings.TrimPrefix(sessionID, "emagy-")
				} else if strings.HasPrefix(sessionID, "emgem-") {
					trimmedID = strings.TrimPrefix(sessionID, "emgem-")
				} else if strings.HasPrefix(sessionID, "emcld-") {
					trimmedID = strings.TrimPrefix(sessionID, "emcld-")
				}
			}
			if trimmedID != sessionID {
				transcriptPath = filepath.Join(homeDir, ".gemini/antigravity-cli/brain", trimmedID, ".system_generated/logs/transcript.jsonl")
				if _, statErr := fs.Stat(transcriptPath); statErr == nil {
					data, err = fs.ReadFile(transcriptPath)
				} else {
					transcriptPath = filepath.Join(homeDir, ".gemini/antigravity-cli/brain", trimmedID, ".system_generated/logs/transcript_full.jsonl")
					if _, statErr := fs.Stat(transcriptPath); statErr == nil {
						data, err = fs.ReadFile(transcriptPath)
					}
				}
			}
		}
	}

	if len(data) == 0 {
		fmt.Fprintf(w, "\nLast %d lines:\n", count)
		fmt.Fprintln(w, "------------")
		fmt.Fprintln(w, "(no transcript logs found for this session)")
		return
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

func InferDetailedState(homeDir, sessionID string, baseState SessionState, pid int) string {
	if baseState != StateOpenTmux && baseState != StateOpenAgy && baseState != StateOpenPrivate {
		switch baseState {
		case StateDeadResuscitatable:
			return "dead_resuscitatable (💤)"
		case StateDeadArchived:
			return "dead_archived (⚫)"
		default:
			return string(baseState)
		}
	}

	// 1. If running in tmux, check if the pane contents indicate it is waiting on user input
	if baseState == StateOpenTmux || baseState == StateOpenPrivate {
		if isTmuxPaneWaitingOnUser(sessionID) {
			return "Waiting on User 💬"
		}
	}

	// 2. Trim prefix for DB path
	trimmedID := sessionID
	if idx := strings.Index(sessionID, "-"); idx != -1 {
		if strings.HasPrefix(sessionID, "emagy-") {
			trimmedID = strings.TrimPrefix(sessionID, "emagy-")
		} else if strings.HasPrefix(sessionID, "emgem-") {
			trimmedID = strings.TrimPrefix(sessionID, "emgem-")
		} else if strings.HasPrefix(sessionID, "emcld-") {
			trimmedID = strings.TrimPrefix(sessionID, "emcld-")
		}
	}

	dbPath := filepath.Join(homeDir, ".gemini/antigravity-cli/conversations", trimmedID+".db")

	// 3. Check child processes of PID if PID is valid
	if pid > 0 {
		if hasChildProcesses(pid) {
			return "Working (Tool Calling) 🛠️"
		}
	}

	// 4. Check steps table in SQLite DB
	stepType, status, err := getLatestStepFromDB(dbPath)
	if err == nil {
		if status == 3 { // Done
			return "Waiting on User 💬"
		} else if stepType > 0 {
			return "Working (Tool Calling) 🛠️"
		}
	}

	// Default to Generating for active sessions if we don't match any specific states
	if baseState == StateOpenPrivate {
		return "Working (Interactive Session) 🖥️"
	}
	return "Working (Generating) ✍️"
}

func isTmuxPaneWaitingOnUser(sessionID string) bool {
	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", sessionID)
	if output, err := cmd.Output(); err == nil {
		content := strings.ToLower(string(output))
		indicators := []string{
			"do you want to proceed?",
			"> 1. yes",
			">  1. yes",
			"do you want to run",
			"do you want to execute",
			"do you want to allow",
			"(y/n)",
			"[y/n]",
			"proceed?",
			"approve?",
			"press enter to",
		}
		for _, ind := range indicators {
			if strings.Contains(content, ind) {
				return true
			}
		}
	}
	return false
}

func hasChildProcesses(parentPID int) bool {
	files, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		pidStr := file.Name()
		_, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		statPath := filepath.Join("/proc", pidStr, "stat")
		data, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}
		content := string(data)
		lastParen := strings.LastIndex(content, ")")
		if lastParen == -1 {
			continue
		}
		parts := strings.Fields(content[lastParen+2:])
		if len(parts) >= 2 {
			ppid, _ := strconv.Atoi(parts[1])
			if ppid == parentPID {
				return true
			}
		}
	}
	return false
}

func getLatestStepFromDB(dbPath string) (int, int, error) {
	cmd := exec.Command("sqlite3", dbPath, "select step_type, status from steps order by idx desc limit 1")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 2 {
		stepType, _ := strconv.Atoi(parts[0])
		status, _ := strconv.Atoi(parts[1])
		return stepType, status, nil
	}

	return 0, 0, nil
}

// GetSessionDetailsAndOptions returns formatted details of a session, its last 10 lines of tmux pane, and any parsed option choices.
func GetSessionDetailsAndOptions(homeDir string, sessionID string) (string, []PaneOption, bool, error) {
	engine := NewClassificationEngine(RealTmuxRunner{}, OSFileSystem{}, homeDir)
	sessions, err := engine.Classify(nil)
	if err != nil {
		return "", nil, false, err
	}

	var target *Session
	for i := range sessions {
		if sessions[i].ID == sessionID ||
			strings.TrimPrefix(sessions[i].ID, "emagy-") == sessionID ||
			strings.TrimPrefix(sessions[i].ID, "emgem-") == sessionID ||
			strings.TrimPrefix(sessions[i].ID, "emcld-") == sessionID {
			target = &sessions[i]
			break
		}
	}
	if target == nil {
		return "", nil, false, fmt.Errorf("session %q not found", sessionID)
	}

	trimmedID := target.ID
	if idx := strings.Index(target.ID, "-"); idx != -1 {
		if strings.HasPrefix(target.ID, "emagy-") {
			trimmedID = strings.TrimPrefix(target.ID, "emagy-")
		} else if strings.HasPrefix(target.ID, "emgem-") {
			trimmedID = strings.TrimPrefix(target.ID, "emgem-")
		} else if strings.HasPrefix(target.ID, "emcld-") {
			trimmedID = strings.TrimPrefix(target.ID, "emcld-")
		}
	}
	activeConvs := engine.FindActiveConvs()
	pid := activeConvs[trimmedID]
	if pid == 0 {
		pid = activeConvs[target.ID]
	}
	detailedState := InferDetailedState(homeDir, target.ID, target.State, pid)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*SESSION ID:* %s\n", target.ID))
	sb.WriteString(fmt.Sprintf("*HARNESS:*    %s\n", target.Harness))
	sb.WriteString(fmt.Sprintf("*STATE:*      %s\n", detailedState))
	folder := strings.ReplaceAll(target.Folder, "/usr/local/google/home/ricc", "~")
	sb.WriteString(fmt.Sprintf("*DIRECTORY:*  %s\n", folder))
	sb.WriteString(fmt.Sprintf("*PROCESS/WINDOW:* %d\n", target.ProcessCount))
	sb.WriteString(fmt.Sprintf("*RESUME CMD:* `%s`\n", target.ResumeCommand))

	// Get last 10 lines
	var tmuxLines []string
	cmd := exec.Command("tmux", "capture-pane", "-p", "-t", target.ID)
	if output, err := cmd.Output(); err == nil {
		rawLines := strings.Split(string(output), "\n")
		for _, line := range rawLines {
			tmuxLines = append(tmuxLines, line)
		}
		for len(tmuxLines) > 0 && strings.TrimSpace(tmuxLines[len(tmuxLines)-1]) == "" {
			tmuxLines = tmuxLines[:len(tmuxLines)-1]
		}
	}

	sb.WriteString("\n*Last 10 lines (captured from tmux pane):*\n")
	sb.WriteString("------------\n")
	if len(tmuxLines) > 0 {
		start := len(tmuxLines) - 10
		if start < 0 {
			start = 0
		}
		var outputLines []string
		for i := start; i < len(tmuxLines); i++ {
			outputLines = append(outputLines, tmuxLines[i])
			sb.WriteString(fmt.Sprintf("`%s`\n", tmuxLines[i]))
		}
		// Parse options from these captured lines
		opts := ParsePaneOptions(outputLines)
		isDead := target.State == StateDeadResuscitatable || target.State == StateDeadArchived
		return sb.String(), opts, isDead, nil
	}

	sb.WriteString("(no active tmux output or pane exited)\n")
	isDead := target.State == StateDeadResuscitatable || target.State == StateDeadArchived
	return sb.String(), nil, isDead, nil
}
