package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendTelegramMessageToChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botfake-token/sendMessage" {
			t.Errorf("expected path /botfake-token/sendMessage, got %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %q", r.Method)
		}
		err := r.ParseForm()
		if err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.Form.Get("chat_id") != "12345" {
			t.Errorf("expected chat_id 12345, got %q", r.Form.Get("chat_id"))
		}
		if r.Form.Get("text") != "hello test" {
			t.Errorf("expected text 'hello test', got %q", r.Form.Get("text"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	BaseURL = server.URL

	err := SendTelegramMessageToChat("fake-token", 12345, "hello test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetTelegramUpdates(t *testing.T) {
	fakeUpdates := TelegramUpdateResponse{
		Ok: true,
		Result: []TelegramUpdate{
			{
				UpdateID: 100,
				Message: &TelegramMessage{
					MessageID: 1,
					Chat: TelegramChat{
						ID: 12345,
					},
					Text: "/status",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botfake-token/getUpdates" {
			t.Errorf("expected path /botfake-token/getUpdates, got %q", r.URL.Path)
		}
		if r.URL.Query().Get("offset") != "50" {
			t.Errorf("expected offset 50, got %q", r.URL.Query().Get("offset"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(fakeUpdates)
	}))
	defer server.Close()

	BaseURL = server.URL

	updates, err := GetTelegramUpdates("fake-token", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].UpdateID != 100 {
		t.Errorf("expected update_id 100, got %d", updates[0].UpdateID)
	}
	if updates[0].Message.Text != "/status" {
		t.Errorf("expected text /status, got %q", updates[0].Message.Text)
	}
}

func TestSendTelegramMessage(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_ID", "env-fake-token")
	t.Setenv("TELEGRAM_CHAT_ID", "98765")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botenv-fake-token/sendMessage" {
			t.Errorf("expected path /botenv-fake-token/sendMessage, got %q", r.URL.Path)
		}
		err := r.ParseForm()
		if err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.Form.Get("chat_id") != "98765" {
			t.Errorf("expected chat_id 98765, got %q", r.Form.Get("chat_id"))
		}
		if r.Form.Get("text") != "test generic message" {
			t.Errorf("expected text 'test generic message', got %q", r.Form.Get("text"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	BaseURL = server.URL

	err := SendTelegramMessage("test generic message")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendTelegramMessageErrors(t *testing.T) {
	// 1. Missing bot token
	t.Setenv("TELEGRAM_BOT_ID", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_APITOKEN", "")
	err := SendTelegramMessage("test")
	if err == nil {
		t.Error("expected error for missing bot token, got nil")
	}

	// 2. Invalid chat ID
	t.Setenv("TELEGRAM_BOT_ID", "token")
	t.Setenv("TELEGRAM_CHAT_ID", "invalid-id")
	err = SendTelegramMessage("test")
	if err == nil {
		t.Error("expected error for invalid chat ID, got nil")
	}
}

func TestHTTPStatusErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	BaseURL = server.URL

	// Send message fails on status code
	err := SendTelegramMessageToChat("token", 123, "test")
	if err == nil {
		t.Error("expected error on non-OK response status, got nil")
	}

	// Get updates fails on status code
	_, err = GetTelegramUpdates("token", 0)
	if err == nil {
		t.Error("expected error on non-OK response status, got nil")
	}
}

func TestNetworkAndJSONErrors(t *testing.T) {
	// Invalid URL/network failure
	BaseURL = "http://invalid-domain-name-that-does-not-exist"
	err := SendTelegramMessageToChat("token", 123, "test")
	if err == nil {
		t.Error("expected network error, got nil")
	}

	_, err = GetTelegramUpdates("token", 0)
	if err == nil {
		t.Error("expected network error, got nil")
	}

	// Invalid JSON decode failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	BaseURL = server.URL
	_, err = GetTelegramUpdates("token", 0)
	if err == nil {
		t.Error("expected JSON decoding error, got nil")
	}
}

func TestBuildKeyboards(t *testing.T) {
	// 1. BuildSessionsKeyboard
	sessButtons := []SessionButton{
		{ID: "emagy-260604-1117", Folder: "/git/gic/bin"},
		{ID: "other-session", Folder: "/git/other"},
	}
	markup, err := BuildSessionsKeyboard(sessButtons)
	if err != nil {
		t.Fatalf("BuildSessionsKeyboard failed: %v", err)
	}
	var decoded InlineKeyboardMarkup
	if err := json.Unmarshal([]byte(markup), &decoded); err != nil {
		t.Fatalf("failed to decode sessions markup: %v", err)
	}
	if len(decoded.InlineKeyboard) != 2 {
		t.Errorf("expected 2 keyboard rows, got %d", len(decoded.InlineKeyboard))
	}
	if decoded.InlineKeyboard[0][0].Text != "emagy-260604-1117 (bin)" {
		t.Errorf("expected text 'emagy-260604-1117 (bin)', got %q", decoded.InlineKeyboard[0][0].Text)
	}
	if decoded.InlineKeyboard[0][0].CallbackData != "show:emagy-260604-1117" {
		t.Errorf("expected callback_data 'show:emagy-260604-1117', got %q", decoded.InlineKeyboard[0][0].CallbackData)
	}

	// 2. BuildOptionsKeyboard
	optButtons := []OptionButton{
		{ID: "1", Text: "Yes"},
		{ID: "2", Text: "No"},
		{ID: "3", Text: "Long text option that will be truncated"},
	}
	optMarkup, err := BuildOptionsKeyboard("emagy-260604-1117", optButtons)
	if err != nil {
		t.Fatalf("BuildOptionsKeyboard failed: %v", err)
	}
	var optDecoded InlineKeyboardMarkup
	if err := json.Unmarshal([]byte(optMarkup), &optDecoded); err != nil {
		t.Fatalf("failed to decode options markup: %v", err)
	}
	// 3 options with 2 per row => 2 rows
	if len(optDecoded.InlineKeyboard) != 2 {
		t.Errorf("expected 2 option rows, got %d", len(optDecoded.InlineKeyboard))
	}
	if len(optDecoded.InlineKeyboard[0]) != 2 {
		t.Errorf("expected row 0 to have 2 buttons, got %d", len(optDecoded.InlineKeyboard[0]))
	}
	if len(optDecoded.InlineKeyboard[1]) != 1 {
		t.Errorf("expected row 1 to have 1 button, got %d", len(optDecoded.InlineKeyboard[1]))
	}
	if optDecoded.InlineKeyboard[0][0].Text != "1: Yes" {
		t.Errorf("expected button text '1: Yes', got %q", optDecoded.InlineKeyboard[0][0].Text)
	}
	if optDecoded.InlineKeyboard[0][0].CallbackData != "exec:emagy-260604-1117:1" {
		t.Errorf("expected callback_data 'exec:emagy-260604-1117:1', got %q", optDecoded.InlineKeyboard[0][0].CallbackData)
	}
	// Truncated text check
	expectedTruncated := "3: Long text option that will ..."
	if optDecoded.InlineKeyboard[1][0].Text != expectedTruncated {
		t.Errorf("expected button text %q, got %q", expectedTruncated, optDecoded.InlineKeyboard[1][0].Text)
	}

	// 3. BuildMenuKeyboard
	menuMarkup, err := BuildMenuKeyboard()
	if err != nil {
		t.Fatalf("BuildMenuKeyboard failed: %v", err)
	}
	var menuDecoded InlineKeyboardMarkup
	if err := json.Unmarshal([]byte(menuMarkup), &menuDecoded); err != nil {
		t.Fatalf("failed to decode menu markup: %v", err)
	}
	if len(menuDecoded.InlineKeyboard) != 3 {
		t.Errorf("expected 3 menu rows, got %d", len(menuDecoded.InlineKeyboard))
	}
	if menuDecoded.InlineKeyboard[2][0].Text != "🔄 Restart Server" {
		t.Errorf("expected text '🔄 Restart Server', got %q", menuDecoded.InlineKeyboard[2][0].Text)
	}
	if menuDecoded.InlineKeyboard[2][0].CallbackData != "menu:restart_server" {
		t.Errorf("expected callback_data 'menu:restart_server', got %q", menuDecoded.InlineKeyboard[2][0].CallbackData)
	}
}
