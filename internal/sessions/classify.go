package sessions

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SessionState string

const (
	StateOpenTmux           SessionState = "open_tmux"
	StateOpenPrivate        SessionState = "open_private"
	StateOpenAgy            SessionState = "open_agy"
	StateDeadResuscitatable SessionState = "dead_resuscitatable"
	StateDeadArchived       SessionState = "dead_archived"
)

type Session struct {
	ID              string       `json:"id"`
	Harness         string       `json:"harness"` // "agy", "gemini", "claude", "unknown"
	State           SessionState `json:"state"`
	Folder          string       `json:"folder"`
	Title           string       `json:"title,omitempty"`
	Description     string       `json:"description,omitempty"`
	WorktreeBranch  string       `json:"worktree_branch,omitempty"`
	IsCron          bool         `json:"is_cron,omitempty"`
	LastActivity    time.Time    `json:"last_activity,omitempty"`
	ProcessCount    int          `json:"process_count"`
	ResumeCommand   string       `json:"resume_command,omitempty"`
	AttachedClients int          `json:"attached_clients"`
}

type TmuxSession struct {
	Name            string
	Path            string
	Attached        bool
	AttachedClients int
	Windows         int
}

type TmuxRunner interface {
	ListSessions() ([]TmuxSession, error)
}

type FileSystem interface {
	ReadDir(dirname string) ([]os.DirEntry, error)
	ReadFile(filename string) ([]byte, error)
	Readlink(name string) (string, error)
	Stat(name string) (os.FileInfo, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Remove(name string) error
}

type OSFileSystem struct{}

func (OSFileSystem) ReadDir(dirname string) ([]os.DirEntry, error) { return os.ReadDir(dirname) }
func (OSFileSystem) ReadFile(filename string) ([]byte, error)     { return os.ReadFile(filename) }
func (OSFileSystem) Readlink(name string) (string, error)         { return os.Readlink(name) }
func (OSFileSystem) Stat(name string) (os.FileInfo, error)        { return os.Stat(name) }
func (OSFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}
func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error { return os.MkdirAll(path, perm) }
func (OSFileSystem) Remove(name string) error                     { return os.Remove(name) }

type RealTmuxRunner struct{}

func (RealTmuxRunner) ListSessions() ([]TmuxSession, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}\t#{session_path}\t#{session_attached}\t#{session_windows}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var list []TmuxSession
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			attachedClients, _ := strconv.Atoi(parts[2])
			attached := attachedClients > 0
			windows, _ := strconv.Atoi(parts[3])
			list = append(list, TmuxSession{
				Name:            parts[0],
				Path:            parts[1],
				Attached:        attached,
				AttachedClients: attachedClients,
				Windows:         windows,
			})
		}
	}
	return list, nil
}

type ClassificationEngine struct {
	tmux    TmuxRunner
	fs      FileSystem
	homeDir string
}

func NewClassificationEngine(tmux TmuxRunner, fs FileSystem, homeDir string) *ClassificationEngine {
	return &ClassificationEngine{
		tmux:    tmux,
		fs:      fs,
		homeDir: homeDir,
	}
}

