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
