// Package remoteagy implements remote Antigravity CLI (agy) session management
// via Telegram. It handles the concierge menu, session lifecycle, output
// polling, input relay, and ask_question detection for remotely controlling
// agy instances from a phone.
package remoteagy

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// MaxConcurrentSessions is the hard limit on concurrent remote agy sessions.
const MaxConcurrentSessions = 3

// DefaultConfigPath is the default location for the folder whitelist YAML.
const DefaultConfigPath = "~/.emorr-agy-folders.yaml"

// ConfigEnvVar is the environment variable to override the config path.
const ConfigEnvVar = "EMORR_AGY_FOLDERS_YAML"

// RemoteSessionPrefix is the tmux session name prefix for Telegram-spawned sessions.
const RemoteSessionPrefix = "remote-agy-"

// --- YAML Config Types ---

// Config represents the top-level YAML configuration for remote agy sessions.
type Config struct {
	DefaultSandbox *bool          `yaml:"default_sandbox"`
	Folders        []FolderConfig `yaml:"folders"`
}

// FolderConfig represents a single whitelisted folder entry in the YAML config.
type FolderConfig struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	AgyFlags string `yaml:"agy_flags"`
	Sandbox  *bool  `yaml:"sandbox"`
}

// IsSandboxed returns true if this folder's agy session should be sandboxed.
// Per-folder sandbox overrides the default; if neither is set, defaults to true.
func (f FolderConfig) IsSandboxed(defaultSandbox *bool) bool {
	if f.Sandbox != nil {
		return *f.Sandbox
	}
	if defaultSandbox != nil {
		return *defaultSandbox
	}
	return true // safe default
}

// --- Session State ---

// SessionState represents the current state of a remote agy session.
type SessionState int

const (
	// SessionRunning means the agy process is active in tmux.
	SessionRunning SessionState = iota
	// SessionExited means the agy process has exited.
	SessionExited
)

// RemoteSession tracks a single remote agy session.
type RemoteSession struct {
	Name      string       // tmux session name (e.g., "remote-agy-gic")
	Folder    FolderConfig // the folder config used to create this session
	Marker    string       // unique marker injected for transcript discovery
	State     SessionState
	StartedAt time.Time
	LastPane  string // last captured pane content for diffing
}

// --- Concierge State ---

// ConciergeMode represents the modal navigation state.
type ConciergeMode int

const (
	// ModeConcierge is the main menu state.
	ModeConcierge ConciergeMode = iota
	// ModeAttached means the user is attached to a specific session.
	ModeAttached
)

// Concierge manages the state machine for remote agy interaction.
type Concierge struct {
	mu              sync.Mutex
	Mode            ConciergeMode
	AttachedSession string // name of the currently attached session
	Sessions        map[string]*RemoteSession
	Config          *Config
	PollingPaused   bool
}

// NewConcierge creates a new Concierge with the given config.
func NewConcierge(cfg *Config) *Concierge {
	return &Concierge{
		Mode:     ModeConcierge,
		Sessions: make(map[string]*RemoteSession),
		Config:   cfg,
	}
}

// GetState returns the current concierge mode and attached session name (thread-safe).
func (c *Concierge) GetState() (ConciergeMode, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Mode, c.AttachedSession
}

// ActiveSessionCount returns the number of active sessions (thread-safe).
func (c *Concierge) ActiveSessionCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.Sessions)
}

// --- Config Loading ---

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot expand ~: %w", err)
	}
	return filepath.Join(home, path[1:]), nil
}

// LoadConfig loads the folder whitelist YAML config from the default path
// or from the path specified by the EMORR_AGY_FOLDERS_YAML env var.
func LoadConfig() (*Config, error) {
	configPath := os.Getenv(ConfigEnvVar)
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	expandedPath, err := expandHome(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", expandedPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate and expand paths
	seen := make(map[string]bool)
	for i := range cfg.Folders {
		f := &cfg.Folders[i]
		if f.Name == "" {
			return nil, fmt.Errorf("folder entry %d has empty name", i)
		}
		if seen[f.Name] {
			return nil, fmt.Errorf("duplicate folder name: %q", f.Name)
		}
		seen[f.Name] = true

		expanded, err := expandHome(f.Path)
		if err != nil {
			return nil, fmt.Errorf("folder %q: %w", f.Name, err)
		}
		f.Path = expanded
	}

	return &cfg, nil
}

// --- Session Management ---

// GenerateMarker creates a unique marker string for transcript discovery.
func GenerateMarker(folderName string) string {
	data := fmt.Sprintf("%s-%d", folderName, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("emorr-agy-marker-%x", hash)
}

// SanitizeSessionName creates a valid tmux session name from a folder name.
func SanitizeSessionName(folderName string) string {
	// Remove emoji and special chars, lowercase, replace spaces with dashes
	name := strings.ToLower(folderName)
	var clean []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean = append(clean, r)
		} else if r == ' ' {
			clean = append(clean, '-')
		}
	}
	result := string(clean)
	// Remove leading/trailing dashes
	result = strings.Trim(result, "-_")
	if result == "" {
		result = "unnamed"
	}
	return RemoteSessionPrefix + result
}

