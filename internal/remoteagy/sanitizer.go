package remoteagy

import (
	"regexp"
	"strings"
	"unicode"
)

// SanitizePaneOutput filters raw tmux pane content from an agy session,
// stripping all the UI chrome (banner, thinking indicators, tool call headers,
// spinners, status bars, separators) and returning only the meaningful content.
//
// This is used before sending output to Telegram so users get clean, readable
// messages instead of raw terminal dumps.
func SanitizePaneOutput(raw string) string {
	lines := strings.Split(raw, "\n")
	var cleaned []string
	prevBlank := false

	for _, line := range lines {
		// Strip trailing whitespace
		line = strings.TrimRightFunc(line, unicode.IsSpace)

		// Skip if it matches any noise pattern
		if isNoiseLine(line) {
			continue
		}

		// Clean up inline noise
		line = cleanInlineNoise(line)

		// Collapse multiple blank lines into one
		isBlank := strings.TrimSpace(line) == ""
		if isBlank {
			if prevBlank {
				continue
			}
			prevBlank = true
		} else {
			prevBlank = false
		}

		cleaned = append(cleaned, line)
	}

	result := strings.TrimSpace(strings.Join(cleaned, "\n"))

	// If after cleaning we only have empty content, skip
	if strings.TrimSpace(result) == "" {
		return ""
	}

	return result
}

// --- Noise detection patterns ---

var (
	// ASCII art banner characters (may have trailing text like "Antigravity CLI 1.1.8")
	bannerPattern = regexp.MustCompile(`^[\s]*[▄▀▐▌█░▒▓]{2,}`)

	// Separator lines (all box-drawing horizontal)
	separatorPattern = regexp.MustCompile(`^[─━═╌╍┄┅┈┉\-]{4,}$`)

	// Thought/thinking indicators
	thoughtPattern = regexp.MustCompile(`^▸\s+Thought for`)

	// Tool call headers
	toolCallPattern = regexp.MustCompile(`^●\s+(Bash|Read|Write|Requested Permission|Created|Updated|Deleted|Searched)`)

	// Spinner characters (braille spinners)
	spinnerPattern = regexp.MustCompile(`^[⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏⣾⣽⣻⢿⡿⣟⣯⣷]\s+`)

	// Status bar lines
	statusBarPattern = regexp.MustCompile(`^(esc to cancel|\? for shortcuts|[a-z] for)`)

	// Antigravity CLI header
	agyHeaderPattern = regexp.MustCompile(`Antigravity CLI \d+\.\d+`)

	// Gemini model indicator in status bar
	geminiStatusPattern = regexp.MustCompile(`Gemini \d+\.\d+\s+Pro\s+·\s+(low|medium|high)$`)
)

func isNoiseLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Empty lines are handled separately (collapse logic)
	if trimmed == "" {
		return false
	}

	// Bare prompt or re-echoed user input (> Some text)
	if trimmed == ">" || strings.HasPrefix(trimmed, "> ") {
		return true
	}

	// ASCII art banner
	if bannerPattern.MatchString(trimmed) {
		return true
	}

	// Separator lines
	if separatorPattern.MatchString(trimmed) {
		return true
	}

	// Thought indicators
	if thoughtPattern.MatchString(trimmed) {
		return true
	}

	// "Prioritizing Tool Usage" and similar thinking labels
	if trimmed == "Prioritizing Tool Usage" || trimmed == "Prioritizing Tool Usage..." {
		return true
	}

	// Tool call headers (● Bash(...), ● Requested Permission, etc.)
	if toolCallPattern.MatchString(trimmed) {
		return true
	}

	// Spinner lines
	if spinnerPattern.MatchString(trimmed) {
		return true
	}

	// Status bar
	if statusBarPattern.MatchString(trimmed) {
		return true
	}

	// Antigravity CLI header line
	if agyHeaderPattern.MatchString(trimmed) {
		return true
	}

	// Gemini model status bar (right-aligned)
	if geminiStatusPattern.MatchString(trimmed) {
		return true
	}

	// Lines that are just the folder path (~/git/something)
	if strings.HasPrefix(trimmed, "~/git/") && !strings.Contains(trimmed, " ") {
		return true
	}

	// "Gemini X.X Pro (Low)" standalone line
	if matched, _ := regexp.MatchString(`^Gemini \d+\.\d+ Pro \((Low|Medium|High)\)$`, trimmed); matched {
		return true
	}

	// "(ctrl+o to expand)" standalone
	if trimmed == "(ctrl+o to expand)" {
		return true
	}

	// "expand)" leftover from wrapped tool call lines
	if trimmed == "expand)" {
		return true
	}

	return false
}

// cleanInlineNoise removes inline noise from a line.
func cleanInlineNoise(line string) string {
	// Remove "(ctrl+o to expand)" suffix
	line = strings.TrimSuffix(strings.TrimSpace(line), "(ctrl+o to expand)")
	line = strings.TrimSuffix(strings.TrimSpace(line), "(ctrl+o to")

	return line
}
