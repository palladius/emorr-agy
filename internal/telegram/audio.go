package telegram

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var FileBaseURL = "https://api.telegram.org/file"

type FileResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		FileID   string `json:"file_id"`
		FilePath string `json:"file_path"`
		FileSize int    `json:"file_size"`
	} `json:"result"`
}

// GetFilePath retrieves the Telegram file path for a given fileID.
func GetFilePath(botToken, fileID string) (string, error) {
	apiURL := fmt.Sprintf("%s/bot%s/getFile?file_id=%s", BaseURL, botToken, fileID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getFile returned status: %s", resp.Status)
	}

	var fileResp FileResponse
	err = json.NewDecoder(resp.Body).Decode(&fileResp)
	if err != nil {
		return "", err
	}

	if !fileResp.Ok {
		return "", fmt.Errorf("getFile response not OK")
	}

	return fileResp.Result.FilePath, nil
}

// DownloadFile downloads a file from Telegram's servers to a local path.
func DownloadFile(botToken, telegramFilePath, localPath string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	apiURL := fmt.Sprintf("%s/bot%s/%s", FileBaseURL, botToken, telegramFilePath)

	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