// TmuxSessionExists checks if a tmux session with the given name exists.
func TmuxSessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// ListRemoteSessions discovers existing remote-agy-* tmux sessions.
func ListRemoteSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// tmux returns error if no sessions exist — that's fine
		return nil, nil
	}

	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.HasPrefix(line, RemoteSessionPrefix) {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

// IsRemoteSession returns true if the session name has the remote-agy- prefix.
func IsRemoteSession(name string) bool {
	return strings.HasPrefix(name, RemoteSessionPrefix)
}

// BuildAgyCommand constructs the full agy launch command for a folder config.
func BuildAgyCommand(folder FolderConfig, marker string, defaultSandbox *bool) string {
	parts := []string{"agy"}

	// Sandbox enforcement: --sandbox enables OS-level terminal sandboxing
	// which prevents the agent from escaping the git repo boundary.
	if folder.IsSandboxed(defaultSandbox) {
		parts = append(parts, "--sandbox")
	}

	if folder.AgyFlags != "" {
		parts = append(parts, folder.AgyFlags)
	}

	return strings.Join(parts, " ")
}

// StartRemoteSession creates a new tmux session and launches agy in the given folder.
func (c *Concierge) StartRemoteSession(folder FolderConfig) (*RemoteSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.Sessions) >= MaxConcurrentSessions {
		return nil, fmt.Errorf("maximum %d concurrent sessions reached; kill one first", MaxConcurrentSessions)
	}

	sessionName := SanitizeSessionName(folder.Name)

	// Check if session already exists
	if TmuxSessionExists(sessionName) {
		return nil, fmt.Errorf("session %q already exists", sessionName)
	}

	marker := GenerateMarker(folder.Name)
	agyCmd := BuildAgyCommand(folder, marker, c.Config.DefaultSandbox)

	// Create tmux session: cd to folder, then launch agy
	fullCmd := fmt.Sprintf("cd %s && %s", shellescape(folder.Path), agyCmd)
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, fullCmd)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create tmux session %q: %w", sessionName, err)
	}

	session := &RemoteSession{
		Name:      sessionName,
		Folder:    folder,
		Marker:    marker,
		State:     SessionRunning,
		StartedAt: time.Now(),
	}
	c.Sessions[sessionName] = session

	return session, nil
}

// KillRemoteSession terminates a remote agy tmux session.
func (c *Concierge) KillRemoteSession(sessionName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %q: %w", sessionName, err)
	}

	delete(c.Sessions, sessionName)

	// If we were attached to this session, go back to concierge
	if c.AttachedSession == sessionName {
		c.Mode = ModeConcierge
		c.AttachedSession = ""
	}

	return nil
}

// AttachSession switches the user to the specified session.
func (c *Concierge) AttachSession(sessionName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !TmuxSessionExists(sessionName) {
		return fmt.Errorf("session %q does not exist", sessionName)
	}

	c.Mode = ModeAttached
	c.AttachedSession = sessionName
	c.PollingPaused = false
	return nil
}

// DetachSession returns the user to the concierge menu.
func (c *Concierge) DetachSession() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Mode = ModeConcierge
	c.AttachedSession = ""
}

// SendInput sends text input to the attached tmux session.
func (c *Concierge) SendInput(text string) error {
	c.mu.Lock()
	name := c.AttachedSession
	c.mu.Unlock()

	if name == "" {
		return fmt.Errorf("not attached to any session")
	}

	cmd := exec.Command("tmux", "send-keys", "-t", name, text, "Enter")
	return cmd.Run()
}

// CapturePane captures the current content of the attached tmux session's pane.
func CapturePane(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p", "-S", "-100")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane for %q: %w", sessionName, err)
	}
	return string(output), nil
}

// shellescape wraps a string in single quotes for safe shell usage.
func shellescape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// --- Message Splitting ---

// EmojiNumbers maps page numbers to emoji digits.
var EmojiNumbers = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣", "9️⃣"}

// SplitMessage splits a long message into Telegram-safe chunks (≤4096 chars)
// with emoji page numbers and ✂️ continuation markers.
func SplitMessage(text string, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = 4096
	}

	if len(text) <= maxLen {
		return []string{text}
	}

	// Reserve space for page header and continuation marker
	// e.g., "1️⃣ " (prefix) + " ✂️" (suffix) ≈ 20 bytes
	effectiveMax := maxLen - 20

	var chunks []string
	for len(text) > 0 {
		end := effectiveMax
		if end > len(text) {
			end = len(text)
		}

		// Try to split at a newline boundary
		if end < len(text) {
			if idx := strings.LastIndex(text[:end], "\n"); idx > effectiveMax/2 {
				end = idx + 1
			}
		}

		chunks = append(chunks, text[:end])
		text = text[end:]
	}

	// Add emoji pagination
	total := len(chunks)
	for i := range chunks {
		pageNum := ""
		if i < len(EmojiNumbers) {
			pageNum = EmojiNumbers[i]
		} else {
			pageNum = fmt.Sprintf("[%d]", i+1)
		}

		if total > 1 {
			if i < total-1 {
				chunks[i] = fmt.Sprintf("%s/%d\n%s\n✂️", pageNum, total, chunks[i])
			} else {
				chunks[i] = fmt.Sprintf("%s/%d\n%s", pageNum, total, chunks[i])
			}
		}
	}

	return chunks
}