// Classify scans active tmux sessions and saved configurations to return a classified list.
func (c *ClassificationEngine) Classify(harnessFilter []string) ([]Session, error) {
	var sessions []Session
	runningMap := make(map[string]bool)

	// A. Find all active conversations based on open DB locks
	activeConvs := c.FindActiveConvs()

	// 1. Get running tmux sessions
	tmuxSessions, err := c.tmux.ListSessions()
	if err == nil {
		for _, ts := range tmuxSessions {
			harness := getHarnessFromPrefix(ts.Name)
			state := StateOpenTmux
			if harness == "unknown" {
				state = StateOpenPrivate
			}

			// Extract session ID from tmux name if it has prefix
			id := ts.Name
			if idx := strings.Index(ts.Name, "-"); idx != -1 && harness != "unknown" {
				if strings.HasPrefix(ts.Name, "emagy-") {
					id = strings.TrimPrefix(ts.Name, "emagy-")
				} else if strings.HasPrefix(ts.Name, "emgem-") {
					id = strings.TrimPrefix(ts.Name, "emgem-")
				} else if strings.HasPrefix(ts.Name, "emcld-") {
					id = strings.TrimPrefix(ts.Name, "emcld-")
				}
			}

			runningMap[id] = true
			runningMap[ts.Name] = true

			// Determine last activity time from SQLite DB
			var lastActivity time.Time
			dbPath := filepath.Join(c.homeDir, ".gemini/antigravity-cli/conversations", id+".db")
			if fi, err := c.fs.Stat(dbPath); err == nil {
				lastActivity = fi.ModTime()
			} else {
				if fi, err := c.fs.Stat(ts.Path); err == nil {
					lastActivity = fi.ModTime()
				}
			}

			// Read description/about from cache
			var description string
			cachePath := filepath.Join(c.homeDir, ".emorr-agy/cache", id+".json")
			if data, err := c.fs.ReadFile(cachePath); err == nil {
				var cacheRes struct {
					About string `json:"about"`
				}
				if err := json.Unmarshal(data, &cacheRes); err == nil {
					description = cacheRes.About
				}
			}
			if description == "" {
				description = c.getTranscriptDescription(id)
			}

			attachedClients := ts.AttachedClients
			if attachedClients == 0 && ts.Attached {
				attachedClients = 1
			}

			s := Session{
				ID:              ts.Name,
				Harness:         harness,
				State:           state,
				Folder:          ts.Path,
				ProcessCount:    ts.Windows,
				LastActivity:    lastActivity,
				ResumeCommand:   "tmux attach -t " + ts.Name,
				AttachedClients: attachedClients,
				Description:     description,
			}
			sessions = append(sessions, s)
		}
	}
	// 1b. Detect running Gemini CLI processes by cwd (they don't hold .db FDs open)
	geminiCwds := make(map[string]bool) // cwd → true for running gemini CLI processes
	if procEntries, err := c.fs.ReadDir("/proc"); err == nil {
		for _, entry := range procEntries {
			if !entry.IsDir() {
				continue
			}
			if _, err := strconv.Atoi(entry.Name()); err != nil {
				continue
			}
			cmdlinePath := filepath.Join("/proc", entry.Name(), "cmdline")
			cmdlineBytes, err := c.fs.ReadFile(cmdlinePath)
			if err != nil {
				continue
			}
			cmdline := string(cmdlineBytes)
			// Match Gemini CLI specifically (node process running /bin/gemini)
			if strings.Contains(cmdline, "/bin/gemini") && !strings.Contains(cmdline, "antigravity") {
				cwdPath := filepath.Join("/proc", entry.Name(), "cwd")
				if cwd, err := c.fs.Readlink(cwdPath); err == nil {
					geminiCwds[cwd] = true
				}
			}
		}
	}

	// 2. Load historical and active non-tmux sessions from cache file
	cacheFile := filepath.Join(c.homeDir, ".gemini/antigravity-cli/cache/last_conversations.json")
	if data, err := c.fs.ReadFile(cacheFile); err == nil {
		var cacheConvs map[string]string
		if err := json.Unmarshal(data, &cacheConvs); err == nil {
			// Deduplicate: cacheConvs maps folder→convID, but the same convID
			// can appear under multiple folders. Invert to convID→folder, keeping
			// the most specific (longest) folder path.
			dedupedConvs := make(map[string]string) // convID → folder
			for folder, convID := range cacheConvs {
				if existing, ok := dedupedConvs[convID]; !ok || len(folder) > len(existing) {
					dedupedConvs[convID] = folder
				}
			}

			for convID, folder := range dedupedConvs {
				// If it's already running in tmux, skip adding a duplicate session
				if runningMap[convID] {
					continue
				}

				harness := "agy" // fallback for historical
				// Check if a Gemini CLI process is running in this folder
				if geminiCwds[folder] {
					harness = "gemini"
				}
				// Check if this conversation belongs to Antigravity IDE (AG2UI)
				ag2uiDbPath := filepath.Join(c.homeDir, ".gemini/antigravity/conversations", convID+".db")
				if _, err := c.fs.Stat(ag2uiDbPath); err == nil {
					harness = "ag2ui"
				}
				var state SessionState
				var cachedAbout string
				worthResuscitate := true

				// Load cache details
				cachePath := filepath.Join(c.homeDir, ".emorr-agy/cache", convID+".json")
				if data, err := c.fs.ReadFile(cachePath); err == nil {
					var cacheRes struct {
						About            string `json:"about"`
						UserInputPending bool   `json:"user_input_pending"`
						WorthResuscitate bool   `json:"worth_resuscitate"`
					}
					if err := json.Unmarshal(data, &cacheRes); err == nil {
						worthResuscitate = cacheRes.WorthResuscitate
						cachedAbout = cacheRes.About
					}
				}
				if cachedAbout == "" {
					cachedAbout = c.getTranscriptDescription(convID)
				}

				// If it's active in background processes or a Gemini CLI is running in this folder
				if _, active := activeConvs[convID]; active || geminiCwds[folder] {
					state = StateOpenAgy
					runningMap[convID] = true
				} else {
					if !worthResuscitate || c.isExcluded(convID, folder) {
						state = StateDeadArchived
					} else {
						state = StateDeadResuscitatable
					}
				}

				// Determine last activity time from SQLite DB
				var lastActivity time.Time
				dbPath := filepath.Join(c.homeDir, ".gemini/antigravity-cli/conversations", convID+".db")
				if fi, err := c.fs.Stat(dbPath); err == nil {
					lastActivity = fi.ModTime()
				} else {
					// Also check AG2UI path
					ag2uiDbPath := filepath.Join(c.homeDir, ".gemini/antigravity/conversations", convID+".db")
					if fi, err := c.fs.Stat(ag2uiDbPath); err == nil {
						lastActivity = fi.ModTime()
					} else if fi, err := c.fs.Stat(folder); err == nil {
						lastActivity = fi.ModTime()
					}
				}

				s := Session{
					ID:            convID,
					Harness:       harness,
					State:         state,
					Folder:        folder,
					LastActivity:  lastActivity,
					ResumeCommand: "emorr-agy resume " + convID,
					Description:   cachedAbout,
				}
				sessions = append(sessions, s)
			}
		}
	}

	// 2b. Scan AG2UI (Antigravity IDE) conversations directory for additional sessions
	ag2uiConvsDir := filepath.Join(c.homeDir, ".gemini/antigravity/conversations")
	if entries, err := c.fs.ReadDir(ag2uiConvsDir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".db-shm") || strings.HasSuffix(name, ".db-wal") {
				continue
			}
			convID := strings.TrimSuffix(name, ".db")
			// Skip if already seen from tmux or CLI cache
			if runningMap[convID] {
				continue
			}

			// Check if already added from the CLI cache dedup step
			alreadyAdded := false
			for _, s := range sessions {
				if s.ID == convID {
					alreadyAdded = true
					break
				}
			}
			if alreadyAdded {
				continue
			}

			runningMap[convID] = true

			// Read annotation (.pbtxt) for title and archived status
			var title string
			var archived bool
			annotPath := filepath.Join(c.homeDir, ".gemini/antigravity/annotations", convID+".pbtxt")
			if annotData, err := c.fs.ReadFile(annotPath); err == nil {
				annotStr := string(annotData)
				// Parse title from text proto: title:"..."
				if idx := strings.Index(annotStr, "title:\""); idx != -1 {
					rest := annotStr[idx+7:]
					if endIdx := strings.Index(rest, "\""); endIdx != -1 {
						title = rest[:endIdx]
					}
				}
				archived = strings.Contains(annotStr, "archived:true")
			}

			// Determine last activity from DB modification time
			var lastActivity time.Time
			dbPath := filepath.Join(ag2uiConvsDir, name)
			if fi, err := c.fs.Stat(dbPath); err == nil {
				lastActivity = fi.ModTime()
			}

			// Determine state
			var state SessionState
			if _, active := activeConvs[convID]; active {
				state = StateOpenAgy
			} else if archived {
				state = StateDeadArchived
			} else if !lastActivity.IsZero() && time.Since(lastActivity) > 7*24*time.Hour {
				// Auto-archive stale AG2UI sessions (>7 days) since they
				// can't be resumed from the CLI and just clutter the list.
				state = StateDeadArchived
			} else {
				state = StateDeadResuscitatable
			}

			// For AG2UI, extract workspace from transcript paths.
			description := c.getTranscriptDescription(convID)
			folder, worktreeBranch := c.getTranscriptWorkspace(convID)

			s := Session{
				ID:             convID,
				Harness:        "ag2ui",
				State:          state,
				Folder:         folder,
				Title:          title,
				WorktreeBranch: worktreeBranch,
				LastActivity:   lastActivity,
				ResumeCommand:  "antigravity2.0 --conversation " + convID,
				Description:    description,
			}
			sessions = append(sessions, s)
		}
	}

	// 2c. Detect cron jobs: AG2UI sessions with duplicate descriptions are from scheduled tasks.
	ag2uiDescCount := make(map[string]int)
	for i := range sessions {
		if sessions[i].Harness == "ag2ui" && sessions[i].Description != "" {
			ag2uiDescCount[sessions[i].Description]++
		}
	}
	for i := range sessions {
		if sessions[i].Harness == "ag2ui" && ag2uiDescCount[sessions[i].Description] > 1 {
			sessions[i].IsCron = true
		}
	}

	// 3. Sort sessions: non-archived sessions on top, archived sessions at the bottom.
	// Within those two groups, sort by last activity (mod time) descending, latest first.
	sort.Slice(sessions, func(i, j int) bool {
		iArchived := sessions[i].State == StateDeadArchived
		jArchived := sessions[j].State == StateDeadArchived
		if iArchived != jArchived {
			return !iArchived
		}
		if sessions[i].LastActivity.Equal(sessions[j].LastActivity) {
			return sessions[i].ID < sessions[j].ID
		}
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	// 4. Filter by harness if filter is provided
	if len(harnessFilter) > 0 {
		var filtered []Session
		for _, s := range sessions {
			for _, f := range harnessFilter {
				if s.Harness == f {
					filtered = append(filtered, s)
					break
				}
			}
		}
		return filtered, nil
	}

	return sessions, nil
}

func (c *ClassificationEngine) FindActiveConvs() map[string]int {
	active := make(map[string]int)
	files, err := c.fs.ReadDir("/proc")
	if err != nil {
		return active
	}

	candidates := []string{"agy", "gemini", "claude", "emorr-agy", "language_server", "antigravity", "python", "python3", "node", "go", "bash", "sh"}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		pidStr := file.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Read comm first to pre-filter
		commPath := filepath.Join("/proc", pidStr, "comm")
		commBytes, err := c.fs.ReadFile(commPath)
		if err != nil {
			continue
		}
		procName := strings.TrimSpace(string(commBytes))

		isCandidate := false
		for _, cand := range candidates {
			if strings.Contains(procName, cand) {
				isCandidate = true
				break
			}
		}
		// Fallback: check cmdline for node-based CLIs (e.g. Gemini CLI reports comm as "MainThread")
		if !isCandidate {
			cmdlinePath := filepath.Join("/proc", pidStr, "cmdline")
			if cmdlineBytes, err := c.fs.ReadFile(cmdlinePath); err == nil {
				cmdline := string(cmdlineBytes)
				for _, cand := range candidates {
					if strings.Contains(cmdline, "/"+cand) || strings.Contains(cmdline, cand) {
						isCandidate = true
						break
					}
				}
			}
		}
		if !isCandidate {
			continue
		}

		fdDir := filepath.Join("/proc", pidStr, "fd")
		fds, err := c.fs.ReadDir(fdDir)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			fdPath := filepath.Join(fdDir, fd.Name())
			target, err := c.fs.Readlink(fdPath)
			if err != nil {
				continue
			}

			if (strings.Contains(target, "/.gemini/antigravity-cli/conversations/") || strings.Contains(target, "/.gemini/antigravity/conversations/")) && strings.HasSuffix(target, ".db") {
				filename := filepath.Base(target)
				convID := strings.TrimSuffix(filename, ".db")
				active[convID] = pid
			}
		}
	}
	return active
}

