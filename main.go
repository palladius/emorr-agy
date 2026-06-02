package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "telegram":
		if len(os.Args) < 4 || os.Args[2] != "send" {
			printUsage()
			os.Exit(1)
		}
		message := strings.Join(os.Args[3:], " ")
		err := sendTelegramMessage(message)
		if err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
		fmt.Println("🎉 Message sent successfully to Telegram!")

	case "monitor":
		err := runMonitor()
		if err != nil {
			log.Fatalf("Error running monitor: %v", err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  emorragy telegram send <message>   - Send a message to Telegram")
	fmt.Println("  emorragy monitor                   - Monitor active agy threads with emojis")
}

func sendTelegramMessage(text string) error {
	botToken := getEnvWithFallback("TELEGRAM_BOT_ID", "TELEGRAM_BOT_TOKEN", "TELEGRAM_APITOKEN")
	botToken = cleanValue(botToken)

	chatID := getEnvWithFallback("TELEGRAM_CHAT_ID", "TELEGRAM_CHANNEL_ID")
	chatID = cleanValue(chatID)
	if chatID == "" {
		// Fallback to Riccardo's default direct Chat ID
		chatID = "605724096"
	}

	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_ID, TELEGRAM_BOT_TOKEN or TELEGRAM_APITOKEN is not set in .env")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {chatID},
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

func getEnvWithFallback(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

func cleanValue(s string) string {
	s = strings.TrimSpace(s)
	// Remove enclosing single quotes
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
	}
	// Remove enclosing double quotes
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
	}
	// Strip trailing 's
	if strings.HasSuffix(s, "'s") {
		s = strings.TrimSuffix(s, "'s")
	}
	// Strip extra enclosing quotes
	s = strings.Trim(s, "\"'")
	return s
}

// --- Monitor Subcommand Logic ---

type ThreadInfo struct {
	ConvID string
	Dir    string
	IsOpen bool
	PID    int
}

func runMonitor() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	cliPath := filepath.Join(homeDir, ".gemini/antigravity-cli")
	cacheFile := filepath.Join(cliPath, "cache/last_conversations.json")

	// 1. Discover open conversations by inspecting /proc filesystem
	openConvs, err := findOpenConversations(cliPath)
	if err != nil {
		return fmt.Errorf("failed to inspect active processes: %w", err)
	}

	// 2. Read historical conversations map from cache
	var cacheConvs map[string]string
	data, err := os.ReadFile(cacheFile)
	if err == nil {
		_ = json.Unmarshal(data, &cacheConvs)
	}

	// 3. Build process tree to check for active child processes
	processTree, err := buildProcessTree()
	if err != nil {
		return fmt.Errorf("failed to build process tree: %w", err)
	}

	// 4. Merge historical and active conversations
	merged := make(map[string]*ThreadInfo)

	// Add historical ones
	for dir, convID := range cacheConvs {
		merged[convID] = &ThreadInfo{
			ConvID: convID,
			Dir:    dir,
			IsOpen: false,
		}
	}

	// Add/Override active ones
	for convID, pid := range openConvs {
		cwdPath := fmt.Sprintf("/proc/%d/cwd", pid)
		dir, err := os.Readlink(cwdPath)
		if err != nil {
			dir = "unknown (exited)"
		}

		merged[convID] = &ThreadInfo{
			ConvID: convID,
			Dir:    dir,
			IsOpen: true,
			PID:    pid,
		}
	}

	// Sort by status (Open first) and then directory
	var threads []*ThreadInfo
	for _, thread := range merged {
		threads = append(threads, thread)
	}
	sort.Slice(threads, func(i, j int) bool {
		if threads[i].IsOpen != threads[j].IsOpen {
			return threads[i].IsOpen // true (Open) comes first
		}
		return threads[i].Dir < threads[j].Dir
	})

	// 5. Output the status list
	fmt.Println("📡 Antigravity (agy) Thread Monitor:")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, thread := range threads {
		shortID := thread.ConvID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		if !thread.IsOpen {
			fmt.Printf("⚫ %s - %s [Closed]\n", shortID, thread.Dir)
			continue
		}

		// Conversation is open (🟢), now infer detailed state
		dbPath := filepath.Join(cliPath, "conversations", thread.ConvID+".db")
		stateEmoji := "✍️"
		stateDetail := "Gemini Writing"

		// A. Check for child processes (Tool Calling/IO)
		if children := processTree[thread.PID]; len(children) > 0 {
			stateEmoji = "🛠️"
			stateDetail = "Tool Calling / IO"
		} else {
			// B. Check SQLite DB for latest step status
			stepType, status, err := getLatestStep(dbPath)
			if err == nil {
				if status == 3 { // Done
					stateEmoji = "💬"
					stateDetail = "Waiting on User"
				} else if stepType > 0 { // Any tool step type in progress
					stateEmoji = "🛠️"
					stateDetail = "Running Tool"
				}
			}
		}

		fmt.Printf("🟢 %s - %s [%s %s]\n", shortID, thread.Dir, stateDetail, stateEmoji)
	}

	return nil
}

func findOpenConversations(cliPath string) (map[string]int, error) {
	openConvs := make(map[string]int)

	files, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		pidStr := file.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue // Not a PID directory
		}

		fdDir := fmt.Sprintf("/proc/%d/fd", pid)
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue // Permission denied or exited
		}

		for _, fd := range fds {
			fdPath := filepath.Join(fdDir, fd.Name())
			target, err := os.Readlink(fdPath)
			if err != nil {
				continue
			}

			if strings.Contains(target, "/.gemini/antigravity-cli/conversations/") && strings.HasSuffix(target, ".db") {
				filename := filepath.Base(target)
				convID := strings.TrimSuffix(filename, ".db")
				openConvs[convID] = pid
			}
		}
	}

	return openConvs, nil
}

func buildProcessTree() (map[int][]int, error) {
	tree := make(map[int][]int)

	files, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		pidStr := file.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		statPath := fmt.Sprintf("/proc/%d/stat", pid)
		data, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}

		content := string(data)
		lastParen := strings.LastIndex(content, ")")
		if lastParen == -1 {
			continue
		}

		afterParen := strings.TrimSpace(content[lastParen+1:])
		fields := strings.Fields(afterParen)
		if len(fields) >= 2 {
			ppidStr := fields[1]
			ppid, err := strconv.Atoi(ppidStr)
			if err == nil {
				tree[ppid] = append(tree[ppid], pid)
			}
		}
	}

	return tree, nil
}

func getLatestStep(dbPath string) (int, int, error) {
	cmd := exec.Command("sqlite3", dbPath, "select step_type, status from steps order by idx desc limit 1")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 2 {
		stepType, _ := strconv.Atoi(parts[0])
		status, _ := strconv.Atoi(parts[1])
		return stepType, status, nil
	}

	return 0, 0, nil
}
