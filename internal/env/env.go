package env

import (
	"os"
	"strings"
)

// CleanValue trims spaces, strips surrounding quotes, and strips trailing 's.
func CleanValue(s string) string {
	s = strings.TrimSpace(s)
	// Remove enclosing single quotes
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
	}
	// Remove enclosing double quotes
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
	}
	// Strip trailing 's
	if strings.HasSuffix(s, "'s") {
		s = strings.TrimSuffix(s, "'s")
	}
	// Strip extra enclosing quotes
	s = strings.Trim(s, "\"'")
	return s
}

// GetEnvWithFallback checks environment variables in order and returns the first non-empty value.
func GetEnvWithFallback(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

// GetGeminiModel returns the configured Gemini model, checking
// AUDIO_TRANSCRIPTION_GEMINI_MODEL and GEMINI_MODEL env vars,
// defaulting to DefaultGeminiModel.
func GetGeminiModel() string {
	model := GetEnvWithFallback("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "GEMINI_MODEL")
	if model == "" {
		return DefaultGeminiModel
	}
	return CleanValue(model)
}

