package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/palladius/emorr-agy/internal/color"
)

type ProcessStatus struct {
	PID          string
	Harness      string
	SessionID    string
	CWD          string
	Status       string
	Emoji        string
	Detail       string
	LastActivity time.Time
}

type TranscriptStep struct {
	StepIndex int             `json:"step_index"`
	Source    string          `json:"source"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	Content   string          `json:"content"`
	ToolCalls []TranscriptTool `json:"tool_calls"`
}

type TranscriptTool struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// PrintProcessStatus retrieves all active sessions and writes a detailed process status list to w.
func PrintProcessStatus(w io.Writer, engine *ClassificationEngine) error {
	activeConvs := engine.FindActiveConvs()
	sessions, err := engine.Classify(nil)
	if err != nil {
		return err
	}

	var activeSessions []Session
	for _, s := range sessions {
		if s.State == StateOpenTmux || s.State == StateOpenAgy || s.State == StateOpenPrivate {
			activeSessions = append(activeSessions, s)
		}
	}

	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
		color.Colorize("PID", color.Plain),
		color.Colorize("ST", color.Plain),
		color.Colorize("HARNESS", color.Plain),
		color.Colorize("CWD", color.Plain),
		color.Colorize("SESSION ID", color.Plain),
		color.Colorize("STATUS / DETAIL", color.Plain),
	)

	for _, s := range activeSessions {
		// Try to resolve PID from activeConvs using trimmed ID or full ID
		trimmedID := s.ID
		for _, prefix := range []string{"emagy-", "emgem-", "emcld-"} {
			if strings.HasPrefix(trimmedID, prefix) {
				trimmedID = strings.TrimPrefix(trimmedID, prefix)
				break
			}
		}

		pid := ""
		if p, ok := activeConvs[trimmedID]; ok {
			pid = strconv.Itoa(p)
		} else if p, ok := activeConvs[s.ID]; ok {
			pid = strconv.Itoa(p)
		}

		cwd := s.Folder
		status := "Active"
		emoji := "🟢"
		detail := ""

		if pid != "" {
			// Read real CWD if possible
			if realCwd, err := os.Readlink(filepath.Join("/proc", pid, "cwd")); err == nil {
				cwd = realCwd
			}

			// Read detailed status from transcript
			st, em, det := GetDetailedStatus(engine.homeDir, trimmedID)
			if st != "" {
				status = st
				emoji = em
				detail = det
			}
		} else {
			pid = "N/A"
			status = "Idle (Shell)"
			emoji = "💤"
		}

		// Shorten CWD for display
		displayCwd := strings.ReplaceAll(cwd, engine.homeDir, "~")
		if len(displayCwd) > 40 {
			displayCwd = displayCwd[:37] + "..."
		}

		statusAndDetail := fmt.Sprintf("%s %s", status, detail)
		if len(statusAndDetail) > 80 {
			statusAndDetail = statusAndDetail[:77] + "..."
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			color.Colorize(fmt.Sprintf("%-6s", pid), color.LightGray),
			color.Colorize(emoji, color.Plain),
			color.Colorize(s.Harness, color.Plain),
			color.Colorize(displayCwd, color.Blue),
			color.Colorize(s.ID, color.BoldWhite),
			color.Colorize(statusAndDetail, color.Cyan),
		)
	}

	return tw.Flush()
}

// GetDetailedStatus parses the last step of the session transcript to determine what the agent is doing.
func GetDetailedStatus(homeDir, sessionID string) (status string, emoji string, detail string) {
	paths := []string{
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript_full.jsonl"),
	}

	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "Running", "🟢", "Agent process is active"
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		return "Running", "🟢", "Agent process is active"
	}

	// Find the last non-empty line
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			lastLine = line
			break
		}
	}

	if lastLine == "" {
		return "Running", "🟢", "Agent process is active"
	}

	var step TranscriptStep
	if err := json.Unmarshal([]byte(lastLine), &step); err != nil {
		return "Running", "🟢", "Agent process is active"
	}

	switch step.Type {
	case "USER_INPUT":
		return "Thinking", "💭", "User request received"

	case "PLANNER_RESPONSE":
		if len(step.ToolCalls) == 0 {
			return "Thinking", "💭", "Planning next steps"
		}

		// Check tool calls
		for _, tc := range step.ToolCalls {
			switch tc.Name {
			case "ask_question":
				var args struct {
					Questions []struct {
						Question string `json:"question"`
					} `json:"questions"`
				}
				qText := "Asking a question"
				if err := json.Unmarshal(tc.Args, &args); err == nil && len(args.Questions) > 0 {
					qText = args.Questions[0].Question
				}
				return "Waiting (Question)", "💬", qText

			case "ask_permission":
				var args struct {
					Action string `json:"Action"`
					Target string `json:"Target"`
				}
				pText := "Requesting permission"
				if err := json.Unmarshal(tc.Args, &args); err == nil {
					pText = fmt.Sprintf("Approve %s: %s", args.Action, args.Target)
				}
				return "Waiting (Permission)", "🔒", pText

			case "run_command":
				var args struct {
					CommandLine string `json:"CommandLine"`
				}
				cmdText := "Running command"
				if err := json.Unmarshal(tc.Args, &args); err == nil {
					cmdText = args.CommandLine
				}
				return "Running Command", "⚙️", cmdText

			default:
				// Return first tool call name
				return "Running Tool", "🔧", tc.Name
			}
		}
		return "Thinking", "💭", "Planning next steps"

	default:
		// If the last step is a tool output, the agent has finished executing it and is generating the next turn.
		return "Thinking", "💭", "Processing tool output"
	}
}

// CleanCommandline replaces null bytes in cmdline content with spaces.
func CleanCommandline(cmdline []byte) string {
	res := strings.ReplaceAll(string(cmdline), "\x00", " ")
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(res, " "))
}