func getHarnessFromPrefix(name string) string {
	if strings.HasPrefix(name, "emagy-") {
		return "agy"
	}
	if strings.HasPrefix(name, "emgem-") {
		return "gemini"
	}
	if strings.HasPrefix(name, "emcld-") {
		return "claude"
	}
	return "unknown"
}

func (c *ClassificationEngine) isExcluded(id, folder string) bool {
	excludesEnv := os.Getenv("EXCLUDE_RESUSCITATE")
	var patterns []string
	if excludesEnv != "" {
		// Split by comma, colon, or spaces
		patterns = strings.FieldsFunc(excludesEnv, func(r rune) bool {
			return r == ',' || r == ':' || r == ' '
		})
	} else {
		// Fallback defaults
		patterns = []string{"Chumbia", "chumbawamba"}
	}

	idLower := strings.ToLower(id)
	folderLower := strings.ToLower(folder)

	for _, pattern := range patterns {
		patternLower := strings.ToLower(strings.TrimSpace(pattern))
		if patternLower == "" {
			continue
		}
		// Clean wildcards for substring check
		cleanPattern := strings.ReplaceAll(patternLower, "*", "")
		if cleanPattern == "" {
			continue
		}
		if strings.Contains(idLower, cleanPattern) || strings.Contains(folderLower, cleanPattern) {
			return true
		}
	}
	return false
}

