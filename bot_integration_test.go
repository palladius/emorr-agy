package main

import (
	"encoding/json"
	"fmt"
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

func TestInteractiveCommandsIntegration(t *testing.T) {
	// 1. Create a temp directory for mock home and mock tmux bin
	tempDir := t.TempDir()
	mockBinDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(mockBinDir, 0755); err != nil {
		t.Fatalf("failed to create mock bin dir: %v", err)
	}

	// Create a mock tmux executable script
	mockTmuxContent := `#!/bin/bash
if [[ "$1" == "list-sessions" ]]; then
  echo -e "emagy-260604-1117\t/git/gic/bin\t0\t1"
elif [[ "$1" == "capture-pane" ]]; then
  echo "  Requesting permission for: git ls-files"
  echo "Do you want to proceed?"
  echo "> 1. Yes"
  echo "  2. Yes, and always allow in this conversation"
  echo "  3. Yes, and always allow for commands"
  echo "  4. No"
elif [[ "$1" == "send-keys" ]]; then
  # check command options/parameters if needed
  exit 0
fi
`
	tmuxPath := filepath.Join(mockBinDir, "tmux")
	if err := os.WriteFile(tmuxPath, []byte(mockTmuxContent), 0755); err != nil {
		t.Fatalf("failed to write mock tmux script: %v", err)
	}

	// Prepend mockBinDir to PATH
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+originalPath)

	// Set HOME directory to redirect config locations
	t.Setenv("HOME", tempDir)

	// Write mock db and cache folders to make sure ClassificationEngine lists them if needed
	cliCacheDir := filepath.Join(tempDir, ".gemini/antigravity-cli/cache")
	if err := os.MkdirAll(cliCacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}
	// Historical session cache JSON
	lastConvsContent := `{"/git/gic/bin": "emagy-260604-1117"}`
	if err := os.WriteFile(filepath.Join(cliCacheDir, "last_conversations.json"), []byte(lastConvsContent), 0644); err != nil {
		t.Fatalf("failed to write last_conversations.json: %v", err)
	}

	// 2. Set up mock Telegram server
	var lastSentText string
	var lastSentMarkup string
	var lastMethod string

	telegramServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = r.ParseForm()
		
		lastMethod = r.URL.Path
		if strings.Contains(r.URL.Path, "sendMessage") || strings.Contains(r.URL.Path, "editMessageText") {
			lastSentText = r.Form.Get("text")
			lastSentMarkup = r.Form.Get("reply_markup")
		}
		
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegramServer.Close()

	// Redirect telegram endpoints
	oldTelegramBaseURL := telegram.BaseURL
	telegram.BaseURL = telegramServer.URL
	defer func() {
		telegram.BaseURL = oldTelegramBaseURL
	}()

	// 3. Test "/list" command (waiting on user interaction)
	listUpdate := telegram.TelegramUpdate{
		UpdateID: 1001,
		Message: &telegram.TelegramMessage{
			MessageID: 101,
			Chat: telegram.TelegramChat{
				ID: 12345,
			},
			Text: "/list",
		},
	}

	err := processUpdate("bot-token", listUpdate)
	if err != nil {
		t.Fatalf("processUpdate failed for /list: %v", err)
	}

	if !strings.Contains(lastSentText, "Sessions Pending Human Interaction:") {
		t.Errorf("expected text to contain sessions header, got %q", lastSentText)
	}
	if !strings.Contains(lastSentMarkup, "emagy-260604-1117") {
		t.Errorf("expected markup to contain session button, got %q", lastSentMarkup)
	}

	// 4. Test "/listall" command (all sessions)
	listAllUpdate := telegram.TelegramUpdate{
		UpdateID: 1002,
		Message: &telegram.TelegramMessage{
			MessageID: 102,
			Chat: telegram.TelegramChat{
				ID: 12345,
			},
			Text: "/listall",
		},
	}

	err = processUpdate("bot-token", listAllUpdate)
	if err != nil {
		t.Fatalf("processUpdate failed for /listall: %v", err)
	}

	if !strings.Contains(lastSentText, "Last 5 Sessions:") {
		t.Errorf("expected text to contain listall header, got %q", lastSentText)
	}
	if !strings.Contains(lastSentMarkup, "emagy-260604-1117") {
		t.Errorf("expected markup to contain session button, got %q", lastSentMarkup)
	}

	// 5. Test callback query "show:emagy-260604-1117"
	showCallbackUpdate := telegram.TelegramUpdate{
		UpdateID: 1003,
		CallbackQuery: &telegram.TelegramCallbackQuery{
			ID: "query-abc",
			Message: &telegram.TelegramMessage{
				MessageID: 201,
				Chat: telegram.TelegramChat{
					ID: 12345,
				},
			},
			Data: "show:emagy-260604-1117",
		},
	}

	err = processUpdate("bot-token", showCallbackUpdate)
	if err != nil {
		t.Fatalf("processUpdate failed for callback query: %v", err)
	}

	if !strings.Contains(lastSentText, "SESSION ID:* emagy-260604-1117") {
		t.Errorf("expected details to contain session ID, got %q", lastSentText)
	}
	if !strings.Contains(lastSentText, "Requesting permission for: git ls-files") {
		t.Errorf("expected details to contain pane output, got %q", lastSentText)
	}
	// Option buttons markup check
	if !strings.Contains(lastSentMarkup, "exec:emagy-260604-1117:1") || !strings.Contains(lastSentMarkup, "1: Yes") {
		t.Errorf("expected markup to contain option buttons, got %q", lastSentMarkup)
	}

	// 6. Test callback query "exec:emagy-260604-1117:1"
	execCallbackUpdate := telegram.TelegramUpdate{
		UpdateID: 1004,
		CallbackQuery: &telegram.TelegramCallbackQuery{
			ID: "query-def",
			Message: &telegram.TelegramMessage{
				MessageID: 201,
				Chat: telegram.TelegramChat{
					ID: 12345,
				},
			},
			Data: "exec:emagy-260604-1117:1",
		},
	}

	err = processUpdate("bot-token", execCallbackUpdate)
	if err != nil {
		t.Fatalf("processUpdate failed for exec callback: %v", err)
	}

	// It should have executed the command and updated message
	if !strings.Contains(lastMethod, "editMessageText") {
		t.Errorf("expected last method to be editMessageText, got %q", lastMethod)
	}
}

