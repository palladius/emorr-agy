package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/palladius/emorr-agy/internal/env"
)

// BaseURL can be overridden in tests to point to a mock HTTP server.
var BaseURL = "https://api.telegram.org"

type TelegramUpdateResponse struct {
	Ok     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

type TelegramUpdate struct {
	UpdateID      int                    `json:"update_id"`
	Message       *TelegramMessage       `json:"message,omitempty"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query,omitempty"`
}

type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    TelegramUser     `json:"from"`
	Message *TelegramMessage `json:"message,omitempty"`
	Data    string           `json:"data"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
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

// SendTelegramMessageToChatWithMarkup sends a message with inline keyboard markup.
func SendTelegramMessageToChatWithMarkup(botToken string, chatID int64, text string, replyMarkup string) error {
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", BaseURL, botToken)

	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":      {strconv.FormatInt(chatID, 10)},
		"text":         {text},
		"parse_mode":   {"Markdown"},
		"reply_markup": {replyMarkup},
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

// EditTelegramMessageText edits an existing message text and optionally updates the inline keyboard markup.
func EditTelegramMessageText(botToken string, chatID int64, messageID int, text string, replyMarkup string) error {
	apiURL := fmt.Sprintf("%s/bot%s/editMessageText", BaseURL, botToken)

	vals := url.Values{
		"chat_id":    {strconv.FormatInt(chatID, 10)},
		"message_id": {strconv.Itoa(messageID)},
		"text":       {text},
		"parse_mode": {"Markdown"},
	}
	if replyMarkup != "" {
		vals.Set("reply_markup", replyMarkup)
	}

	resp, err := http.PostForm(apiURL, vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status: %s", resp.Status)
	}

	return nil
}

// AnswerCallbackQuery answers a callback query to stop the loading spinner in Telegram.
func AnswerCallbackQuery(botToken string, callbackQueryID string, text string) error {
	apiURL := fmt.Sprintf("%s/bot%s/answerCallbackQuery", BaseURL, botToken)

	vals := url.Values{
		"callback_query_id": {callbackQueryID},
	}
	if text != "" {
		vals.Set("text", text)
	}

	resp, err := http.PostForm(apiURL, vals)
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

// SessionButton represents basic info needed to build a session button.
type SessionButton struct {
	ID     string
	Folder string
}

// OptionButton represents basic info needed to build an option button.
type OptionButton struct {
	ID   string
	Text string
}

// BuildSessionsKeyboard constructs the inline keyboard markup JSON for a list of sessions.
func BuildSessionsKeyboard(buttons []SessionButton) (string, error) {
	var keyboard [][]InlineKeyboardButton
	for _, b := range buttons {
		// Extract folder basename
		base := b.Folder
		if idx := strings.LastIndex(b.Folder, "/"); idx != -1 {
			base = b.Folder[idx+1:]
		}
		// If base is empty, use the whole folder
		if base == "" {
			base = b.Folder
		}
		text := fmt.Sprintf("%s (%s)", b.ID, base)
		keyboard = append(keyboard, []InlineKeyboardButton{
			{
				Text:         text,
				CallbackData: "show:" + b.ID,
			},
		})
	}
	markup := InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
	data, err := json.Marshal(markup)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BuildOptionsKeyboard constructs the inline keyboard markup JSON for choices in a session.
func BuildOptionsKeyboard(sessionID string, options []OptionButton) (string, error) {
	var keyboard [][]InlineKeyboardButton
	var row []InlineKeyboardButton
	for _, opt := range options {
		// Truncate text if it's too long
		text := opt.Text
		if len(text) > 30 {
			text = text[:27] + "..."
		}
		btnText := fmt.Sprintf("%s: %s", opt.ID, text)
		row = append(row, InlineKeyboardButton{
			Text:         btnText,
			CallbackData: fmt.Sprintf("exec:%s:%s", sessionID, opt.ID),
		})
		// Put 2 buttons per row for better layout
		if len(row) == 2 {
			keyboard = append(keyboard, row)
			row = []InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}
	markup := InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
	data, err := json.Marshal(markup)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BuildMenuKeyboard constructs the inline keyboard markup JSON for the main menu options.
func BuildMenuKeyboard() (string, error) {
	markup := InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{
				{
					Text:         "📋 List Active (Waiting)",
					CallbackData: "menu:list_active",
				},
			},
			{
				{
					Text:         "🗂️ List All Sessions",
					CallbackData: "menu:list_all",
				},
			},
			{
				{
					Text:         "🔄 Restart Server",
					CallbackData: "menu:restart_server",
				},
			},
		},
	}
	data, err := json.Marshal(markup)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

