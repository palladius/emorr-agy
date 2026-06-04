package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
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
		fmt.Fprintln(tw, "STATUS\tSESSION ID\tDIRECTORY\tHARNESS\tWINDOWS\tRESUME COMMAND")
		for _, s := range sessions {
			emoji := getEmojiForState(s.State)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n", emoji, s.ID, s.Folder, s.Harness, s.ProcessCount, s.ResumeCommand)
		}
		tw.Flush()

	default: // "short" or fallback
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "STATUS\tSESSION ID\tDIRECTORY")
		for _, s := range sessions {
			emoji := getEmojiForState(s.State)
			fmt.Fprintf(tw, "%s\t%s\t%s\n", emoji, s.ID, s.Folder)
		}
		tw.Flush()
	}

	return nil
}

func getEmojiForState(state SessionState) string {
	switch state {
	case StateOpenTmux:
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
