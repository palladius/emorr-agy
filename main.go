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
	"time"

	"github.com/joho/godotenv"
)

const Version = "0.1.1"

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

	case "status":
		err := runStatus()
		if err != nil {
			log.Fatalf("Error running status: %v", err)
		}

	case "server":
		err := runServer()
		if err != nil {
			log.Fatalf("Error running server: %v", err)
		}

	case "check":
		err := runCheck()
		if err != nil {
			log.Fatalf("Error running check: %v", err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  emorr-agy telegram send <message>   - Send a message to Telegram")
	fmt.Println("  emorr-agy monitor                   - Monitor active agy threads with emojis")
	fmt.Println("  emorr-agy status                    - Show status of system, tmux, and threads")
	fmt.Println("  emorr-agy server                    - Run the Telegram bot daemon receiver")
	fmt.Println("  emorr-agy check                     - Verify tmux installation and mouse settings")
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

func sendStartupNotification() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	msg := fmt.Sprintf("Emorr-Agy v%s started on %s", Version, hostname)
	err = sendTelegramMessage(msg)
	if err != nil {
		log.Printf("Failed to send startup notification: %v", err)
	}
}

func runStatus() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	botToken := getEnvWithFallback("TELEGRAM_BOT_ID", "TELEGRAM_BOT_TOKEN", "TELEGRAM_APITOKEN")
	botToken = cleanValue(botToken)
	telegramConfigured := "❌ Not Configured"
	if botToken != "" {
		telegramConfigured = "✅ Configured"
	}

	homeDir, homeErr := os.UserHomeDir()
	serverStatus := "❌ Not Running"
	if homeErr == nil {
		pidPath := filepath.Join(homeDir, ".emorr-agy-server.pid")
		if data, err := os.ReadFile(pidPath); err == nil {
			if oldPID, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
				if _, err := os.Stat(fmt.Sprintf("/proc/%d", oldPID)); err == nil {
					serverStatus = fmt.Sprintf("🟢 Running (PID %d)", oldPID)
				}
			}
		}
	}

	fmt.Println("📡 Emorr-Agy Status:")
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("Version:      v%s\n", Version)
	fmt.Printf("Hostname:     %s\n", hostname)
	fmt.Printf("Telegram:     %s\n", telegramConfigured)
	fmt.Printf("Server:       %s\n", serverStatus)
	fmt.Println()

	// 1. Get tmux sessions
	fmt.Println("Active tmux Sessions:")
	fmt.Println("--------------------")
	tmuxCmd := exec.Command("tmux", "list-sessions", "-F", "#S: #{?session_attached,attached,detached} (#{session_windows} windows)")
	tmuxOutput, err := tmuxCmd.Output()
	if err != nil {
		fmt.Println("  No active tmux sessions found (or tmux server not running).")
	} else {
		lines := strings.Split(strings.TrimSpace(string(tmuxOutput)), "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Printf("  • %s\n", line)
			}
		}
	}
	fmt.Println()

	// 2. Get Antigravity Thread counts
	fmt.Println("Antigravity Threads:")
	fmt.Println("-------------------")
	if homeErr == nil {
		cliPath := filepath.Join(homeDir, ".gemini/antigravity-cli")
		openConvs, err := findOpenConversations(cliPath)
		if err == nil {
			activeCount := len(openConvs)
			cacheFile := filepath.Join(cliPath, "cache/last_conversations.json")
			var cacheConvs map[string]string
			data, err := os.ReadFile(cacheFile)
			if err == nil {
				_ = json.Unmarshal(data, &cacheConvs)
			}
			closedCount := 0
			for _, convID := range cacheConvs {
				if _, ok := openConvs[convID]; !ok {
					closedCount++
				}
			}
			fmt.Printf("  🟢 %d Active Threads (monitoring)\n", activeCount)
			fmt.Printf("  ⚫ %d Closed Threads (history)\n", closedCount)
		} else {
			fmt.Println("  Failed to query active threads.")
		}
	} else {
		fmt.Println("  Failed to query active threads (home dir unavailable).")
	}

	return nil
}

// --- Server Subcommand Logic ---

