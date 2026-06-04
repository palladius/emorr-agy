package sessions

import (
	"reflect"
	"testing"
)

func TestParsePaneOptions(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected []PaneOption
	}{
		{
			name: "Standard Antigravity Proceed",
			lines: []string{
				"  Requesting permission for: git ls-files",
				"Do you want to proceed?",
				"> 1. Yes",
				"  2. Yes, and always allow in this conversation",
				"  3. Yes, and always allow for commands (Persist)",
				"  4. No",
				"",
				"  ↑/↓ Navigate · tab Amend",
			},
			expected: []PaneOption{
				{ID: "1", Text: "Yes"},
				{ID: "2", Text: "Yes, and always allow in this conversation"},
				{ID: "3", Text: "Yes, and always allow for commands (Persist)"},
				{ID: "4", Text: "No"},
			},
		},
		{
			name: "Parentheses option format",
			lines: []string{
				"Select an action:",
				"  1) Install dependencies",
				"  2) Run local server",
				"  3) Run automated tests",
			},
			expected: []PaneOption{
				{ID: "1", Text: "Install dependencies"},
				{ID: "2", Text: "Run local server"},
				{ID: "3", Text: "Run automated tests"},
			},
		},
		{
			name: "No options present",
			lines: []string{
				"Server started successfully on port 8080.",
				"Listening for incoming connections...",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParsePaneOptions(tt.lines)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("ParsePaneOptions() = %v, want %v", actual, tt.expected)
			}
		})
	}
}
