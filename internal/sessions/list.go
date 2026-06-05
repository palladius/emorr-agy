package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/palladius/emorr-agy/internal/color"
)

type ListOptions struct {
	Harness []string
	Format  string // "short", "long", "json"
	All     bool   // If false, exclude archived sessions
}

// ListSessions retrieves and prints the classified sessions to the provided writer.
func ListSessions(w io.Writer, engine *ClassificationEngine, opts ListOptions) error {
	sessions, err := engine.Classify(opts.Harness)
	if err != nil {
		return err
	}

	if !opts.All {
		var filtered []Session
		for _, s := range sessions {
			if s.State != StateDeadArchived {
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
			folder := strings.ReplaceAll(s.Folder, "/usr/local/google/home/ricc", "~")
			if len(folder) > 25 {
				folder = folder[:22] + "..."
			}

			ageColor := color.LightGray
			if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
				ageColor = color.Green
			} else if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			desc := strings.TrimSpace(s.Description)
			desc = strings.ReplaceAll(desc, "\n", " ")
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			statusAndHarness := fmt.Sprintf("%s %s", emoji, harnessEmoji)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				color.Colorize(statusAndHarness, color.Plain),
				color.Colorize(s.ID, color.BoldWhite),
				color.Colorize(fmt.Sprintf("%3s", age), ageColor),
				color.Colorize(folder, color.Blue),
				color.Colorize(s.Harness, color.Plain),
				color.Colorize(strconv.Itoa(s.ProcessCount), color.Plain),
				color.Colorize(s.ResumeCommand, color.Plain),
				color.Colorize(desc, color.Cyan),
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
			folder := strings.ReplaceAll(s.Folder, "/usr/local/google/home/ricc", "~")
			if len(folder) > 25 {
				folder = folder[:22] + "..."
			}

			ageColor := color.LightGray
			if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
				ageColor = color.Green
			} else if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			desc := strings.TrimSpace(s.Description)
			desc = strings.ReplaceAll(desc, "\n", " ")
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			statusAndHarness := fmt.Sprintf("%s %s", emoji, harnessEmoji)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				color.Colorize(statusAndHarness, color.Plain),
				color.Colorize(s.ID, color.BoldWhite),
				color.Colorize(fmt.Sprintf("%3s", age), ageColor),
				color.Colorize(folder, color.Blue),
				color.Colorize(desc, color.Cyan),
			)
		}
		tw.Flush()
	}

	return nil
}

func getEmojiForHarness(harness string) string {
	switch harness {
	case "gemini":
		return "♊"
	case "agy":
		return "⬆"
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
