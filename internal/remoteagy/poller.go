package remoteagy

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Poller watches the attached tmux session for new output and calls
// the provided callback when new content is detected.
type Poller struct {
	mu           sync.Mutex
	sessionName  string
	lastContent  string
	running      bool
	paused       bool
	stopCh       chan struct{}
	interval     time.Duration
	onNewOutput  func(newContent string) // called when new output is detected
	onSessionEnd func()                  // called when tmux session exits
}

// NewPoller creates a new output poller for the given session.
func NewPoller(sessionName string, interval time.Duration, onNewOutput func(string), onSessionEnd func()) *Poller {
	return &Poller{
		sessionName:  sessionName,
		interval:     interval,
		onNewOutput:  onNewOutput,
		onSessionEnd: onSessionEnd,
		stopCh:       make(chan struct{}),
	}
}

// Start begins polling the tmux session for new output.
func (p *Poller) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	go p.pollLoop()
}

// Stop terminates the polling loop.
func (p *Poller) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		close(p.stopCh)
		p.running = false
	}
}

// Pause pauses output delivery without stopping the poller.
func (p *Poller) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = true
}

// Resume resumes output delivery.
func (p *Poller) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = false
}

// IsPaused returns whether the poller is paused.
func (p *Poller) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}

// ForceRefresh captures and sends the current pane content regardless of diff.
func (p *Poller) ForceRefresh() (string, error) {
	content, err := CapturePane(p.sessionName)
	if err != nil {
		return "", err
	}
	p.mu.Lock()
	p.lastContent = content
	p.mu.Unlock()
	return content, nil
}

func (p *Poller) pollLoop() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.poll()
		}
	}
}

func (p *Poller) poll() {
	// Check if session still exists
	if !TmuxSessionExists(p.sessionName) {
		if p.onSessionEnd != nil {
			p.onSessionEnd()
		}
		p.Stop()
		return
	}

	content, err := CapturePane(p.sessionName)
	if err != nil {
		return // silently skip on capture errors
	}

	p.mu.Lock()
	paused := p.paused
	lastContent := p.lastContent
	p.lastContent = content
	p.mu.Unlock()

	if paused {
		return
	}

	// Diff: only send new content
	newContent := diffContent(lastContent, content)
	if newContent != "" && p.onNewOutput != nil {
		p.onNewOutput(newContent)
	}
}

// diffContent returns only the new lines that appear in 'current' but not in 'previous'.
// Uses a simple suffix-matching approach: finds the longest suffix of 'previous'
// that appears as a prefix of 'current', and returns the remainder.
func diffContent(previous, current string) string {
	if previous == "" {
		return strings.TrimSpace(current)
	}
	if previous == current {
		return ""
	}

	// Find the common part: look for the end of previous in current
	prevLines := strings.Split(previous, "\n")
	currLines := strings.Split(current, "\n")

	// Find the longest matching suffix of prevLines in currLines
	matchLen := 0
	for tryLen := len(prevLines); tryLen > 0; tryLen-- {
		suffix := prevLines[len(prevLines)-tryLen:]
		if hasPrefix(currLines, suffix) {
			matchLen = tryLen
			break
		}
	}

	if matchLen > 0 && matchLen < len(currLines) {
		newLines := currLines[matchLen:]
		result := strings.TrimSpace(strings.Join(newLines, "\n"))
		return result
	}

	// No overlap found — return everything
	return strings.TrimSpace(current)
}

// hasPrefix checks if 'lines' starts with 'prefix'.
func hasPrefix(lines, prefix []string) bool {
	if len(prefix) > len(lines) {
		return false
	}
	for i, p := range prefix {
		if strings.TrimRight(lines[i], " ") != strings.TrimRight(p, " ") {
			return false
		}
	}
	return true
}

// --- Session Info ---

// SessionInfo returns a formatted string with session metadata.
func SessionInfo(session *RemoteSession) string {
	uptime := time.Since(session.StartedAt).Round(time.Second)
	sandboxStatus := "🔒 Sandboxed"
	if session.Folder.Sandbox != nil && !*session.Folder.Sandbox {
		sandboxStatus = "⚠️ UNSANDBOXED"
	}

	return fmt.Sprintf(
		"🔍 *Session Info*\n"+
			"├ *Name:* `%s`\n"+
			"├ *Folder:* `%s`\n"+
			"├ *Flags:* `%s`\n"+
			"├ *Sandbox:* %s\n"+
			"├ *Uptime:* %s\n"+
			"└ *Marker:* `%.20s...`",
		session.Name,
		session.Folder.Path,
		session.Folder.AgyFlags,
		sandboxStatus,
		uptime.String(),
		session.Marker,
	)
}
