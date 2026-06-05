package color

import "os"

// ShouldColor returns true if ANSI colors should be enabled.
// It respects the NO_COLOR environment variable (https://no-color.org/) and TERM=dumb.
func ShouldColor() bool {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// ANSI Escape Codes - all start sequences are exactly 8 runes long to ensure tabwriter alignment
const (
	BoldWhite = "\033[01;37m"
	Blue      = "\033[00;94m" // Light blue is cleaner on dark terminals
	LightGray = "\033[00;37m"
	DarkGray  = "\033[00;90m"
	Cyan      = "\033[00;36m"
	Yellow    = "\033[00;33m"
	Green     = "\033[00;32m"
	Plain     = "\033[00;00m" // Dummy style to preserve alignment for uncolored cells
	Reset     = "\033[0m"
)

// Colorize wraps text in ANSI escape sequence if coloring is enabled.
func Colorize(text, code string) string {
	if !ShouldColor() {
		return text
	}
	return code + text + Reset
}

