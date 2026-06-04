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