func (c *ClassificationEngine) getTranscriptDescription(sessionID string) string {
	trimmedID := sessionID
	if strings.HasPrefix(sessionID, "emagy-") {
		trimmedID = strings.TrimPrefix(sessionID, "emagy-")
	} else if strings.HasPrefix(sessionID, "emgem-") {
		trimmedID = strings.TrimPrefix(sessionID, "emgem-")
	} else if strings.HasPrefix(sessionID, "emcld-") {
		trimmedID = strings.TrimPrefix(sessionID, "emcld-")
	}

	paths := []string{
		filepath.Join(c.homeDir, ".gemini/antigravity-cli/brain", trimmedID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity-cli/brain", trimmedID, ".system_generated/logs/transcript_full.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript_full.jsonl"),
		// AG2UI paths
		filepath.Join(c.homeDir, ".gemini/antigravity/brain", trimmedID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity/brain", trimmedID, ".system_generated/logs/transcript_full.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity/brain", sessionID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity/brain", sessionID, ".system_generated/logs/transcript_full.jsonl"),
	}

	for _, p := range paths {
		data, err := c.fs.ReadFile(p)
		if err != nil {
			continue
		}
		lines := strings.SplitN(string(data), "\n", 2)
		if len(lines) == 0 || lines[0] == "" {
			continue
		}
		var step struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(lines[0]), &step); err == nil {
			if step.Type == "USER_INPUT" && step.Content != "" {
				content := step.Content
				if startIdx := strings.Index(content, "<USER_REQUEST>"); startIdx != -1 {
					content = content[startIdx+len("<USER_REQUEST>"):]
					if endIdx := strings.Index(content, "</USER_REQUEST>"); endIdx != -1 {
						content = content[:endIdx]
					}
				}
				content = strings.TrimSpace(content)
				content = strings.ReplaceAll(content, "\n", " ")
				if len(content) > 120 {
					content = content[:117] + "..."
				}
				return content
			}
		}
	}
	return ""
}

