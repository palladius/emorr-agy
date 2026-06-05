package sessions

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Mockable package-level variables for execution
var execCommand = exec.Command
var syscallExec = func(argv0 string, argv []string, envv []string) error {
	return syscall.Exec(argv0, argv, envv)
}

// ResumeSession attempts to resume a session by its ID.
// If it is already running, it attaches to the existing tmux session.
// If it is dead/resuscitatable, it spawns a new tmux session and runs the harness.
func ResumeSession(engine *ClassificationEngine, sessionID string) error {
	sessions, err := engine.Classify(nil)
	if err != nil {
		return fmt.Errorf("failed to classify sessions: %w", err)
	}

	var target *Session
	for i := range sessions {
		// Match exact ID or prefixed version
		if sessions[i].ID == sessionID ||
			strings.TrimPrefix(sessions[i].ID, "emagy-") == sessionID ||
			strings.TrimPrefix(sessions[i].ID, "emgem-") == sessionID ||
			strings.TrimPrefix(sessions[i].ID, "emcld-") == sessionID {
			target = &sessions[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("session %q not found", sessionID)
	}

	// 1. If session is already active/running in tmux, attach to it.
	if target.State == StateOpenTmux || target.State == StateOpenPrivate {
		fmt.Printf("Session %q is already running in tmux. Attaching...\n", target.ID)
		return attachTmux(target.ID)
	}

	// 2. Otherwise, start a new tmux session and attach to it.
	// Clean the ID to get the raw conversation ID (without prefix)
	rawID := target.ID
	if strings.HasPrefix(rawID, "emagy-") {
		rawID = strings.TrimPrefix(rawID, "emagy-")
	} else if strings.HasPrefix(rawID, "emgem-") {
		rawID = strings.TrimPrefix(rawID, "emgem-")
	} else if strings.HasPrefix(rawID, "emcld-") {
		rawID = strings.TrimPrefix(rawID, "emcld-")
	}

	// Build the tmux session name
	tmuxSessionName := target.ID
	if !strings.HasPrefix(tmuxSessionName, "emagy-") && !strings.HasPrefix(tmuxSessionName, "emgem-") && !strings.HasPrefix(tmuxSessionName, "emcld-") {
		prefix := "emagy-"
		if target.Harness == "gemini" {
			prefix = "emgem-"
		} else if target.Harness == "claude" {
			prefix = "emcld-"
		}
		tmuxSessionName = prefix + rawID
	}

	// Resolve harness command
	var harnessCmd string
	switch target.Harness {
	case "gemini":
		harnessCmd = "gemini --resume " + rawID
	case "claude":
		harnessCmd = "claude --resume " + rawID
	default:
		harnessCmd = "agy --conversation " + rawID
	}

	folder := target.Folder
	if strings.HasPrefix(folder, "~") {
		folder = filepath.Join(engine.homeDir, strings.TrimPrefix(folder, "~"))
	}

	fmt.Printf("Resuscitating session %s in folder %s...\n", rawID, folder)
	fmt.Printf("Command: tmux new-session -d -s %s -c %s '%s'\n", tmuxSessionName, folder, harnessCmd)

	// Create new tmux session in background
	cmd := execCommand("tmux", "new-session", "-d", "-s", tmuxSessionName, "-c", folder, harnessCmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start new tmux session: %w (stderr: %s)", err, stderr.String())
	}

	// Attach to it
	return attachTmux(tmuxSessionName)
}

func attachTmux(sessionName string) error {
	binary, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux command not found: %w", err)
	}

	args := []string{"tmux", "attach-session", "-t", sessionName}
	env := os.Environ()

	// syscallExec replaces the current process with tmux
	return syscallExec(binary, args, env)
}
