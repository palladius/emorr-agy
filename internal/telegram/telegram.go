package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/palladius/emorr-agy/internal/env"
)

// BaseURL can be overridden in tests to point to a mock HTTP server.
var BaseURL = "https://api.telegram.org"

type TelegramUpdateResponse struct {
	Ok     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

type TelegramUpdate struct {
	UpdateID int              `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	MessageID int            `json:"message_id"`
	Chat      TelegramChat   `json:"chat"`
	Text      string         `json:"text"`
	Voice     *TelegramVoice `json:"voice,omitempty"`
	Audio     *TelegramAudio `json:"audio,omitempty"`
}

type TelegramVoice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"`
	FileSize int    `json:"file_size"`
}

type TelegramAudio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"`
	FileSize int    `json:"file_size"`
}

type TelegramChat struct {
	ID int64 `json:"id"`
}

// SendTelegramMessage sends a message to Telegram using environment variable keys.
func SendTelegramMessage(text string) error {
	botToken := env.GetEnvWithFallback("TELEGRAM_BOT_ID", "TELEGRAM_BOT_TOKEN", "TELEGRAM_APITOKEN")
	botToken = env.CleanValue(botToken)

	chatID := env.GetEnvWithFallback("TELEGRAM_CHAT_ID", "TELEGRAM_CHANNEL_ID")
	chatID = env.CleanValue(chatID)
	if chatID == "" {
		// Fallback to Riccardo's default direct Chat ID
		chatID = "605724096"
	}

	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_ID, TELEGRAM_BOT_TOKEN or TELEGRAM_APITOKEN is not set in .env")
	}

	idVal, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID %q: %w", chatID, err)
	}

	return SendTelegramMessageToChat(botToken, idVal, text)
}

// SendTelegramMessageToChat sends a message directly to a specific chat ID.
func SendTelegramMessageToChat(botToken string, chatID int64, text string) error {
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", BaseURL, botToken)

	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {strconv.FormatInt(chatID, 10)},
		"text":       {text},
		"parse_mode": {"Markdown"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status: %s", resp.Status)
	}

	return nil
}

// GetTelegramUpdates fetches updates from the Telegram bot updates API.
func GetTelegramUpdates(botToken string, offset int) ([]TelegramUpdate, error) {
	apiURL := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d&timeout=10", BaseURL, botToken, offset)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK status: %s", resp.Status)
	}

	var updateResp TelegramUpdateResponse
	err = json.NewDecoder(resp.Body).Decode(&updateResp)
	if err != nil {
		return nil, err
	}

	return updateResp.Result, nil
}
