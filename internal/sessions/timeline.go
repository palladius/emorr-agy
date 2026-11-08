package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/palladius/emorr-agy/internal/color"
)

// crop truncates a string to maxLen, removing newlines. If maxLen <= 0, no truncation is applied.
func crop(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if maxLen > 0 && len(s) > maxLen {
		if maxLen > 3 {
			return s[:maxLen-3] + "..."
		}
		return s[:maxLen]
	}
	return s
}

// formatDuration formats a duration into a human-readable string (e.g., "12m 34s", "1h 5m").
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// FindTranscriptPath locates the transcript file for a given session ID or prefix.
// It checks both ~/.gemini/antigravity-cli/brain and ~/.gemini/antigravity/brain.
func FindTranscriptPath(homeDir, sessionID string) (string, string, error) {
	trimmedID := sessionID
	for _, prefix := range []string{"emagy-", "emgem-", "emcld-"} {
		if strings.HasPrefix(trimmedID, prefix) {
			trimmedID = strings.TrimPrefix(trimmedID, prefix)
			break
		}
	}

	searchDirs := []string{
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain"),
		filepath.Join(homeDir, ".gemini/antigravity/brain"),
	}

	for _, baseDir := range searchDirs {
		// 1. Direct match
		exactPath := filepath.Join(baseDir, trimmedID, ".system_generated/logs/transcript.jsonl")
		if _, err := os.Stat(exactPath); err == nil {
			return exactPath, trimmedID, nil
		}
		exactFull := filepath.Join(baseDir, trimmedID, ".system_generated/logs/transcript_full.jsonl")
		if _, err := os.Stat(exactFull); err == nil {
			return exactFull, trimmedID, nil
		}

		// 2. Prefix scan
		entries, err := os.ReadDir(baseDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() && strings.HasPrefix(e.Name(), trimmedID) {
					tPath := filepath.Join(baseDir, e.Name(), ".system_generated/logs/transcript.jsonl")
					if _, err := os.Stat(tPath); err == nil {
						return tPath, e.Name(), nil
					}
					tFull := filepath.Join(baseDir, e.Name(), ".system_generated/logs/transcript_full.jsonl")
					if _, err := os.Stat(tFull); err == nil {
						return tFull, e.Name(), nil
					}
				}
			}
		}
	}

	return "", trimmedID, fmt.Errorf("no transcript found for session %q", sessionID)
}

