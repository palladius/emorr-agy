package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFilePath(t *testing.T) {
	fakeResp := FileResponse{
		Ok: true,
	}
	fakeResp.Result.FileID = "voice-file-id"
	fakeResp.Result.FilePath = "voice/file_0.ogg"
	fakeResp.Result.FileSize = 12345

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botfake-token/getFile" {
			t.Errorf("expected path /botfake-token/getFile, got %q", r.URL.Path)
		}
		if r.URL.Query().Get("file_id") != "voice-file-id" {
			t.Errorf("expected file_id voice-file-id, got %q", r.URL.Query().Get("file_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(fakeResp)
	}))
	defer server.Close()

	// Override BaseURL
	oldBaseURL := BaseURL
	BaseURL = server.URL
	defer func() { BaseURL = oldBaseURL }()

	path, err := GetFilePath("fake-token", "voice-file-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != "voice/file_0.ogg" {
		t.Errorf("expected path voice/file_0.ogg, got %q", path)
	}
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botfake-token/voice/file_0.ogg" {
			t.Errorf("expected path /botfake-token/voice/file_0.ogg, got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock audio binary data"))
	}))
	defer server.Close()

	// Override FileBaseURL
	oldFileBaseURL := FileBaseURL
	FileBaseURL = server.URL
	defer func() { FileBaseURL = oldFileBaseURL }()

	tmpDir, err := os.MkdirTemp("", "telegram-audio-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	localPath := filepath.Join(tmpDir, "downloaded_voice.ogg")
	err = DownloadFile("fake-token", "voice/file_0.ogg", localPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}

	if string(data) != "mock audio binary data" {
		t.Errorf("expected 'mock audio binary data', got %q", string(data))
	}
}
