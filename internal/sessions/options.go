package sessions

import (
	"regexp"
	"strings"
)

type PaneOption struct {
	ID   string
	Text string
}

// optionRegex matches numeric options formatted as "1. Option text" or "  1) Option text"
// with optional cursor indicators like "> 1. Option text".
var optionRegex = regexp.MustCompile(`^[\s>]*(\d+)[\.\)]\s*(.+)$`)

// ParsePaneOptions parses detected choice options from tmux terminal output lines.
func ParsePaneOptions(lines []string) []PaneOption {
	var opts []PaneOption
	hasYesNo := false

	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "(y/n)") ||
			strings.Contains(lowerLine, "[y/n]") ||
			strings.Contains(lowerLine, "y/n?") ||
			strings.Contains(lowerLine, "proceed?") ||
			strings.Contains(lowerLine, "approve?") {
			hasYesNo = true
		}

		matches := optionRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			id := matches[1]
			text := strings.TrimSpace(matches[2])
			opts = append(opts, PaneOption{
				ID:   id,
				Text: text,
			})
		}
	}

	if hasYesNo && len(opts) == 0 {
		opts = append(opts, PaneOption{ID: "y", Text: "Yes / Approve"})
		opts = append(opts, PaneOption{ID: "n", Text: "No / Deny"})
	}

	if len(opts) == 0 {
		return nil
	}
	return opts
}
