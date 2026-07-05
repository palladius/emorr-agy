package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/palladius/emorr-agy/internal/color"
)

// folderColor returns Cyan for symlink paths, Blue for normal dirs.
func folderColor(path string) string {
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return color.Cyan
	}
	return color.Blue
}

type ListOptions struct {
	Harness    []string
	Format     string // "short", "long", "json"
	All        bool   // If false, exclude archived sessions
	ActiveOnly bool   // If true, show only active/running sessions
	Folder     string // If set, filter sessions by folder prefix match
}

// ListSessions retrieves and prints the classified sessions to the provided writer.
func ListSessions(w io.Writer, engine *ClassificationEngine, opts ListOptions) error {
	sessions, err := engine.Classify(opts.Harness)
	if err != nil {
		return err
	}

	// Resolve home dir for tilde substitution in folder display
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = engine.homeDir
	}

	if opts.ActiveOnly {
		var filtered []Session
		for _, s := range sessions {
			if s.State == StateOpenTmux || s.State == StateOpenAgy || s.State == StateOpenPrivate {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	} else if !opts.All {
		var filtered []Session
		for _, s := range sessions {
			if s.State != StateDeadArchived {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Apply folder filter if set
	if opts.Folder != "" {
		var filtered []Session
		for _, s := range sessions {
			if IsPathMatch(s.Folder, opts.Folder) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	format := opts.Format
	if format == "" {
		format = "short"
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		_, err = w.Write(data)
		if err != nil {
			return err
		}
		_, _ = w.Write([]byte("\n"))

	case "long":
		tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			color.Colorize("ST T", color.Plain),
			color.Colorize("  SESSION ID", color.Plain),
			color.Colorize("AGE", color.Plain),
			color.Colorize("DIR", color.Plain),
			color.Colorize("HARNESS", color.Plain),
			color.Colorize("WINDOWS", color.Plain),
			color.Colorize("RESUME COMMAND", color.Plain),
			color.Colorize("DESCRIPTION", color.Plain),
		)
		hasPrintedSeparator := false
		for _, s := range sessions {
			if s.State == StateDeadArchived && !hasPrintedSeparator {
				fmt.Fprintln(tw, "---\t---\t---\t---\t---\t---\t---\t---")
				hasPrintedSeparator = true
			}
			emoji := formatStatus(s)
			harnessEmoji := getEmojiForHarness(s.Harness)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, homeDir, "~")
			if folder == "" {
				folder = "-"
			}
			if len(folder) > 30 {
				folder = folder[:27] + "..."
			}

			ageColor := color.LightGray
			if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
				ageColor = color.Green
			} else if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			desc := strings.TrimSpace(s.Description)
			desc = strings.ReplaceAll(desc, "\n", " ")
			if s.Title != "" {
				desc = s.Title + " │ " + desc
			}
			if s.IsCron {
				desc = "🔁 " + desc
			}
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}

			coloredDesc := color.Colorize(desc, color.Cyan)
			if s.WorktreeBranch != "" {
				coloredDesc = "🌳 " + color.Colorize(s.WorktreeBranch, color.Green) + " " + coloredDesc
			}

			statusAndHarness := fmt.Sprintf("%s %s", emoji, harnessEmoji)
			fmt.Fprintf(tw, "%s\t %s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				color.Colorize(statusAndHarness, color.Plain),
				color.Colorize(s.ID, color.BoldWhite),
				color.Colorize(fmt.Sprintf("%3s", age), ageColor),
				color.Colorize(folder, folderColor(s.Folder)),
				color.Colorize(s.Harness, color.Plain),
				color.Colorize(strconv.Itoa(s.ProcessCount), color.Plain),
				color.Colorize(s.ResumeCommand, color.Plain),
				coloredDesc,
			)
		}
		tw.Flush()

	default: // "short" or fallback
		tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			color.Colorize("ST T", color.Plain),
			color.Colorize("  SESSION ID", color.Plain),
			color.Colorize("AGE", color.Plain),
			color.Colorize("DIR", color.Plain),
			color.Colorize("DESCRIPTION", color.Plain),
		)
		hasPrintedSeparator := false
		for _, s := range sessions {
			if s.State == StateDeadArchived && !hasPrintedSeparator {
				fmt.Fprintln(tw, "---\t---\t---\t---\t---")
				hasPrintedSeparator = true
			}
			emoji := formatStatus(s)
			harnessEmoji := getEmojiForHarness(s.Harness)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, homeDir, "~")
			if folder == "" {
				folder = "-"
			}
			if len(folder) > 30 {
				folder = folder[:27] + "..."
			}

			ageColor := color.LightGray
			if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
				ageColor = color.Green
			} else if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			desc := strings.TrimSpace(s.Description)
			desc = strings.ReplaceAll(desc, "\n", " ")
			// Prepend title if present (AG2UI annotation titles)
			if s.Title != "" {
				desc = s.Title + " │ " + desc
			}
			if s.IsCron {
				desc = "🔁 " + desc
			}
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}

			// Build colored description: worktree prefix in green, rest in cyan
			coloredDesc := color.Colorize(desc, color.Cyan)
			if s.WorktreeBranch != "" {
				coloredDesc = "🌳 " + color.Colorize(s.WorktreeBranch, color.Green) + " " + coloredDesc
			}

			statusAndHarness := fmt.Sprintf("%s %s", emoji, harnessEmoji)
			fmt.Fprintf(tw, "%s\t %s\t%s\t%s\t%s\n",
				color.Colorize(statusAndHarness, color.Plain),
				color.Colorize(s.ID, color.BoldWhite),
				color.Colorize(fmt.Sprintf("%3s", age), ageColor),
				color.Colorize(folder, folderColor(s.Folder)),
				coloredDesc,
			)
		}
		tw.Flush()
	}

	return nil
}

// IsPathMatch returns true if dir matches the filter path. An empty filter matches
// everything. The match is prefix-based: /workspace/proj1 matches /workspace/proj1/sub
// but not /workspace/proj10 (full path component matching via trailing separator).
func IsPathMatch(dir, filter string) bool {
	if filter == "" {
		return true
	}
	dir = filepath.Clean(dir)
	filter = filepath.Clean(filter)

	if dir == filter {
		return true
	}

	sep := string(filepath.Separator)
	if !strings.HasSuffix(filter, sep) {
		filter += sep
	}
	return strings.HasPrefix(dir, filter)
}

func getEmojiForHarness(harness string) string {
	switch harness {
	case "gemini":
		return "♊"
	case "agy":
		return "⬆️"
	case "ag2ui":
		return "🖥️"
	case "claude":
		return "🇫🇷"
	default:
		return "❓"
	}
}

func getEmojiForState(state SessionState) string {
	switch state {
	case StateOpenTmux:
		return "💻"
	case StateOpenAgy:
		return "🟢"
	case StateOpenPrivate:
		return "🔒"
	case StateDeadResuscitatable:
		return "💤"
	case StateDeadArchived:
		return "⚫"
	default:
		return "❓"
	}
}

func formatStatus(s Session) string {
	emoji := getEmojiForState(s.State)
	if s.State == StateOpenTmux {
		if s.AttachedClients == 0 {
			return "💤"
		}
		if s.AttachedClients > 1 {
			return fmt.Sprintf("💻(%d)", s.AttachedClients)
		}
		return "💻"
	}
	if s.State == StateOpenPrivate {
		if s.AttachedClients > 1 {
			return fmt.Sprintf("🔒(%d)", s.AttachedClients)
		}
		return "🔒"
	}
	return emoji
}

func FormatAge(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	d := time.Since(t)
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}
