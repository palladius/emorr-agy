package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/palladius/emorr-agy/internal/env"
)

var GeminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s"

type TranscriptionResult struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

type GeminiTranscriber struct {
	APIKey string
}

func NewGeminiTranscriber(apiKey string) *GeminiTranscriber {
	return &GeminiTranscriber{
		APIKey: apiKey,
	}
}

// Transcribe reads the audio file, encodes it to base64, and calls the Gemini API.
func (g *GeminiTranscriber) Transcribe(audioPath, mimeType string) (*TranscriptionResult, error) {
	if g.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is empty")
	}

	fileData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file %s: %w", audioPath, err)
	}

	base64Data := base64.StdEncoding.EncodeToString(fileData)

	prompt := "Transcribe the audio text verbatim. Identify the primary language of the speech (returning a lowercase ISO language code like 'it', 'en', 'es', etc.). Return JSON matching this schema:\n{\n  \"text\": \"Transcription verbatim\",\n  \"language\": \"it\"\n}"

	reqPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"inlineData": map[string]string{
							"mimeType": mimeType,
							"data":     base64Data,
						},
					},
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	apiURL := g.getAPIURL()
	client := &http.Client{Timeout: 30 * time.Second}
	
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API returned status %s: %s", resp.Status, string(bodyBytes))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini API response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	rawJSON := geminiResp.Candidates[0].Content.Parts[0].Text

	var result TranscriptionResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse transcription response %q: %w", rawJSON, err)
	}

	return &result, nil
}

func (g *GeminiTranscriber) getAPIURL() string {
	defaultURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s"
	if GeminiAPIURL != defaultURL {
		return fmt.Sprintf(GeminiAPIURL, g.APIKey)
	}
	model := env.GetGeminiModel()
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, g.APIKey)
}

// MapLanguageToFlag maps an ISO language code to its flag emoji.
func MapLanguageToFlag(langCode string) string {
	switch strings.ToLower(strings.TrimSpace(langCode)) {
	case "it":
		return "🇮🇹"
	case "en", "gb":
		return "🇬🇧"
	case "us":
		return "🇺🇸"
	case "es":
		return "🇪🇸"
	case "fr":
		return "🇫🇷"
	case "de":
		return "🇩🇪"
	case "pt":
		return "🇵🇹"
	case "ja":
		return "🇯🇵"
	case "zh":
		return "🇨🇳"
	case "ru":
		return "🇷🇺"
	default:
		return "🌐"
	}
}
