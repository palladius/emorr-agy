package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	tokenMu     sync.Mutex
	cachedToken string
	tokenExpiry time.Time
)

type metadataToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetAccessToken retrieves a valid GCP OAuth2 access token, caching it to avoid excessive API/gcloud calls.
func GetAccessToken() (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	if cachedToken != "" && time.Now().Before(tokenExpiry) {
		return cachedToken, nil
	}

	// 1. Try Metadata Server (GCP Runtime environment)
	client := http.Client{Timeout: 500 * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err == nil {
		req.Header.Set("Metadata-Flavor", "Google")
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var tok metadataToken
			if err := json.NewDecoder(resp.Body).Decode(&tok); err == nil {
				cachedToken = tok.AccessToken
				tokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn-120) * time.Second)
				return cachedToken, nil
			}
		}
	}

	// 2. Fallback to gcloud command-line utility (local development/Derek)
	cmd := exec.Command("gcloud", "auth", "print-access-token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch access token via gcloud: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("gcloud returned empty access token")
	}

	cachedToken = token
	tokenExpiry = time.Now().Add(30 * time.Minute) // Default cache for 30 minutes
	return cachedToken, nil
}
