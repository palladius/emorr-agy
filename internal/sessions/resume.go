package sessions

import (
	"bytes"
	"encoding/json"
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
// If it is dead/resuscitatable, it spawns a new tmux session and runs the harness, then attaches.
func ResumeSession(engine *ClassificationEngine, sessionID string) error {
	sessions, err := engine.Classify(nil)
	if err != nil {
		return fmt.Errorf("failed to classify sessions: %w", err)
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
		return fmt.Errorf("session %q not found", sessionID)
	}

	if target.State == StateOpenTmux || target.State == StateOpenPrivate {
		fmt.Printf("Session %q is already running in tmux. Attaching...\n", target.ID)
		return attachTmux(target.ID)
	}

	// Resuscitate in background
	if err := ResuscitateSession(engine, sessionID); err != nil {
		return err
	}

	// Build the tmux session name to attach to
	rawID := target.ID
	if strings.HasPrefix(rawID, "emagy-") {
		rawID = strings.TrimPrefix(rawID, "emagy-")
	} else if strings.HasPrefix(rawID, "emgem-") {
		rawID = strings.TrimPrefix(rawID, "emgem-")
	} else if strings.HasPrefix(rawID, "emcld-") {
		rawID = strings.TrimPrefix(rawID, "emcld-")
	}

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

	return attachTmux(tmuxSessionName)
}

// ResuscitateSession starts a dead session in the background without attaching to it.
func ResuscitateSession(engine *ClassificationEngine, sessionID string) error {
	sessions, err := engine.Classify(nil)
	if err != nil {
		return fmt.Errorf("failed to classify sessions: %w", err)
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
		return fmt.Errorf("session %q not found", sessionID)
	}

	if target.State == StateOpenTmux || target.State == StateOpenPrivate {
		return nil // Already running
	}

	rawID := target.ID
	if strings.HasPrefix(rawID, "emagy-") {
		rawID = strings.TrimPrefix(rawID, "emagy-")
	} else if strings.HasPrefix(rawID, "emgem-") {
		rawID = strings.TrimPrefix(rawID, "emgem-")
	} else if strings.HasPrefix(rawID, "emcld-") {
		rawID = strings.TrimPrefix(rawID, "emcld-")
	}

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

	cmd := execCommand("tmux", "new-session", "-d", "-s", tmuxSessionName, "-c", folder, harnessCmd)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start new tmux session: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// ArchiveSession logically archives a session by writing/updating its local cache JSON file.
func ArchiveSession(engine *ClassificationEngine, sessionID string) error {
	rawID := sessionID
	if strings.HasPrefix(rawID, "emagy-") {
		rawID = strings.TrimPrefix(rawID, "emagy-")
	} else if strings.HasPrefix(rawID, "emgem-") {
		rawID = strings.TrimPrefix(rawID, "emgem-")
	} else if strings.HasPrefix(rawID, "emcld-") {
		rawID = strings.TrimPrefix(rawID, "emcld-")
	}

	cacheDir := filepath.Join(engine.homeDir, ".emorr-agy/cache")
	_ = engine.fs.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, rawID+".json")

	var existing struct {
		About            string `json:"about"`
		UserInputPending bool   `json:"user_input_pending"`
		WorthResuscitate bool   `json:"worth_resuscitate"`
	}
	existing.WorthResuscitate = false

	if data, err := engine.fs.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(data, &existing)
		existing.WorthResuscitate = false
	}

	data, err := json.Marshal(existing)
	if err != nil {
		return err
	}

	return engine.fs.WriteFile(cachePath, data, 0644)
}

func attachTmux(sessionName string) error {
	binary, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux command not found: %w", err)
	}

	args := []string{"tmux", "attach-session", "-t", sessionName}
	env := os.Environ()

	return syscallExec(binary, args, env)
}