type TelegramUpdateResponse struct {
	Ok     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

type TelegramUpdate struct {
	UpdateID int             `json:"update_id"`
	Message  *TelegramMessage `json:"message"`
}

type TelegramMessage struct {
	MessageID int           `json:"message_id"`
	Chat      TelegramChat  `json:"chat"`
	Text      string        `json:"text"`
}

type TelegramChat struct {
	ID int64 `json:"id"`
}

func runServer() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pidPath := filepath.Join(homeDir, ".emorr-agy-server.pid")

	// Check if already running
	if data, err := os.ReadFile(pidPath); err == nil {
		if oldPID, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if _, err := os.Stat(fmt.Sprintf("/proc/%d", oldPID)); err == nil {
				return fmt.Errorf("server is already running with PID %d", oldPID)
			}
		}
	}

	// Write current PID
	currentPID := os.Getpid()
	err = os.WriteFile(pidPath, []byte(strconv.Itoa(currentPID)), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer os.Remove(pidPath)

	botToken := getEnvWithFallback("TELEGRAM_BOT_ID", "TELEGRAM_BOT_TOKEN", "TELEGRAM_APITOKEN")
	botToken = cleanValue(botToken)
	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_ID is not configured in environment")
	}

	fmt.Printf("Server started with PID %d, listening to Telegram...\n", currentPID)
	sendStartupNotification()

	offset := 0
	for {
		updates, err := getTelegramUpdates(botToken, offset)
		if err != nil {
			log.Printf("Error getting updates: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			if update.Message == nil {
				continue
			}

			text := strings.TrimSpace(update.Message.Text)
			chatID := update.Message.Chat.ID

			log.Printf("Received message from chat %d: %q", chatID, text)

			switch {
			case strings.HasPrefix(text, "/status") || text == "status":
				statusOutput, err := getStatusOutput()
				if err != nil {
					sendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error getting status: %v", err))
				} else {
					sendTelegramMessageToChat(botToken, chatID, statusOutput)
				}

			case strings.HasPrefix(text, "/monitor") || text == "monitor":
				monitorOutput, err := getMonitorOutput()
				if err != nil {
					sendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error getting monitor: %v", err))
				} else {
					sendTelegramMessageToChat(botToken, chatID, monitorOutput)
				}

			case strings.HasPrefix(text, "/help") || strings.HasPrefix(text, "/start") || text == "help":
				helpMsg := "📡 *Emorr-Agy Bot Help*\n\nAvailable commands:\n• `/status` - Show system, tmux, and thread status\n• `/monitor` - Show detailed active threads"
				sendTelegramMessageToChat(botToken, chatID, helpMsg)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func getTelegramUpdates(botToken string, offset int) ([]TelegramUpdate, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=10", botToken, offset)
	
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

func sendTelegramMessageToChat(botToken string, chatID int64, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	
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

func getStatusOutput() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, "status")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func getMonitorOutput() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, "monitor")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func runCheck() error {
	fmt.Println("🔍 Emorr-Agy System Check:")
	fmt.Println("--------------------------------------------------------------------------------")

	// 1. Check if tmux is installed
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Println("❌ tmux: Not installed (not found in PATH)")
		return nil
	}

	versionCmd := exec.Command("tmux", "-V")
	versionOutput, _ := versionCmd.Output()
	tmuxVersion := strings.TrimSpace(string(versionOutput))
	fmt.Printf("✅ tmux: Installed at %s (%s)\n", tmuxPath, tmuxVersion)

	// 2. Check mouse and scrolling support in ~/.tmux.conf
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("⚠️  Cannot check ~/.tmux.conf: Home directory unavailable")
		return nil
	}

	tmuxConfPath := filepath.Join(homeDir, ".tmux.conf")
	confExists := true
	if _, err := os.Stat(tmuxConfPath); os.IsNotExist(err) {
		confExists = false
	}

	if !confExists {
		fmt.Println("❌ ~/.tmux.conf: File does not exist")
		fmt.Println("   👉 Tip: Create ~/.tmux.conf and add 'set -g mouse on' to enable scrolling & mouse clicks.")
		return nil
	}

	data, err := os.ReadFile(tmuxConfPath)
	if err != nil {
		fmt.Printf("❌ ~/.tmux.conf: Exists but failed to read: %v\n", err)
		return nil
	}

	lines := strings.Split(string(data), "\n")
	mouseEnabled := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "set") && strings.Contains(line, "mouse") && strings.Contains(line, "on") {
			mouseEnabled = true
			break
		}
	}

	if mouseEnabled {
		fmt.Println("✅ ~/.tmux.conf: Mouse and scrolling support is enabled ('set -g mouse on')")
	} else {
		fmt.Println("❌ ~/.tmux.conf: Mouse support is not enabled")
		fmt.Println("   👉 Tip: Add 'set -g mouse on' to ~/.tmux.conf to enable scrolling & mouse clicks.")
	}

	return nil
}
