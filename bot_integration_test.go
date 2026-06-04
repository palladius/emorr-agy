package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/palladius/emorr-agy/internal/gemini"
	"github.com/palladius/emorr-agy/internal/telegram"
)

func TestProcessVoiceUpdateIntegration(t *testing.T) {
	// Set up local file directory setup
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home: %v", err)
	}
	tmpPathDir := filepath.Join(homeDir, ".gemini/antigravity-cli/tmp")
	_ = os.MkdirAll(tmpPathDir, 0755)

	// Mock Telegram API server
	telegramServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Handle getFile request
		if strings.Contains(r.URL.Path, "getFile") {
			var resp telegram.FileResponse
			resp.Ok = true
			resp.Result.FileID = "voice-12345"
			resp.Result.FilePath = "voice/voice-12345.ogg"
			resp.Result.FileSize = 100
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Handle sendMessage request
		if strings.Contains(r.URL.Path, "sendMessage") {
			_ = r.ParseForm()
			chat_id := r.Form.Get("chat_id")
			text := r.Form.Get("text")
			if chat_id != "605724096" {
				t.Errorf("expected chat_id 605724096, got %s", chat_id)
			}
			// Voice transcription reply is italicized with emoji: "🇮🇹 _status_"
			// Check if it is the transcription reply or command response
			if !strings.Contains(text, "🇮🇹 _status_") && !strings.Contains(text, "📡 Emorr-Agy Status:") {
				t.Errorf("unexpected sendMessage text: %q", text)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer telegramServer.Close()

	// Mock Telegram file download server
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "voice/voice-12345.ogg") {
			t.Errorf("unexpected file download path: %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fake-ogg-audio-binary"))
	}))
	defer fileServer.Close()

	// Mock Gemini API server
	geminiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		var mockResp struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		mockResp.Candidates = make([]struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		}, 1)
		mockResp.Candidates[0].Content.Parts = []struct {
			Text string `json:"text"`
		}{
			{Text: `{"text": "status", "language": "it"}`},
		}

		_ = json.NewEncoder(w).Encode(mockResp)
	}))
	defer geminiServer.Close()

	// Override global endpoints and command functions
	oldTelegramBaseURL := telegram.BaseURL
	oldTelegramFileBaseURL := telegram.FileBaseURL
	oldGeminiAPIURL := gemini.GeminiAPIURL
	oldStatusFunc := getStatusOutputFunc
	oldMonitorFunc := getMonitorOutputFunc

	telegram.BaseURL = telegramServer.URL
	telegram.FileBaseURL = fileServer.URL
	gemini.GeminiAPIURL = geminiServer.URL + "?key=%s"
	getStatusOutputFunc = func() (string, error) {
		return "📡 Emorr-Agy Status:\nFake Status", nil
	}
	getMonitorOutputFunc = func() (string, error) {
		return "📡 Antigravity Thread Monitor:\nFake Monitor", nil
	}

	defer func() {
		telegram.BaseURL = oldTelegramBaseURL
		telegram.FileBaseURL = oldTelegramFileBaseURL
		gemini.GeminiAPIURL = oldGeminiAPIURL
		getStatusOutputFunc = oldStatusFunc
		getMonitorOutputFunc = oldMonitorFunc
	}()

	t.Setenv("GEMINI_API_KEY", "fake-gemini-key")

	// Trigger processUpdate
	update := telegram.TelegramUpdate{
		UpdateID: 456,
		Message: &telegram.TelegramMessage{
			MessageID: 12,
			Chat: telegram.TelegramChat{
				ID: 605724096,
			},
			Voice: &telegram.TelegramVoice{
				FileID:   "voice-12345",
				Duration: 5,
				MimeType: "audio/ogg",
				FileSize: 100,
			},
		},
	}

	err = processUpdate("bot-token", update)
	if err != nil {
		t.Fatalf("unexpected error processing update: %v", err)
	}
}
