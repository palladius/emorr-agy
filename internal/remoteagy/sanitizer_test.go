package remoteagy

import (
	"strings"
	"testing"
)

func TestSanitizePaneOutput_RealAgyOutput(t *testing.T) {
	// Real output from tmux capture-pane of an agy session
	raw := `      ▄▀▀▄        Antigravity CLI 1.1.8
     ▀▀▀▀▀▀       Gemini 3.1 Pro (Low)
    ▀▀▀▀▀▀▀▀      ~/git/gic
   ▄▀▀    ▀▀▄
  ▄▀▀      ▀▀▄

────────────────────────────────────────────────────────────
> Dimmi quali sono i cinque script più nuovi nella cartella bin.

▸ Thought for 2s, 774 tokens
  Prioritizing Tool Usage

● Bash(ls -lt ~/git/gic/bin | grep -v '^d' | head -n 6) (ctrl+o to expand)

▸ Thought for 5s, 406 tokens
  Prioritizing Tool Usage

● Requested Permission: read_file(/Users/ricc/git/gic/bin) (ctrl+o to expand)
▸ Thought for 6s, 488 tokens
  Prioritizing Tool Usage

● Bash(find /Users/ricc/git/gic/bin -maxdepth 1 -type f -ex...) (ctrl+o to
expand)
▸ Thought for 2s, 692 tokens
  Prioritizing Tool Usage

● Bash(find /Users/ricc/git/gic/bin -maxdepth 1 -type f -ex...) (ctrl+o to
expand)

▸ Thought for 4s, 344 tokens
  Prioritizing Tool Usage
  I cinque script più nuovi nella cartella bin (/Users/ricc/git/gic/bin)
  sono:

  1. omni-video-gen.py
  2. obsidian-touch-beacon
  3. autopush-enabled
  4. repos-ricc
  5. git-wazzup

  C'è qualcos'altro in cui posso aiutarti con questi script, RiccardinoCM26
  🦖?

────────────────────────────────────────────────────────────────────────────────
>
────────────────────────────────────────────────────────────────────────────────
? for shortcuts                                             Gemini 3.1 Pro · low`

	result := SanitizePaneOutput(raw)

	// Should contain the actual answer
	if !strings.Contains(result, "omni-video-gen.py") {
		t.Error("should contain the script name omni-video-gen.py")
	}
	if !strings.Contains(result, "obsidian-touch-beacon") {
		t.Error("should contain obsidian-touch-beacon")
	}
	if !strings.Contains(result, "RiccardinoCM26") {
		t.Error("should contain RiccardinoCM26")
	}

	// Should NOT contain noise
	if strings.Contains(result, "Antigravity CLI") {
		t.Error("should not contain Antigravity CLI header")
	}
	if strings.Contains(result, "▸ Thought for") {
		t.Error("should not contain thought indicators")
	}
	if strings.Contains(result, "Prioritizing Tool Usage") {
		t.Error("should not contain thinking labels")
	}
	if strings.Contains(result, "● Bash(") {
		t.Error("should not contain tool call headers")
	}
	if strings.Contains(result, "● Requested Permission") {
		t.Error("should not contain permission requests")
	}
	if strings.Contains(result, "ctrl+o to expand") {
		t.Error("should not contain ctrl+o hint")
	}
	if strings.Contains(result, "? for shortcuts") {
		t.Error("should not contain status bar")
	}
	if strings.Contains(result, "────────") {
		t.Error("should not contain separator lines")
	}
	if strings.Contains(result, "▄▀▀") {
		t.Error("should not contain ASCII art banner")
	}
	if strings.Contains(result, "esc to cancel") {
		t.Error("should not contain esc status bar")
	}

	t.Logf("Sanitized output:\n---\n%s\n---", result)
}

func TestSanitizePaneOutput_EmptyAfterClean(t *testing.T) {
	// Just a banner + status bar, no content
	raw := `      ▄▀▀▄        Antigravity CLI 1.1.8
     ▀▀▀▀▀▀       Gemini 3.1 Pro (Low)
    ▀▀▀▀▀▀▀▀      ~/git/gic
   ▄▀▀    ▀▀▄
  ▄▀▀      ▀▀▄

────────────────────────────────────────────────────────────
>
────────────────────────────────────────────────────────────
? for shortcuts                                             Gemini 3.1 Pro · low`

	result := SanitizePaneOutput(raw)
	if result != "" {
		t.Errorf("expected empty after cleaning pure chrome, got: %q", result)
	}
}

func TestSanitizePaneOutput_SpinnerOnly(t *testing.T) {
	raw := `⣻  Prioritizing Tool Usage...`
	result := SanitizePaneOutput(raw)
	if result != "" {
		t.Errorf("expected empty for spinner-only content, got: %q", result)
	}
}

func TestSanitizePaneOutput_PlainText(t *testing.T) {
	raw := "Hello, this is a normal response.\nLine two."
	result := SanitizePaneOutput(raw)
	if result != raw {
		t.Errorf("plain text should pass through unchanged, got: %q", result)
	}
}

func TestSanitizePaneOutput_CollapsesBlankLines(t *testing.T) {
	raw := "Line 1\n\n\n\n\nLine 2"
	result := SanitizePaneOutput(raw)
	expected := "Line 1\n\nLine 2"
	if result != expected {
		t.Errorf("expected collapsed blanks %q, got %q", expected, result)
	}
}
