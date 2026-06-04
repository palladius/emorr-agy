package env

import (
	"testing"
)

func TestCleanValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"'hello'", "hello"},
		{"\"world\"", "world"},
		{"'foo's", "foo"},
		{"\"bar's\"", "bar"},
		{"  trimmed  ", "trimmed"},
		{"'trimmed_and_quoted'", "trimmed_and_quoted"},
	}

	for _, tt := range tests {
		actual := CleanValue(tt.input)
		if actual != tt.expected {
			t.Errorf("CleanValue(%q) = %q; expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestGetEnvWithFallback(t *testing.T) {
	t.Setenv("TEST_KEY_A", "val_a")
	t.Setenv("TEST_KEY_B", "val_b")

	// Match first key
	if val := GetEnvWithFallback("TEST_KEY_A", "TEST_KEY_B"); val != "val_a" {
		t.Errorf("expected val_a, got %q", val)
	}

	// Match fallback
	if val := GetEnvWithFallback("TEST_KEY_NONEXISTENT", "TEST_KEY_B"); val != "val_b" {
		t.Errorf("expected val_b, got %q", val)
	}

	// Match none
	if val := GetEnvWithFallback("TEST_KEY_NONEXISTENT_A", "TEST_KEY_NONEXISTENT_B"); val != "" {
		t.Errorf("expected empty string, got %q", val)
	}
}

func TestGetGeminiModel(t *testing.T) {
	// 1. Test fallback default
	t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "")
	t.Setenv("GEMINI_MODEL", "")
	if model := GetGeminiModel(); model != DefaultGeminiModel {
		t.Errorf("expected %q, got %q", DefaultGeminiModel, model)
	}

	// 2. Test GEMINI_MODEL
	t.Setenv("GEMINI_MODEL", "gemini-test-1")
	if model := GetGeminiModel(); model != "gemini-test-1" {
		t.Errorf("expected gemini-test-1, got %q", model)
	}

	// 3. Test AUDIO_TRANSCRIPTION_GEMINI_MODEL overrides GEMINI_MODEL
	t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "gemini-test-2")
	if model := GetGeminiModel(); model != "gemini-test-2" {
		t.Errorf("expected gemini-test-2, got %q", model)
	}

	// 4. Test CleanValue is applied
	t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "'gemini-test-3'")
	if model := GetGeminiModel(); model != "gemini-test-3" {
		t.Errorf("expected gemini-test-3, got %q", model)
	}
}

