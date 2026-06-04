package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type gcpLogEntry struct {
	Level   LogLevel
	Message string
	Caller  string
	Time    time.Time
}

var (
	logQueue   = make(chan gcpLogEntry, 1000)
	workerOnce sync.Once
	gcpLoggingURL = "https://logging.googleapis.com/v2/entries:write"
)

type gcpLogRequest struct {
	Entries []gcpEntry `json:"entries"`
}

type gcpEntry struct {
	LogName     string      `json:"logName"`
	Resource    gcpResource `json:"resource"`
	JSONPayload gcpPayload  `json:"jsonPayload"`
	Severity    string      `json:"severity"`
	Timestamp   string      `json:"timestamp"`
}

type gcpResource struct {
	Type string `json:"type"`
}

type gcpPayload struct {
	Message string `json:"message"`
	Caller  string `json:"caller,omitempty"`
}

func queueGCPLog(level LogLevel, msg, caller string) {
	workerOnce.Do(func() {
		go gcpWorker()
	})

	entry := gcpLogEntry{
		Level:   level,
		Message: msg,
		Caller:  caller,
		Time:    time.Now(),
	}

	select {
	case logQueue <- entry:
	default:
		// Queue full, drop log or output locally
	}
}

func gcpWorker() {
	client := &http.Client{Timeout: 5 * time.Second}

	for {
		entry, ok := <-logQueue
		if !ok {
			return
		}

		entries := []gcpLogEntry{entry}
		batchSize := 20
	collectLoop:
		for len(entries) < batchSize {
			select {
			case nextEntry := <-logQueue:
				entries = append(entries, nextEntry)
			default:
				break collectLoop
			}
		}

		sendBatch(client, entries)
	}
}

func sendBatch(client *http.Client, batch []gcpLogEntry) {
	token, err := GetAccessToken()
	if err != nil {
		// Just fail silently for GCP logging
		return
	}

	logName := fmt.Sprintf("projects/%s/logs/emorr-agy-server", projID)
	reqEntries := make([]gcpEntry, len(batch))
	for i, e := range batch {
		reqEntries[i] = gcpEntry{
			LogName:  logName,
			Resource: gcpResource{Type: "global"},
			JSONPayload: gcpPayload{
				Message: e.Message,
				Caller:  e.Caller,
			},
			Severity:  string(e.Level),
			Timestamp: e.Time.UTC().Format(time.RFC3339Nano),
		}
	}

	reqBody, err := json.Marshal(gcpLogRequest{Entries: reqEntries})
	if err != nil {
		return
	}

	apiURL := gcpLoggingURL
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err == nil {
		_ = resp.Body.Close()
	}
}
