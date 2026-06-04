package gemini

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTranscribe(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gemini-transcribe-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	audioPath := filepath.Join(tmpDir, "test.ogg")
	if err := os.WriteFile(audioPath, []byte("fake-audio-payload"), 0644); err != nil {
		t.Fatalf("failed to write test audio: %v", err)
	}

	// Prepare mock response
	mockRespText := `{"text": "status", "language": "it"}`
	mockGeminiResp := map[string]interface{}{
		"candidates": []map[string]interface{}{
			{
				"content": map[string]interface{}{
					"parts": []map[string]interface{}{
						{"text": mockRespText},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-api-key" {
			t.Errorf("expected API key 'test-api-key', got %q", r.URL.Query().Get("key"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockGeminiResp)
	}))
	defer server.Close()

	// Override API URL
	oldURL := GeminiAPIURL
	GeminiAPIURL = server.URL + "?key=%s"
	defer func() { GeminiAPIURL = oldURL }()

	transcriber := NewGeminiTranscriber("test-api-key")
	res, err := transcriber.Transcribe(audioPath, "audio/ogg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Text != "status" {
		t.Errorf("expected text 'status', got %q", res.Text)
	}
	if res.Language != "it" {
		t.Errorf("expected language 'it', got %q", res.Language)
	}
}

func TestTranscribeErrors(t *testing.T) {
	transcriber := NewGeminiTranscriber("")
	_, err := transcriber.Transcribe("path/to/file", "audio/ogg")
	if err == nil {
		t.Error("expected error for empty API key, got nil")
	}

	transcriberValid := NewGeminiTranscriber("key")
	_, err = transcriberValid.Transcribe("nonexistent-file.ogg", "audio/ogg")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestMapLanguageToFlag(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"it", "🇮🇹"},
		{"en", "🇬🇧"},
		{"gb", "🇬🇧"},
		{"us", "🇺🇸"},
		{"es", "🇪🇸"},
		{"fr", "🇫🇷"},
		{"de", "🇩🇪"},
		{"pt", "🇵🇹"},
		{"ja", "🇯🇵"},
		{"zh", "🇨🇳"},
		{"ru", "🇷🇺"},
		{"unknown", "🌐"},
		{"  IT  ", "🇮🇹"},
	}

	for _, tc := range cases {
		actual := MapLanguageToFlag(tc.input)
		if actual != tc.expected {
			t.Errorf("MapLanguageToFlag(%q) = %q; expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestGetAPIURLDynamic(t *testing.T) {
	transcriber := NewGeminiTranscriber("my-key")

	// 1. Without env override (defaulting to gemini-3.1-flash-lite)
	t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "")
	t.Setenv("GEMINI_MODEL", "")
	url := transcriber.getAPIURL()
	expected := "https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite:generateContent?key=my-key"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}

	// 2. With env override
	t.Setenv("AUDIO_TRANSCRIPTION_GEMINI_MODEL", "gemini-test-override")
	url = transcriber.getAPIURL()
	expected = "https://generativelanguage.googleapis.com/v1beta/models/gemini-test-override:generateContent?key=my-key"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}
}

