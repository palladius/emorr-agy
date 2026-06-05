package sessions

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestListSessionsFormats(t *testing.T) {
	mockTmux := &MockTmuxRunner{
		sessions: []TmuxSession{
			{Name: "emagy-session-1", Path: "/workspace/proj1", Attached: true, Windows: 2},
			{Name: "my-private-session", Path: "/home/ricc", Attached: true, Windows: 3},
		},
	}

	mockFS := &MockFileSystem{
		files: map[string][]byte{
			"/home/ricc/.gemini/antigravity-cli/cache/last_conversations.json": []byte(`{
				"/workspace/proj3": "session-Chumbia",
				"/workspace/proj4": "session-dead-active"
			}`),
			"/home/ricc/.emorr-agy/cache/session-1.json": []byte(`{
				"about": "This is a super long description that is definitely longer than thirty characters"
			}`),
		},
	}

	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	t.Run("Short Format (Default - Excludes Archived)", func(t *testing.T) {
		var buf bytes.Buffer
		err := ListSessions(&buf, engine, ListOptions{Format: "short"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		// Expect header and columns
		if !strings.Contains(output, "ST") || !strings.Contains(output, "T") || !strings.Contains(output, "SESSION ID") || !strings.Contains(output, "DIR") || !strings.Contains(output, "DESCRIPTION") {
			t.Errorf("missing headers in short format output: %q", output)
		}
		if !strings.Contains(output, "🎈") {
			t.Errorf("missing harness type emoji (🎈) in short format output: %q", output)
		}
		if !strings.Contains(output, "💻") || !strings.Contains(output, "emagy-session-1") || !strings.Contains(output, "/workspace/p...") {
			t.Errorf("missing open tmux session details in short format: %q", output)
		}
		if !strings.Contains(output, "This is a super long description that is defini...") {
			t.Errorf("missing truncated description in short format: %q", output)
		}
		if !strings.Contains(output, "🔒") || !strings.Contains(output, "my-private-session") {
			t.Errorf("missing private session details in short format: %q", output)
		}
		if strings.Contains(output, "⚫") || strings.Contains(output, "session-Chumbia") {
			t.Errorf("archived session should be filtered out by default: %q", output)
		}
		if !strings.Contains(output, "💤") || !strings.Contains(output, "session-dead-active") {
			t.Errorf("missing dead resuscitatable session details in short format: %q", output)
		}
	})

	t.Run("Short Format (All - Includes Archived and Separator)", func(t *testing.T) {
		var buf bytes.Buffer
		err := ListSessions(&buf, engine, ListOptions{Format: "short", All: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "⚫") || !strings.Contains(output, "session-Chumbia") {
			t.Errorf("missing dead archived session details when All is true: %q", output)
		}
		if !strings.Contains(output, "---") {
			t.Errorf("missing visual separator when archived sessions are included: %q", output)
		}
	})

	t.Run("Long Format", func(t *testing.T) {
		var buf bytes.Buffer
		err := ListSessions(&buf, engine, ListOptions{Format: "long", All: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "HARNESS") || !strings.Contains(output, "WINDOWS") || !strings.Contains(output, "RESUME COMMAND") || !strings.Contains(output, "DESCRIPTION") {
			t.Errorf("missing headers in long format output: %q", output)
		}
		if !strings.Contains(output, "tmux attach -t emagy-session-1") && !strings.Contains(output, "emorr-agy resume") {
			t.Errorf("missing resume command in long format output: %q", output)
		}
	})

	t.Run("JSON Format (All)", func(t *testing.T) {
		var buf bytes.Buffer
		err := ListSessions(&buf, engine, ListOptions{Format: "json", All: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var sessionsList []Session
		if err := json.Unmarshal(buf.Bytes(), &sessionsList); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		if len(sessionsList) != 4 {
			t.Errorf("expected 4 sessions in JSON, got %d", len(sessionsList))
		}
	})

	t.Run("Filter by Harness (All)", func(t *testing.T) {
		var buf bytes.Buffer
		err := ListSessions(&buf, engine, ListOptions{Format: "json", Harness: []string{"agy"}, All: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var sessionsList []Session
		if err := json.Unmarshal(buf.Bytes(), &sessionsList); err != nil {
			t.Fatalf("failed to parse JSON output: %v", err)
		}

		// emagy-session-1 (agy), session-Chumbia (agy), session-dead-active (agy)
		if len(sessionsList) != 3 {
			t.Errorf("expected 3 sessions in filtered JSON, got %d", len(sessionsList))
		}
		for _, s := range sessionsList {
			if s.Harness != "agy" {
				t.Errorf("expected only 'agy' harness, got %s for session %s", s.Harness, s.ID)
			}
		}
	})
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name     string
		session  Session
		expected string
	}{
		{
			name: "Open Tmux Attached 1",
			session: Session{
				State:           StateOpenTmux,
				AttachedClients: 1,
			},
			expected: "💻",
		},
		{
			name: "Open Tmux Attached 2",
			session: Session{
				State:           StateOpenTmux,
				AttachedClients: 2,
			},
			expected: "💻(2)",
		},
		{
			name: "Open Tmux Detached",
			session: Session{
				State:           StateOpenTmux,
				AttachedClients: 0,
			},
			expected: "💤",
		},
		{
			name: "Open Private Attached 1",
			session: Session{
				State:           StateOpenPrivate,
				AttachedClients: 1,
			},
			expected: "🔒",
		},
		{
			name: "Open Private Attached 3",
			session: Session{
				State:           StateOpenPrivate,
				AttachedClients: 3,
			},
			expected: "🔒(3)",
		},
		{
			name: "Open Private Detached",
			session: Session{
				State:           StateOpenPrivate,
				AttachedClients: 0,
			},
			expected: "🔒",
		},
		{
			name: "Open Agy",
			session: Session{
				State: StateOpenAgy,
			},
			expected: "🟢",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatStatus(tc.session)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

