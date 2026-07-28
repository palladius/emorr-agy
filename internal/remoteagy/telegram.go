package remoteagy

import (
	"encoding/json"
	"fmt"
	"strings"
)

// --- Telegram Keyboard Builders ---

// InlineButton represents a single Telegram inline keyboard button.
type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

// InlineKeyboardMarkup is the Telegram inline keyboard JSON structure.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineButton `json:"inline_keyboard"`
}

// BuildConciergeKeyboard builds the main concierge menu with folder buttons,
// active session resume buttons, and control options.
func BuildConciergeKeyboard(cfg *Config, activeSessions map[string]*RemoteSession) (string, error) {
	var rows [][]InlineButton

	// Section 1: New session buttons from folder config
	for _, folder := range cfg.Folders {
		sessionName := SanitizeSessionName(folder.Name)
		// Check if a session already exists for this folder
		if _, exists := activeSessions[sessionName]; exists {
			// Show as "Resume" instead of "New"
			rows = append(rows, []InlineButton{{
				Text:         fmt.Sprintf("▶️ Resume: %s", folder.Name),
				CallbackData: fmt.Sprintf("ragy:attach:%s", sessionName),
			}})
		} else {
			rows = append(rows, []InlineButton{{
				Text:         fmt.Sprintf("🚀 %s", folder.Name),
				CallbackData: fmt.Sprintf("ragy:new:%s", folder.Name),
			}})
		}
	}

	// Section 2: Active sessions not in config (orphans from previous configs)
	for name, session := range activeSessions {
		found := false
		for _, f := range cfg.Folders {
			if SanitizeSessionName(f.Name) == name {
				found = true
				break
			}
		}
		if !found {
			rows = append(rows, []InlineButton{{
				Text:         fmt.Sprintf("👻 Resume: %s", session.Folder.Name),
				CallbackData: fmt.Sprintf("ragy:attach:%s", name),
			}})
		}
	}

	// Section 3: Control row
	controlRow := []InlineButton{
		{Text: "📋 Sessions", CallbackData: "ragy:sessions"},
		{Text: "🔒 Lock", CallbackData: "ragy:lock"},
	}
	rows = append(rows, controlRow)

	markup := InlineKeyboardMarkup{InlineKeyboard: rows}
	data, err := json.Marshal(markup)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BuildAttachedKeyboard builds the control buttons shown when attached to a session.
func BuildAttachedKeyboard(sessionName string) (string, error) {
	rows := [][]InlineButton{
		// Row 1: Main actions
		{
			{Text: "🔄 Refresh", CallbackData: fmt.Sprintf("ragy:refresh:%s", sessionName)},
			{Text: "⏸️ Pause", CallbackData: fmt.Sprintf("ragy:pause:%s", sessionName)},
			{Text: "ℹ️ Info", CallbackData: fmt.Sprintf("ragy:info:%s", sessionName)},
		},
		// Row 2: Navigation
		{
			{Text: "📋 Menu", CallbackData: "ragy:menu"},
			{Text: "💀 Kill", CallbackData: fmt.Sprintf("ragy:kill:%s", sessionName)},
		},
	}

	markup := InlineKeyboardMarkup{InlineKeyboard: rows}
	data, err := json.Marshal(markup)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// --- ask_question Detection ---

// AskQuestionChoice represents a parsed option from an ask_question modal.
type AskQuestionChoice struct {
	Index int
	Text  string
}

// DetectedAskQuestion holds the parsed result of an ask_question detection.
type DetectedAskQuestion struct {
	Question string
	Choices  []AskQuestionChoice
}

// DetectAskQuestion attempts to find an ask_question modal in the tmux pane output.
// It looks for patterns like numbered options with a question prompt.
// Returns nil if no ask_question is detected.
func DetectAskQuestion(paneContent string) *DetectedAskQuestion {
	lines := strings.Split(paneContent, "\n")

	// Look for patterns that indicate a question with numbered options:
	// 1) text
	// 2) text
	// or
	// 1. text
	// 2. text
	var questionLine string
	var choices []AskQuestionChoice
	inOptions := false

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if inOptions && len(choices) >= 2 {
				// Found enough options, the line before options is the question
				if i > 0 {
					questionLine = strings.TrimSpace(lines[i])
				}
				break
			}
			continue
		}

		// Check for numbered option patterns
		if isNumberedOption(line) {
			idx, text := parseNumberedOption(line)
			if idx > 0 {
				choices = append([]AskQuestionChoice{{Index: idx, Text: text}}, choices...)
				inOptions = true
				continue
			}
		}

		// If we were in options and hit a non-option line, that's the question
		if inOptions && len(choices) >= 2 {
			questionLine = line
			break
		}

		// Reset if not in a valid pattern
		if !inOptions {
			continue
		}
	}

	if len(choices) < 2 {
		return nil
	}

	// If we didn't find a question line, use a generic one
	if questionLine == "" {
		questionLine = "Select an option:"
	}

	return &DetectedAskQuestion{
		Question: questionLine,
		Choices:  choices,
	}
}

// BuildAskQuestionKeyboard creates inline buttons for detected ask_question choices.
func BuildAskQuestionKeyboard(sessionName string, aq *DetectedAskQuestion) (string, error) {
	var rows [][]InlineButton

	for _, choice := range aq.Choices {
		// Truncate long option text for button display
		displayText := choice.Text
		if len(displayText) > 40 {
			displayText = displayText[:37] + "..."
		}

		rows = append(rows, []InlineButton{{
			Text:         fmt.Sprintf("%d. %s", choice.Index, displayText),
			CallbackData: fmt.Sprintf("ragy:answer:%s:%d", sessionName, choice.Index),
		}})
	}

	// Add a "Write-in" button for custom input
	rows = append(rows, []InlineButton{{
		Text:         "✍️ Write-in answer",
		CallbackData: fmt.Sprintf("ragy:writein:%s", sessionName),
	}})

	markup := InlineKeyboardMarkup{InlineKeyboard: rows}
	data, err := json.Marshal(markup)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// isNumberedOption checks if a line looks like a numbered option.
func isNumberedOption(line string) bool {
	if len(line) < 3 {
		return false
	}
	// Check for patterns: "1) ", "1. ", "1: "
	for i, c := range line {
		if c >= '0' && c <= '9' {
			continue
		}
		if i > 0 && (c == ')' || c == '.' || c == ':') {
			if i+1 < len(line) && line[i+1] == ' ' {
				return true
			}
		}
		break
	}
	return false
}

// parseNumberedOption extracts the index and text from a numbered option line.
func parseNumberedOption(line string) (int, string) {
	idx := 0
	i := 0
	for ; i < len(line); i++ {
		c := line[i]
		if c >= '0' && c <= '9' {
			idx = idx*10 + int(c-'0')
		} else {
			break
		}
	}
	if i < len(line) && (line[i] == ')' || line[i] == '.' || line[i] == ':') {
		i++
	}
	if i < len(line) && line[i] == ' ' {
		i++
	}
	return idx, strings.TrimSpace(line[i:])
}