// ShowTimeline parses a transcript.jsonl file and renders a chronological timeline to w.
func ShowTimeline(w io.Writer, transcriptPath, sessionUUID, homeDir string) error {
	steps, err := readTranscriptSteps(transcriptPath)
	if err != nil {
		return fmt.Errorf("failed to read transcript %s: %w", transcriptPath, err)
	}

	if len(steps) == 0 {
		fmt.Fprintln(w, "(transcript is empty)")
		return nil
	}

	var lastPrintedDate string
	var startTime, endTime time.Time

	for i, step := range steps {
		var stepTime time.Time
		if step.CreatedAt != "" {
			t, err := time.Parse(time.RFC3339Nano, step.CreatedAt)
			if err != nil {
				t, err = time.Parse(time.RFC3339, step.CreatedAt)
			}
			if err == nil {
				stepTime = t.Local()
			}
		}

		if stepTime.IsZero() {
			stepTime = time.Now()
		}

		if i == 0 || startTime.IsZero() {
			startTime = stepTime
		}
		endTime = stepTime

		currentDate := stepTime.Format("2006-01-02")
		ts := stepTime.Format("15:04")

		if currentDate != lastPrintedDate {
			fmt.Fprintf(w, "\n📅 \033[1;36m%s\033[0m :\n", currentDate)
			lastPrintedDate = currentDate
		}

		switch step.Type {
		case "USER_INPUT":
			prompt := cleanUserContent(step.Content)
			fmt.Fprintf(w, "  \033[1;32m%s\033[0m 👤 \033[1;32m%s\033[0m\n", ts, crop(prompt, 120))

		case "PLANNER_RESPONSE":
			if content := strings.TrimSpace(step.Content); content != "" {
				fmt.Fprintf(w, "  \033[1;35m%s\033[0m 🚀 %s\n", ts, crop(content, 120))
			}

			for _, tc := range step.ToolCalls {
				name := tc.Name
				var argsMap map[string]interface{}
				_ = json.Unmarshal(tc.Args, &argsMap)

				switch name {
				case "run_command":
					cmd := ""
					desc := ""
					if v, ok := argsMap["CommandLine"].(string); ok {
						cmd = v
					} else if v, ok := argsMap["command"].(string); ok {
						cmd = v
					}
					if v, ok := argsMap["toolAction"].(string); ok {
						desc = v
					} else if v, ok := argsMap["description"].(string); ok {
						desc = v
					} else if v, ok := argsMap["toolSummary"].(string); ok {
						desc = v
					}
					display := cmd
					if desc != "" && desc != cmd {
						display = fmt.Sprintf("%s # %s", cmd, desc)
					}
					fmt.Fprintf(w, "  \033[1;35m%s\033[0m 🚀 🐚 \033[1;32m%s\033[0m\n", ts, crop(display, 100))

				case "view_file", "read_file":
					file := ""
					if v, ok := argsMap["AbsolutePath"].(string); ok {
						file = v
					} else if v, ok := argsMap["file_path"].(string); ok {
						file = v
					} else if v, ok := argsMap["Path"].(string); ok {
						file = v
					} else if v, ok := argsMap["path"].(string); ok {
						file = v
					}
					fmt.Fprintf(w, "  \033[1;35m%s\033[0m 🚀 👀 🛠️  \033[1mview_file\033[0m(\033[1;33m%s\033[0m)\n", ts, crop(file, 55))

				default:
					var pairs []string
					for k, v := range argsMap {
						pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
					}
					argsStr := strings.Join(pairs, ", ")
					fmt.Fprintf(w, "  \033[1;35m%s\033[0m 🚀 🛠️  \033[1m%s\033[0m(%s)\n", ts, name, crop(argsStr, 55))
				}
			}
		}
	}

	// Stats section
	fmt.Fprintln(w, strings.Repeat("─", 80))
	var duration time.Duration
	if !endTime.Before(startTime) {
		duration = endTime.Sub(startTime)
	}
	fmt.Fprintln(w, "📊 Stats della Sessione:")
	fmt.Fprintf(w, "   ⏱️  Duration:  \033[1;32m%s\033[0m\n", formatDuration(duration))
	fmt.Fprintf(w, "   🏁 Start:     %s\n", startTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "   🛑 End:       %s\n", endTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "   🔢 Steps:     %d\n", len(steps))
	fmt.Fprintln(w)

	// Resume hints
	repoName := ExtractRepoName(transcriptPath)
	shortUUID := sessionUUID
	if len(shortUUID) > 8 {
		shortUUID = shortUUID[:8]
	}

	if repoName != "" && repoName != "(system)" {
		fmt.Fprintf(w, "💡 Resume with: \033[1;33mcd ~/git/%s && agy --conversation %s\033[0m\n", repoName, sessionUUID)
		if os.Getenv("TMUX") != "" {
			fmt.Fprintf(w, "💡 Tmux with:   \033[1;33mtmux new-window -n \"agysb-%s\" -c \"%s/git/%s\" \"agy --conversation %s\"\033[0m\n", shortUUID, homeDir, repoName, sessionUUID)
		} else {
			fmt.Fprintf(w, "💡 Tmux with:   \033[1;33mtmux new-session -s \"agysb-%s\" -c \"%s/git/%s\" \"agy --conversation %s\"\033[0m\n", shortUUID, homeDir, repoName, sessionUUID)
		}
	} else {
		fmt.Fprintf(w, "💡 Resume with: \033[1;33magy --conversation %s\033[0m\n", sessionUUID)
		if os.Getenv("TMUX") != "" {
			fmt.Fprintf(w, "💡 Tmux with:   \033[1;33mtmux new-window -n \"agysb-%s\" \"agy --conversation %s\"\033[0m\n", shortUUID, sessionUUID)
		} else {
			fmt.Fprintf(w, "💡 Tmux with:   \033[1;33mtmux new-session -s \"agysb-%s\" \"agy --conversation %s\"\033[0m\n", shortUUID, sessionUUID)
		}
	}
	fmt.Fprintln(w)

	_ = color.Reset // reference color to prevent unused import warnings if any
	return nil
}
