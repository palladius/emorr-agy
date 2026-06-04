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
}

// ListSessions retrieves and prints the classified sessions to the provided writer.
func ListSessions(w io.Writer, engine *ClassificationEngine, opts ListOptions) error {
	sessions, err := engine.Classify(opts.Harness)
	if err != nil {
		return err
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
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			color.Colorize("STATUS", color.Plain),
			color.Colorize("TYPE", color.Plain),
			color.Colorize("SESSION ID", color.Plain),
			color.Colorize("AGE", color.Plain),
			color.Colorize("DIRECTORY", color.Plain),
			color.Colorize("HARNESS", color.Plain),
			color.Colorize("WINDOWS", color.Plain),
			color.Colorize("RESUME COMMAND", color.Plain),
		)
		for _, s := range sessions {
			emoji := getEmojiForState(s.State)
			harnessEmoji := getEmojiForHarness(s.Harness)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, "/usr/local/google/home/ricc", "~")

			ageColor := color.LightGray
			if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				color.Colorize(emoji, color.Plain),
				color.Colorize(harnessEmoji, color.Plain),
				color.Colorize(s.ID, color.BoldWhite),
				color.Colorize(age, ageColor),
				color.Colorize(folder, color.Blue),
				color.Colorize(s.Harness, color.Plain),
				color.Colorize(strconv.Itoa(s.ProcessCount), color.Plain),
				color.Colorize(s.ResumeCommand, color.Plain),
			)
		}
		tw.Flush()

	default: // "short" or fallback
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			color.Colorize("STATUS", color.Plain),
			color.Colorize("TYPE", color.Plain),
			color.Colorize("SESSION ID", color.Plain),
			color.Colorize("AGE", color.Plain),
			color.Colorize("DIRECTORY", color.Plain),
		)
		for _, s := range sessions {
			emoji := getEmojiForState(s.State)
			harnessEmoji := getEmojiForHarness(s.Harness)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, "/usr/local/google/home/ricc", "~")

			ageColor := color.LightGray
			if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
				color.Colorize(emoji, color.Plain),
				color.Colorize(harnessEmoji, color.Plain),
				color.Colorize(s.ID, color.BoldWhite),
				color.Colorize(age, ageColor),
				color.Colorize(folder, color.Blue),
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
		return "🎈"
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

func FormatAge(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
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