func TestPingate(t *testing.T) {
	// Mock Telegram API server to check messages sent
	var lastSentText string
	telegramServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = r.ParseForm()
		if strings.Contains(r.URL.Path, "sendMessage") {
			lastSentText = r.Form.Get("text")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegramServer.Close()

	// Redirect telegram endpoints
	oldTelegramBaseURL := telegram.BaseURL
	telegram.BaseURL = telegramServer.URL
	defer func() {
		telegram.BaseURL = oldTelegramBaseURL
	}()

	// Override isTestMode to return false so we enter the PIN gate
	oldIsTestMode := isTestMode
	isTestMode = func() bool { return false }
	defer func() { isTestMode = oldIsTestMode }()

	// Reset PIN gate states
	serverUnlocked = false
	wrongPinAttempts = 0

	// 1. Send wrong PIN
	updateWrong := telegram.TelegramUpdate{
		UpdateID: 101,
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat: telegram.TelegramChat{ID: 12345},
			Text: "1111",
		},
	}
	err := processUpdate("bot-token", updateWrong)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serverUnlocked {
		t.Error("expected server to remain locked after wrong PIN")
	}
	if wrongPinAttempts != 1 {
		t.Errorf("expected wrongPinAttempts to be 1, got %d", wrongPinAttempts)
	}
	if !strings.Contains(lastSentText, "Server is locked") {
		t.Errorf("expected warning message, got %q", lastSentText)
	}

	// 2. Send correct PIN
	updateCorrect := telegram.TelegramUpdate{
		UpdateID: 102,
		Message: &telegram.TelegramMessage{
			MessageID: 2,
			Chat: telegram.TelegramChat{ID: 12345},
			Text: "4242",
		},
	}
	err = processUpdate("bot-token", updateCorrect)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !serverUnlocked {
		t.Error("expected server to be unlocked after correct PIN")
	}
	if wrongPinAttempts != 0 {
		t.Errorf("expected wrongPinAttempts to be reset to 0, got %d", wrongPinAttempts)
	}
	if !strings.Contains(lastSentText, "Server unlocked successfully") {
		t.Errorf("expected success message, got %q", lastSentText)
	}
}

