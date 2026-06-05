package sessions

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
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
		if !strings.Contains(output, "⬆️") {
			t.Errorf("missing harness type emoji (⬆️) in short format output: %q", output)
		}
		if !strings.Contains(output, "💻") || !strings.Contains(output, "emagy-session-1") || !strings.Contains(output, "/workspace/proj1") {
			t.Errorf("missing open tmux session details in short format: %q", output)
		}
		if !strings.Contains(output, "This is a super long description that is definitely longe...") {
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

type MockFileInfoWithTime struct {
	name    string
	isDir   bool
	modTime time.Time
}

func (m MockFileInfoWithTime) Name() string       { return m.name }
func (m MockFileInfoWithTime) Size() int64        { return 0 }
func (m MockFileInfoWithTime) Mode() os.FileMode  { return 0 }
func (m MockFileInfoWithTime) ModTime() time.Time { return m.modTime }
func (m MockFileInfoWithTime) IsDir() bool        { return m.isDir }
func (m MockFileInfoWithTime) Sys() interface{}   { return nil }

func stripANSI(s string) string {
	var result strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if s[i] == 'm' {
				inEsc = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}

func TestListSessionsAlignment(t *testing.T) {
	now := time.Now()
	// Create mock sessions where the modification time of their DB or folder results in different age formats
	// 2s (active, <1h, should be colored green but we check alignment)
	// 15m (active, <1h, should be green)
	// 19h (active, >1h)
	// 1d (historical/archived if we include all, or just another session)
	mockTmux := &MockTmuxRunner{
		sessions: []TmuxSession{
			{Name: "emagy-session-2s", Path: "/workspace/p2s", Attached: true, Windows: 1},
			{Name: "emagy-session-15m", Path: "/workspace/p15m", Attached: true, Windows: 1},
			{Name: "emagy-session-19h", Path: "/workspace/p19h", Attached: true, Windows: 1},
			{Name: "emagy-session-1d", Path: "/workspace/p1d", Attached: true, Windows: 1},
		},
	}

	mockFS := &MockFileSystem{
		files: map[string][]byte{},
		stats: map[string]os.FileInfo{
			// Stat maps for SQLite DB files to control the modTime exactly
			"/home/ricc/.gemini/antigravity-cli/conversations/session-2s.db": MockFileInfoWithTime{
				name: "session-2s.db", modTime: now.Add(-2 * time.Second),
			},
			"/home/ricc/.gemini/antigravity-cli/conversations/session-15m.db": MockFileInfoWithTime{
				name: "session-15m.db", modTime: now.Add(-15 * time.Minute),
			},
			"/home/ricc/.gemini/antigravity-cli/conversations/session-19h.db": MockFileInfoWithTime{
				name: "session-19h.db", modTime: now.Add(-19 * time.Hour),
			},
			"/home/ricc/.gemini/antigravity-cli/conversations/session-1d.db": MockFileInfoWithTime{
				name: "session-1d.db", modTime: now.Add(-25 * time.Hour),
			},
		},
	}

	engine := NewClassificationEngine(mockTmux, mockFS, "/home/ricc")

	for _, format := range []string{"short", "long"} {
		t.Run(format, func(t *testing.T) {
			var buf bytes.Buffer
			err := ListSessions(&buf, engine, ListOptions{Format: format, All: true})
			if err != nil {
				t.Fatalf("unexpected error listing: %v", err)
			}

			output := buf.String()
			stripped := stripANSI(output)
			lines := strings.Split(strings.TrimSpace(stripped), "\n")
			if len(lines) < 2 {
				t.Fatalf("expected at least header and one data row, got output: %q", stripped)
			}

			// Find header line and the visual boundary of "AGE" column using runes
			headerRunes := []rune(lines[0])
			ageIdx := -1
			for i := 0; i <= len(headerRunes)-3; i++ {
				if headerRunes[i] == 'A' && headerRunes[i+1] == 'G' && headerRunes[i+2] == 'E' {
					ageIdx = i
					break
				}
			}
			if ageIdx == -1 {
				t.Fatalf("could not find AGE column in header: %q", lines[0])
			}
			ageRightBoundary := ageIdx + 3 // AGE has length 3

			// Verify that every subsequent line (except separators) has the age value right-aligned at ageRightBoundary
			for idx, line := range lines[1:] {
				if strings.Contains(line, "---") || len(strings.TrimSpace(line)) == 0 {
					continue
				}

				lineRunes := []rune(line)
				if len(lineRunes) <= ageRightBoundary {
					t.Errorf("line %d is too short to contain AGE: %q", idx+1, line)
					continue
				}

				// The 3-rune block [ageIdx : ageRightBoundary] contains the right-aligned age
				ageBlock := string(lineRunes[ageIdx:ageRightBoundary])
				trimmedAge := strings.TrimSpace(ageBlock)
				if len(trimmedAge) == 0 {
					t.Errorf("line %d: age block is empty: %q", idx+1, line)
				}

				lastChar := lineRunes[ageRightBoundary-1]
				nextChar := lineRunes[ageRightBoundary]

				if lastChar == ' ' {
					t.Errorf("line %d: AGE value is not right-aligned (character before column boundary is a space): %q", idx+1, line)
				}
				if nextChar != ' ' {
					t.Errorf("line %d: AGE value extends past right boundary (character at boundary is not a space): %q", idx+1, line)
				}
			}
		})
	}
}


