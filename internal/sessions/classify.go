package sessions

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type SessionState string

const (
	StateOpenTmux           SessionState = "open_tmux"
	StateOpenPrivate        SessionState = "open_private"
	StateDeadResuscitatable SessionState = "dead_resuscitatable"
	StateDeadArchived       SessionState = "dead_archived"
)

type Session struct {
	ID            string       `json:"id"`
	Harness       string       `json:"harness"` // "agy", "gemini", "claude", "unknown"
	State         SessionState `json:"state"`
	Folder        string       `json:"folder"`
	Title         string       `json:"title,omitempty"`
	Description   string       `json:"description,omitempty"`
	LastActivity  time.Time    `json:"last_activity,omitempty"`
	ProcessCount  int          `json:"process_count"`
	ResumeCommand string       `json:"resume_command,omitempty"`
}

type TmuxSession struct {
	Name     string
	Path     string
	Attached bool
	Windows  int
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
			attached := parts[2] == "1"
			windows, _ := strconv.Atoi(parts[3])
			list = append(list, TmuxSession{
				Name:     parts[0],
				Path:     parts[1],
				Attached: attached,
				Windows:  windows,
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
				// E.g., emagy-session-12345 -> session-12345 (or just take the rest)
				// Wait, let's keep the ID as the suffix or the full name? Let's check the test expectation:
				// The test expected: `s.ID == "session-12345" || s.ID == "emagy-session-12345"`
				// Let's extract ID from the prefix suffix, but if it has a prefix, let's match both.
				// Actually, we can keep the full tmux name, but if we map from historical map we should correlate.
				// Let's extract the conversation ID suffix if it starts with one of our prefixes.
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

			s := Session{
				ID:            ts.Name, // Keep name as ID or extracted ID? Let's keep ts.Name as ID but populate other fields
				Harness:       harness,
				State:         state,
				Folder:        ts.Path,
				ProcessCount:  ts.Windows, // fallback/rough count
				ResumeCommand: "tmux attach -t " + ts.Name,
			}
			sessions = append(sessions, s)
		}
	}

	// 2. Load historical sessions from cache file
	cacheFile := filepath.Join(c.homeDir, ".gemini/antigravity-cli/cache/last_conversations.json")
	if data, err := c.fs.ReadFile(cacheFile); err == nil {
		var cacheConvs map[string]string
		if err := json.Unmarshal(data, &cacheConvs); err == nil {
			for folder, convID := range cacheConvs {
				// If it's already running in tmux, skip adding a duplicate dead session
				if runningMap[convID] {
					continue
				}

				harness := "agy" // fallback for historical
				state := StateDeadResuscitatable

				// Check exclusions
				if c.isExcluded(convID, folder) {
					state = StateDeadArchived
				}

				s := Session{
					ID:            convID,
					Harness:       harness,
					State:         state,
					Folder:        folder,
					ResumeCommand: "emorr-agy resume " + convID, // fallback resume CLI
				}
				sessions = append(sessions, s)
			}
		}
	}

	// 3. Filter by harness if filter is provided
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
