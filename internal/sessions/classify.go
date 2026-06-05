package sessions

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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

	// 2. Load historical and active non-tmux sessions from cache file
	cacheFile := filepath.Join(c.homeDir, ".gemini/antigravity-cli/cache/last_conversations.json")
	if data, err := c.fs.ReadFile(cacheFile); err == nil {
		var cacheConvs map[string]string
		if err := json.Unmarshal(data, &cacheConvs); err == nil {
			for folder, convID := range cacheConvs {
				// If it's already running in tmux, skip adding a duplicate session
				if runningMap[convID] {
					continue
				}

				harness := "agy" // fallback for historical
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

				// If it's active in background processes but not in tmux, it's StateOpenAgy (emoji 🟢)
				if _, active := activeConvs[convID]; active {
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
					if fi, err := c.fs.Stat(folder); err == nil {
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

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		pidStr := file.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
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

			if strings.Contains(target, "/.gemini/antigravity-cli/conversations/") && strings.HasSuffix(target, ".db") {
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