// getTranscriptWorkspace extracts the workspace folder from an AG2UI transcript.
// Returns (folder, worktreeBranch). For worktrees under .gemini/antigravity/worktrees/,
// folder is the base project path (e.g. ~/git/banche-pvt) and worktreeBranch is the track name.
func (c *ClassificationEngine) getTranscriptWorkspace(sessionID string) (string, string) {
	transcriptPaths := []string{
		filepath.Join(c.homeDir, ".gemini/antigravity/brain", sessionID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(c.homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript.jsonl"),
	}

	homeEscaped := regexp.QuoteMeta(c.homeDir)
	worktreeRe := regexp.MustCompile(homeEscaped + `/.gemini/antigravity/worktrees/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)`)
	workspaceRe := regexp.MustCompile(homeEscaped + `/([a-zA-Z0-9_][a-zA-Z0-9_.-]{2,}(?:/[a-zA-Z0-9_][a-zA-Z0-9_.-]{2,})?)(?:[/"\\s]|$)`)

	for _, p := range transcriptPaths {
		data, err := c.fs.ReadFile(p)
		if err != nil {
			continue
		}

		lines := strings.SplitN(string(data), "\n", 50)
		worktreeCounts := make(map[string]int)
		workspaceCounts := make(map[string]int)

		for _, line := range lines {
			for _, m := range worktreeRe.FindAllStringSubmatch(line, -1) {
				worktreeCounts[m[1]+"/"+m[2]]++
			}
			for _, m := range workspaceRe.FindAllStringSubmatch(line, -1) {
				rel := m[1]
				if strings.HasPrefix(rel, ".gemini") || strings.HasPrefix(rel, ".config") ||
					strings.HasPrefix(rel, ".hermes") || strings.HasPrefix(rel, ".local") {
					continue
				}
				parts := strings.SplitN(rel, "/", 3)
				var key string
				if len(parts) >= 2 {
					key = parts[0] + "/" + parts[1]
				} else {
					key = parts[0]
				}
				workspaceCounts[key]++
			}
		}

		if len(worktreeCounts) > 0 {
			var bestKey string
			var bestCount int
			for k, v := range worktreeCounts {
				if v > bestCount {
					bestKey = k
					bestCount = v
				}
			}
			if bestKey != "" {
				parts := strings.SplitN(bestKey, "/", 2)
				project, track := parts[0], parts[1]
				folder := ""
				for _, cand := range []string{
					filepath.Join(c.homeDir, "git", project),
					filepath.Join(c.homeDir, project),
				} {
					if _, err := c.fs.Stat(cand); err == nil {
						folder = cand
						break
					}
				}
				return folder, track
			}
		}

		if len(workspaceCounts) > 0 {
			var bestKey string
			var bestCount int
			for k, v := range workspaceCounts {
				if v > bestCount {
					bestKey = k
					bestCount = v
				}
			}
			if bestKey != "" {
				return filepath.Join(c.homeDir, bestKey), ""
			}
		}
	}

	return "", ""
}
