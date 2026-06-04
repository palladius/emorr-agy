package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/palladius/emorr-agy/internal/env"
)

type GeminiClassifier struct {
	APIKey  string
	homeDir string
}

func NewGeminiClassifier(apiKey, homeDir string) *GeminiClassifier {
	return &GeminiClassifier{
		APIKey:  apiKey,
		homeDir: homeDir,
	}
}

var geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s"

func (g *GeminiClassifier) Classify(sessionID string) (*LLMResult, error) {
	if g.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is empty")
	}

	// 1. Get conversation logs from SQLite
	logs, err := g.getConversationLog(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read conversation logs: %w", err)
	}

	if len(logs) == 0 {
		return &LLMResult{
			About:            "No logs available in database",
			UserInputPending: false,
			WorthResuscitate: false,
		}, nil
	}

	// Limit logs size to avoid token limit
	if len(logs) > 8000 {
		logs = logs[len(logs)-8000:]
	}

	// 2. Prepare API Request
	apiURL := g.getAPIURL()

	prompt := fmt.Sprintf(`Analyze the following conversation logs of an autonomous coding agent.
Summarize what the agent was working on in a single sentence, check if it's currently waiting on user input (e.g. asking a question, finished task, or explicitly paused), and decide if it is worth resuscitating (resuming) or if it has finished/failed and should remain dead.
Return JSON ONLY matching this schema:
{
  "about": "Single-sentence summary of the task",
  "user_input_pending": true/false,
  "worth_resuscitate": true/false
}

CONVERSATION LOGS:
%s`, logs)

	reqPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
		},
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 3. Make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Gemini API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API returned status %s", resp.Status)
	}

	// 4. Parse response
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
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini API")
	}

	rawJSON := geminiResp.Candidates[0].Content.Parts[0].Text

	var result LLMResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse classification JSON response %q: %w", rawJSON, err)
	}

	return &result, nil
}

func (g *GeminiClassifier) getConversationLog(sessionID string) (string, error) {
	dbPath := filepath.Join(g.homeDir, ".gemini/antigravity-cli/conversations", sessionID+".db")
	
	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", nil
	}

	cmd := exec.Command("sqlite3", dbPath, "select cast(step_payload as text) from steps order by idx desc limit 15")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Clean output to keep printable characters
	var cleaned []rune
	for _, r := range string(output) {
		if (r >= 32 && r <= 126) || r == '\n' || r == '\t' {
			cleaned = append(cleaned, r)
		}
	}
	return string(cleaned), nil
}

func (g *GeminiClassifier) getAPIURL() string {
	defaultURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s"
	if geminiAPIURL != defaultURL {
		return fmt.Sprintf(geminiAPIURL, g.APIKey)
	}
	model := env.GetGeminiModel()
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, g.APIKey)
}