func TestPingateFatalKills(t *testing.T) {
	// Mock Telegram API server
	telegramServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegramServer.Close()

	oldTelegramBaseURL := telegram.BaseURL
	telegram.BaseURL = telegramServer.URL
	defer func() {
		telegram.BaseURL = oldTelegramBaseURL
	}()

	// Override isTestMode
	oldIsTestMode := isTestMode
	isTestMode = func() bool { return false }
	defer func() { isTestMode = oldIsTestMode }()

	// Reset PIN gate states
	serverUnlocked = false
	wrongPinAttempts = 0

	// Mock logFatalf
	fatalCalled := false
	var fatalMsg string
	oldLogFatalf := logFatalf
	logFatalf = func(format string, v ...interface{}) {
		fatalCalled = true
		fatalMsg = fmt.Sprintf(format, v...)
	}
	defer func() { logFatalf = oldLogFatalf }()

	updateWrong := telegram.TelegramUpdate{
		UpdateID: 101,
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat: telegram.TelegramChat{ID: 12345},
			Text: "wrong",
		},
	}

	// Send wrong PIN 3 times
	for i := 0; i < 3; i++ {
		_ = processUpdate("bot-token", updateWrong)
	}

	if !fatalCalled {
		t.Error("expected logFatalf to be called after 3 wrong PIN attempts")
	}
	if !strings.Contains(fatalMsg, "3 wrong PIN attempts") {
		t.Errorf("expected fatal msg to mention 3 wrong PIN attempts, got %q", fatalMsg)
	}
}

func TestNewCommand(t *testing.T) {
	// Mock Telegram API server
	var lastSentText string
	var lastSentMarkup string
	telegramServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = r.ParseForm()
		if strings.Contains(r.URL.Path, "sendMessage") {
			lastSentText = r.Form.Get("text")
			lastSentMarkup = r.Form.Get("reply_markup")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer telegramServer.Close()

	oldTelegramBaseURL := telegram.BaseURL
	telegram.BaseURL = telegramServer.URL
	defer func() {
		telegram.BaseURL = oldTelegramBaseURL
	}()

	// Mock a bin folder for tmux
	tempDir := t.TempDir()
	mockBinDir := filepath.Join(tempDir, "bin")
	_ = os.MkdirAll(mockBinDir, 0755)
	
	// Create mock tmux script
	mockTmuxContent := `#!/bin/bash
echo "mock tmux called with args: $@"
exit 0
`
	_ = os.WriteFile(filepath.Join(mockBinDir, "tmux"), []byte(mockTmuxContent), 0755)

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", mockBinDir+":"+originalPath)
	t.Setenv("HOME", tempDir)

	updateNew := telegram.TelegramUpdate{
		UpdateID: 101,
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat: telegram.TelegramChat{ID: 12345},
			Text: "/new gemini write a go test",
		},
	}

	err := processUpdate("bot-token", updateNew)
	if err != nil {
		t.Fatalf("unexpected error processing /new: %v", err)
	}

	if !strings.Contains(lastSentText, "Spawned new tmux session") {
		t.Errorf("expected text to mention spawned session, got %q", lastSentText)
	}
	if !strings.Contains(lastSentText, "harness `gemini`") {
		t.Errorf("expected harness to be gemini, got %q", lastSentText)
	}
	if !strings.Contains(lastSentText, "query: \"write a go test\"") {
		t.Errorf("expected query to be parsed, got %q", lastSentText)
	}
	// Verify that the reply markup contains the session name button
	if !strings.Contains(lastSentMarkup, "emgem-") {
		t.Errorf("expected markup to contain inline button for session, got %q", lastSentMarkup)
	}

	// Test case: voice command "new claude test voice input"
	updateVoice := telegram.TelegramUpdate{
		UpdateID: 102,
		Message: &telegram.TelegramMessage{
			MessageID: 2,
			Chat: telegram.TelegramChat{ID: 12345},
			Text: "new claude test voice input",
		},
	}
	err = processUpdate("bot-token", updateVoice)
	if err != nil {
		t.Fatalf("unexpected error processing voice new: %v", err)
	}

	if !strings.Contains(lastSentText, "Spawned new tmux session") {
		t.Errorf("expected text to mention spawned session, got %q", lastSentText)
	}
	if !strings.Contains(lastSentText, "harness `claude`") {
		t.Errorf("expected harness to be claude, got %q", lastSentText)
	}
	if !strings.Contains(lastSentText, "query: \"test voice input\"") {
		t.Errorf("expected query to be parsed, got %q", lastSentText)
	}
	if !strings.Contains(lastSentMarkup, "emcld-") {
		t.Errorf("expected markup to contain inline button for session, got %q", lastSentMarkup)
	}
}
