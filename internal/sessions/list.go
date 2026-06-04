package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
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
		fmt.Fprintln(tw, "STATUS\tSESSION ID\tAGE\tDIRECTORY\tHARNESS\tWINDOWS\tRESUME COMMAND")
		for _, s := range sessions {
			emoji := getEmojiForState(s.State)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, "/usr/local/google/home/ricc", "~")
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n", emoji, s.ID, age, folder, s.Harness, s.ProcessCount, s.ResumeCommand)
		}
		tw.Flush()

	default: // "short" or fallback
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "STATUS\tSESSION ID\tAGE\tDIRECTORY")
		for _, s := range sessions {
			emoji := getEmojiForState(s.State)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, "/usr/local/google/home/ricc", "~")
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", emoji, s.ID, age, folder)
		}
		tw.Flush()
	}

	return nil
}

func getEmojiForState(state SessionState) string {
	switch state {
	case StateOpenTmux:
		return "🖥️"
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
