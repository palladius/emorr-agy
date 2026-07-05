package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/palladius/emorr-agy/internal/color"
	"github.com/palladius/emorr-agy/internal/gemini"
	"github.com/palladius/emorr-agy/internal/logger"
	"github.com/palladius/emorr-agy/internal/sessions"
	"github.com/palladius/emorr-agy/internal/telegram"
)

const Version = "0.1.5"

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
		err := telegram.SendTelegramMessage(message)
		if err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
		fmt.Println("🎉 Message sent successfully to Telegram!")

	case "monitor":
		// Parse flags
		watchMode := false
		onlyOpen := false
		jsonFormat := false
		sendToTelegram := false
		inspectConvID := ""
		folderFilter := ""
		useLLM := false

		for i := 2; i < len(os.Args); i++ {
			arg := os.Args[i]
			switch arg {
			case "--watch", "-w", "watch":
				watchMode = true
			case "--open", "-o", "open":
				onlyOpen = true
			case "--json", "-j", "json":
				jsonFormat = true
			case "--telegram", "-t", "telegram":
				sendToTelegram = true
			case "--llm", "--classify":
				useLLM = true
			case "--folder", "-f", "folder":
				if i+1 < len(os.Args) {
					folderFilter = os.Args[i+1]
					i++
				} else {
					log.Fatalf("Error: --folder/-f requires a directory path")
				}
			case "info", "inspect", "show":
				if i+1 < len(os.Args) {
					inspectConvID = os.Args[i+1]
					i++
				} else {
					log.Fatalf("Error: 'inspect' requires a conversation ID argument")
				}
			default:
				if !strings.HasPrefix(arg, "-") && inspectConvID == "" {
					inspectConvID = arg
				} else {
					fmt.Printf("Warning: unknown argument %s\n", arg)
				}
			}
		}

		if inspectConvID != "" {
			err := runInspect(inspectConvID, useLLM)
			if err != nil {
				log.Fatalf("Error inspecting thread: %v", err)
			}
		} else {
			err := runMonitor(watchMode, onlyOpen, jsonFormat, sendToTelegram, folderFilter)
			if err != nil {
				log.Fatalf("Error running monitor: %v", err)
			}
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

	case "sessions":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		subcommand := os.Args[2]
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home dir: %v", err)
		}

		switch subcommand {
		case "list":
			fs := flag.NewFlagSet("sessions list", flag.ExitOnError)
			var harnessFlag string
			var folderFlag string
			var jsonFlag, longFlag, shortFlag bool
			var allFlag, activeFlag bool

			fs.StringVar(&harnessFlag, "harness", "", "Filter by harness type (comma-separated list, e.g. agy,gemini)")
			fs.BoolVar(&jsonFlag, "json", false, "JSON output format")
			fs.BoolVar(&longFlag, "long", false, "Long tabular output format")
			fs.BoolVar(&shortFlag, "short", false, "Short tabular output format (default)")
			fs.BoolVar(&allFlag, "all", false, "Include archived sessions")
			fs.BoolVar(&allFlag, "a", false, "Include archived sessions (shorthand)")
			fs.BoolVar(&activeFlag, "active", false, "Show only active/running sessions")
			fs.StringVar(&folderFlag, "folder", "", "Filter by directory path (supports ~, relative, and . for cwd)")
			fs.StringVar(&folderFlag, "f", "", "Filter by directory path (shorthand)")

			_ = fs.Parse(os.Args[3:])

			format := "short"
			if jsonFlag {
				format = "json"
			} else if longFlag {
				format = "long"
			} else if shortFlag {
				format = "short"
			}

			var harnesses []string
			if harnessFlag != "" {
				parts := strings.Split(harnessFlag, ",")
				for _, p := range parts {
					harnesses = append(harnesses, strings.TrimSpace(p))
				}
			}

			// Resolve folder filter: handle ~, relative paths, and .
			resolvedFolder := ""
			if folderFlag != "" {
				path := folderFlag
				if strings.HasPrefix(path, "~") {
					path = filepath.Join(homeDir, path[1:])
				}
				abs, err := filepath.Abs(path)
				if err == nil {
					resolvedFolder = abs
				} else {
					resolvedFolder = path
				}
			}

			engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
			opts := sessions.ListOptions{
				Harness:    harnesses,
				Format:     format,
				All:        allFlag,
				ActiveOnly: activeFlag,
				Folder:     resolvedFolder,
			}
			if err := sessions.ListSessions(os.Stdout, engine, opts); err != nil {
				log.Fatalf("Error listing sessions: %v", err)
			}

		case "show":
			fs := flag.NewFlagSet("sessions show", flag.ExitOnError)
			var classifyFlag, llmFlag bool
			fs.BoolVar(&classifyFlag, "classify", false, "Enable LLM classification for session")
			fs.BoolVar(&llmFlag, "llm", false, "Enable LLM classification for session")

			_ = fs.Parse(os.Args[3:])

			positionalArgs := fs.Args()
			if len(positionalArgs) < 1 {
				fmt.Println("Error: missing session ID")
				printUsage()
				os.Exit(1)
			}
			sessionID := positionalArgs[0]

			engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
			opts := sessions.ShowOptions{
				Classify: classifyFlag || llmFlag,
			}

			apiKey := os.Getenv("GEMINI_API_KEY")
			var classifier sessions.LLMClassifier
			if opts.Classify {
				if apiKey == "" {
					log.Fatalf("Error: GEMINI_API_KEY environment variable is required when --classify or --llm is set")
				}
				classifier = sessions.NewGeminiClassifier(apiKey, homeDir)
			}

			if err := sessions.ShowSession(os.Stdout, engine, sessionID, opts, classifier); err != nil {
				log.Fatalf("Error showing session %q: %v", sessionID, err)
			}

		default:
			printUsage()
			os.Exit(1)
		}

	case "resume":
		if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}
		sessionID := os.Args[2]
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home dir: %v", err)
		}

		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		if err := sessions.ResumeSession(engine, sessionID); err != nil {
			log.Fatalf("Error resuming session %q: %v", sessionID, err)
		}

	case "ps":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home dir: %v", err)
		}
		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		if err := sessions.PrintProcessStatus(os.Stdout, engine); err != nil {
			log.Fatalf("Error listing process status: %v", err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  emorr-agy telegram send <message>   - Send a message to Telegram")
	fmt.Println("  emorr-agy monitor [flags]           - Monitor active agy threads with emojis")
	fmt.Println("  emorr-agy monitor inspect <convID> [--llm] - Inspect a specific session in detail (optional LLM classification)")
	fmt.Println("  emorr-agy status                    - Show status of system, tmux, and threads")
	fmt.Println("  emorr-agy server                    - Run the Telegram bot daemon receiver")
	fmt.Println("  emorr-agy check                     - Verify tmux installation and mouse settings")
	fmt.Println("  emorr-agy sessions list [options]   - List active and history sessions")
	fmt.Println("  emorr-agy sessions show <id> [opts] - Show session details and LLM status")
	fmt.Println("  emorr-agy resume <id>               - Resume/resuscitate a dead or active session")
	fmt.Println("  emorr-agy ps                        - Show active harness processes, CWD, and dynamic status")
	fmt.Println("\nSessions List Flags:")
	fmt.Println("  --folder, -f <path>       - Filter by directory path (supports ~, relative, . for cwd)")
	fmt.Println("  --harness <type>          - Filter by harness type (comma-separated list)")
	fmt.Println("  --all, -a                 - Include archived sessions")
	fmt.Println("  --active                  - Show only active/running sessions")
	fmt.Println("  --json / --long / --short  - Output format")
	fmt.Println("\nMonitor Flags:")
	fmt.Println("  --watch, -w, watch        - Enable continuous live watch mode")
	fmt.Println("  --open, -o, open          - Show only active/open sessions")
	fmt.Println("  --json, -j, json          - Output thread information in JSON format")
	fmt.Println("  --telegram, -t, telegram  - Send current monitor status report to Telegram")
	fmt.Println("  --folder, -f, folder       - Filter by directory path (relative or absolute)")
	fmt.Println("  --llm, --classify          - Enable LLM classification (requires GEMINI_API_KEY)")
	printFooter()
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
	ConvID       string    `json:"conv_id"`
	Dir          string    `json:"dir"`
	IsOpen       bool      `json:"is_open"`
	PID          int       `json:"pid,omitempty"`
	LastActivity time.Time `json:"-"`
	StepCount    int       `json:"step_count,omitempty"`
	StateDetail  string    `json:"state_detail"`
	StateEmoji   string    `json:"state_emoji"`
	LastActive   string    `json:"last_active"`
	Description  string    `json:"description,omitempty"`
}

func fetchThreads(cliPath string, onlyOpen bool, folderFilter string) ([]*ThreadInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	cacheFile := filepath.Join(cliPath, "cache/last_conversations.json")

	// 1. Discover open conversations by inspecting /proc filesystem
	openConvs, err := findOpenConversations(cliPath)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect active processes: %w", err)
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
		return nil, fmt.Errorf("failed to build process tree: %w", err)
	}

	// 4. Merge historical and active conversations
	merged := make(map[string]*ThreadInfo)

	// Add historical ones
	for dir, convID := range cacheConvs {
		dbPath := filepath.Join(cliPath, "conversations", convID+".db")
		var lastActivity time.Time
		if fi, err := os.Stat(dbPath); err == nil {
			lastActivity = fi.ModTime()
		} else {
			if fi, err := os.Stat(dir); err == nil {
				lastActivity = fi.ModTime()
			}
		}
		merged[convID] = &ThreadInfo{
			ConvID:       convID,
			Dir:          dir,
			IsOpen:       false,
			LastActivity: lastActivity,
			StateDetail:  "Closed",
			StateEmoji:   "⚫",
		}
	}

	// Add/Override active ones
	for convID, pid := range openConvs {
		cwdPath := fmt.Sprintf("/proc/%d/cwd", pid)
		dir, err := os.Readlink(cwdPath)
		if err != nil {
			dir = "unknown (exited)"
		}

		dbPath := filepath.Join(cliPath, "conversations", convID+".db")
		var lastActivity time.Time
		if fi, err := os.Stat(dbPath); err == nil {
			lastActivity = fi.ModTime()
		} else {
			if fi, err := os.Stat(dir); err == nil {
				lastActivity = fi.ModTime()
			}
		}

		merged[convID] = &ThreadInfo{
			ConvID:       convID,
			Dir:          dir,
			IsOpen:       true,
			PID:          pid,
			LastActivity: lastActivity,
		}
	}

	// Resolve path, expand ~ for folder filter
	var absFilter string
	if folderFilter != "" {
		path := folderFilter
		if strings.HasPrefix(path, "~") {
			path = filepath.Join(homeDir, path[1:])
		}
		abs, err := filepath.Abs(path)
		if err == nil {
			absFilter = abs
		} else {
			absFilter = path
		}
	}

	// Fill in details for all threads
	for _, thread := range merged {
		thread.Description = getTranscriptDescription(homeDir, thread.ConvID)
		dbPath := filepath.Join(cliPath, "conversations", thread.ConvID+".db")

		// Fill in last active time using file modification time
		fi, err := os.Stat(dbPath)
		if err == nil {
			diff := time.Since(fi.ModTime())
			if diff < time.Second*5 {
				thread.LastActive = "just now"
			} else if diff < time.Minute {
				thread.LastActive = fmt.Sprintf("%ds ago", int(diff.Seconds()))
			} else if diff < time.Hour {
				thread.LastActive = fmt.Sprintf("%dm ago", int(diff.Minutes()))
			} else {
				thread.LastActive = fmt.Sprintf("%dh ago", int(diff.Hours()))
			}
		} else {
			thread.LastActive = "unknown"
		}

		if !thread.IsOpen {
			continue
		}

		// Conversation is open, infer detailed state and step count
		stateEmoji := "✍️"
		stateDetail := "Gemini Writing"
		stepIdx := 0

		// A. Check for child processes (Tool Calling/IO)
		if children := processTree[thread.PID]; len(children) > 0 {
			stateEmoji = "🛠️"
			stateDetail = "Tool Calling / IO"
		}

		// B. Check SQLite DB for latest step status
		idx, stepType, status, err := getLatestStep(dbPath)
		if err == nil {
			stepIdx = idx
			if stateDetail != "Tool Calling / IO" {
				if status == 3 { // Done
					stateEmoji = "💬"
					stateDetail = "Waiting on User"
				} else if stepType > 0 {
					stateEmoji = "🛠️"
					stateDetail = "Running Tool"
				}
			}
		}

		thread.StepCount = stepIdx
		thread.StateEmoji = stateEmoji
		thread.StateDetail = stateDetail
	}

	// Sort by status (Open first) and then directory
	var threads []*ThreadInfo
	for _, thread := range merged {
		if onlyOpen && !thread.IsOpen {
			continue
		}
		if folderFilter != "" {
			if !sessions.IsPathMatch(thread.Dir, absFilter) {
				continue
			}
		}
		threads = append(threads, thread)
	}
	sort.Slice(threads, func(i, j int) bool {
		if threads[i].LastActivity.Equal(threads[j].LastActivity) {
			return threads[i].ConvID < threads[j].ConvID
		}
		return threads[i].LastActivity.After(threads[j].LastActivity)
	})

	return threads, nil
}

func isPathMatch(dir, filter string) bool {
	if filter == "" {
		return true
	}
	dir = filepath.Clean(dir)
	filter = filepath.Clean(filter)

	if dir == filter {
		return true
	}

	sep := string(filepath.Separator)
	if !strings.HasSuffix(filter, sep) {
		filter += sep
	}
	return strings.HasPrefix(dir, filter)
}

func runMonitor(watchMode, onlyOpen, jsonFormat, sendToTelegram bool, folderFilter string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	cliPath := filepath.Join(homeDir, ".gemini/antigravity-cli")

	if watchMode {
		for {
			// Clear screen
			fmt.Print("\033[H\033[2J")

			// Get current time
			nowStr := time.Now().Format("15:04:05")
			fmt.Printf("📡 Antigravity (agy) Thread Monitor (watching, Ctrl+C to exit) - %s\n", nowStr)
			fmt.Println("--------------------------------------------------------------------------------")

			threads, err := fetchThreads(cliPath, onlyOpen, folderFilter)
			if err != nil {
				fmt.Printf("Error fetching threads: %v\n", err)
			} else {
				printThreadTable(threads)
			}

			time.Sleep(2 * time.Second)
		}
	}

	threads, err := fetchThreads(cliPath, onlyOpen, folderFilter)
	if err != nil {
		return err
	}

	if jsonFormat {
		jsonData, err := json.MarshalIndent(threads, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal threads to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}

	if sendToTelegram {
		return sendMonitorToTelegram(threads)
	}

	// Default printed format
	fmt.Println("📡 Antigravity (agy) Thread Monitor:")
	fmt.Println("--------------------------------------------------------------------------------")
	printThreadTable(threads)
	return nil
}

func printThreadTable(threads []*ThreadInfo) {
	// Header
	fmt.Printf("%-4s%-12s%-6s%-24s%-30s%s\n",
		"ST",
		"SESSION ID",
		"AGE",
		"STATE",
		"DIRECTORY",
		"DESCRIPTION",
	)
	fmt.Println(strings.Repeat("-", 120))

	homeDir, _ := os.UserHomeDir()

	for _, thread := range threads {
		shortID := thread.ConvID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		// Visible fields
		statusVal := "⚫"
		statusColor := color.Plain
		if thread.IsOpen {
			statusVal = "🟢"
		}

		age := sessions.FormatAge(thread.LastActivity)
		ageColor := color.LightGray
		if strings.Contains(age, "d") || age == "n/a" {
			ageColor = color.DarkGray
		}

		stateText := "CLOSED"
		if thread.IsOpen {
			stepStr := ""
			if thread.StepCount > 0 {
				stepStr = fmt.Sprintf(" [Step %d]", thread.StepCount)
			}
			stateText = fmt.Sprintf("%s %s%s", thread.StateDetail, thread.StateEmoji, stepStr)
		}

		folder := strings.ReplaceAll(thread.Dir, "/usr/local/google/home/ricc", "~")
		// Clean and truncate folder to 28 chars
		displayFolder := folder
		if len(displayFolder) > 28 {
			displayFolder = displayFolder[:25] + "..."
		}

		desc := thread.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}

		// Pad visible fields to their column widths
		colStatus := padRight(statusVal, 4)
		colID := padRight(shortID, 12)
		colAge := padRight(age, 6)
		colState := padRight(stateText, 24)
		colFolder := padRight(displayFolder, 30)

		// Colorize/hyperlink the padded strings
		colorStatus := color.Colorize(colStatus, statusColor)

		var colorID string
		if color.ShouldColor() {
			trimmedID := thread.ConvID
			if strings.HasPrefix(thread.ConvID, "emagy-") {
				trimmedID = strings.TrimPrefix(thread.ConvID, "emagy-")
			} else if strings.HasPrefix(thread.ConvID, "emgem-") {
				trimmedID = strings.TrimPrefix(thread.ConvID, "emgem-")
			} else if strings.HasPrefix(thread.ConvID, "emcld-") {
				trimmedID = strings.TrimPrefix(thread.ConvID, "emcld-")
			}
			brainPath := filepath.Join(homeDir, ".gemini/antigravity-cli/brain", trimmedID)
			url := "file://" + brainPath
			hyperlink := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, colID)
			colorID = color.Colorize(hyperlink, color.BoldWhite)
		} else {
			colorID = color.Colorize(colID, color.BoldWhite)
		}

		colorAge := color.Colorize(colAge, ageColor)
		colorState := color.Colorize(colState, color.Plain)
		colorFolder := color.Colorize(colFolder, color.Blue)
		colorDesc := color.Colorize(desc, color.Cyan)

		fmt.Printf("%s%s%s%s%s%s\n",
			colorStatus,
			colorID,
			colorAge,
			colorState,
			colorFolder,
			colorDesc,
		)
	}
}

// padRight pads a string with spaces on the right to reach the target width.
// It handles multi-byte characters (like emojis) by counting visual columns.
func padRight(s string, width int) string {
	runes := []rune(s)
	visualWidth := 0
	for _, r := range runes {
		// Emojis typically take 2 terminal cells
		if r == '🟢' || r == '⚫' || r == '🛠' || r == '💬' || r == '✍' || r == '♊' || r == '🇫' || r == '🇷' || r == '❓' || r == '💤' || r == '💻' || r == '🔒' || r == '🖥' {
			visualWidth += 2
		} else {
			visualWidth += 1
		}
	}
	if visualWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualWidth)
}

func getTranscriptDescription(homeDir, sessionID string) string {
	trimmedID := sessionID
	if strings.HasPrefix(sessionID, "emagy-") {
		trimmedID = strings.TrimPrefix(sessionID, "emagy-")
	} else if strings.HasPrefix(sessionID, "emgem-") {
		trimmedID = strings.TrimPrefix(sessionID, "emgem-")
	} else if strings.HasPrefix(sessionID, "emcld-") {
		trimmedID = strings.TrimPrefix(sessionID, "emcld-")
	}

	paths := []string{
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain", trimmedID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain", trimmedID, ".system_generated/logs/transcript_full.jsonl"),
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript.jsonl"),
		filepath.Join(homeDir, ".gemini/antigravity-cli/brain", sessionID, ".system_generated/logs/transcript_full.jsonl"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		lines := strings.SplitN(string(data), "\n", 2)
		if len(lines) == 0 || lines[0] == "" {
			continue
		}
		var step struct {
			Source  string `json:"source"`
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(lines[0]), &step); err == nil {
			if (step.Type == "USER_INPUT" || step.Source == "USER_EXPLICIT") && step.Content != "" {
				content := step.Content
				if startIdx := strings.Index(content, "<USER_REQUEST>"); startIdx != -1 {
					content = content[startIdx+len("<USER_REQUEST>"):]
					if endIdx := strings.Index(content, "</USER_REQUEST>"); endIdx != -1 {
						content = content[:endIdx]
					}
				}
				content = strings.TrimSpace(content)
				content = strings.ReplaceAll(content, "\n", " ")
				if len(content) > 120 {
					content = content[:117] + "..."
				}
				return content
			}
		}
	}
	return ""
}

func sendMonitorToTelegram(threads []*ThreadInfo) error {
	var sb strings.Builder
	sb.WriteString("📡 *Antigravity Thread Monitor Report*\n")
	sb.WriteString("----------------------------------------\n\n")

	openCount := 0
	for _, t := range threads {
		shortID := t.ConvID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		statusStr := "Closed ⚫"
		if t.IsOpen {
			openCount++
			statusStr = fmt.Sprintf("%s %s", t.StateDetail, t.StateEmoji)
		}

		emoji := "⚫"
		if t.IsOpen {
			emoji = "🟢"
		}

		sb.WriteString(fmt.Sprintf("%s *%s* - `%s`\n", emoji, shortID, t.Dir))
		if t.IsOpen {
			sb.WriteString(fmt.Sprintf("  ↳ *Status*: %s\n", statusStr))
			if t.StepCount > 0 {
				sb.WriteString(fmt.Sprintf("  ↳ *Steps*: %d | *Active*: %s\n", t.StepCount, t.LastActive))
			} else {
				sb.WriteString(fmt.Sprintf("  ↳ *Active*: %s\n", t.LastActive))
			}
		} else {
			sb.WriteString(fmt.Sprintf("  ↳ *Status*: Closed ⚫ | *Last Active*: %s\n", t.LastActive))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("📊 *Summary*: %d Active Sessions\n", openCount))
	sb.WriteString(fmt.Sprintf("🕒 _Report generated at: %s_", time.Now().Format("2006-01-02 15:04:05")))

	return telegram.SendTelegramMessage(sb.String())
}

func runInspect(convID string, useLLM bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	cliPath := filepath.Join(homeDir, ".gemini/antigravity-cli")
	convsDir := filepath.Join(cliPath, "conversations")

	// Find the matching db file (either exact match or prefix match)
	var dbPath string
	var fullConvID string

	files, err := os.ReadDir(convsDir)
	if err != nil {
		return fmt.Errorf("failed to read conversations directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".db") {
			continue
		}
		name := strings.TrimSuffix(file.Name(), ".db")
		if name == convID || strings.HasPrefix(name, convID) {
			dbPath = filepath.Join(convsDir, file.Name())
			fullConvID = name
			break
		}
	}

	if dbPath == "" {
		return fmt.Errorf("no conversation database found matching ID: %s", convID)
	}

	fmt.Printf("🔍 Inspecting Antigravity Session: %s\n", fullConvID)
	fmt.Printf("🗄️ Database Path: %s\n", dbPath)

	fi, err := os.Stat(dbPath)
	if err == nil {
		fmt.Printf("🕒 Last Active: %s (%s ago)\n", fi.ModTime().Format("2006-01-02 15:04:05"), time.Since(fi.ModTime()).Round(time.Second))
	}

	// Get latest step info
	idx, stepType, status, err := getLatestStep(dbPath)
	if err == nil {
		fmt.Printf("📊 Latest Step: Index=%d, Type=%d, Status=%d\n", idx, stepType, status)
	}

	// Query last 5 steps in detail
	fmt.Println("\n📋 Recent Trajectory Steps:")
	fmt.Println("--------------------------------------------------------------------------------")
	cmd := exec.Command("sqlite3", dbPath, "select idx, step_type, status, has_subtrajectory from steps order by idx desc limit 5")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		fmt.Print(string(output))
	} else {
		fmt.Println("No steps found or error reading steps table.")
	}

	if useLLM {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY environment variable is required when --llm or --classify is set")
		}
		classifier := sessions.NewGeminiClassifier(apiKey, homeDir)
		res, err := classifier.Classify(fullConvID)
		if err != nil {
			return fmt.Errorf("classification failed: %w", err)
		}
		fmt.Printf("\n🧠 LLM CLASSIFICATION:\n")
		fmt.Printf("  About:               %s\n", res.About)
		fmt.Printf("  User Input Pending:  %t\n", res.UserInputPending)
		fmt.Printf("  Worth Resuscitating: %t\n", res.WorthResuscitate)
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

			if (strings.Contains(target, "/.gemini/antigravity-cli/conversations/") || strings.Contains(target, "/.gemini/antigravity/conversations/")) && strings.HasSuffix(target, ".db") {
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

func getLatestStep(dbPath string) (int, int, int, error) {
	cmd := exec.Command("sqlite3", dbPath, "select idx, step_type, status from steps order by idx desc limit 1")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 3 {
		idx, _ := strconv.Atoi(parts[0])
		stepType, _ := strconv.Atoi(parts[1])
		status, _ := strconv.Atoi(parts[2])
		return idx, stepType, status, nil
	} else if len(parts) == 2 {
		stepType, _ := strconv.Atoi(parts[0])
		status, _ := strconv.Atoi(parts[1])
		return 0, stepType, status, nil
	}

	return 0, 0, 0, nil
}

func sendStartupNotification() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	botToken := getEnvWithFallback("TELEGRAM_BOT_ID", "TELEGRAM_BOT_TOKEN", "TELEGRAM_APITOKEN")
	botToken = cleanValue(botToken)
	chatID := getEnvWithFallback("TELEGRAM_CHAT_ID", "TELEGRAM_CHANNEL_ID")
	chatID = cleanValue(chatID)
	if chatID == "" {
		chatID = "605724096"
	}
	idVal, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		idVal = 605724096
	}

	markup, err := telegram.BuildReplyKeyboard()

	lockStatus := ""
	if !serverUnlocked && !isTestMode() {
		lockStatus = "\n🔒 *Server is locked.* Please enter the PIN (`4242`) to unlock."
	}

	if err == nil && botToken != "" {
		msg := fmt.Sprintf("🟢 *Emorr-Agy v%s started on %s*%s", Version, hostname, lockStatus)
		err = telegram.SendTelegramMessageToChatWithMarkup(botToken, idVal, msg, markup)
	} else {
		msg := fmt.Sprintf("Emorr-Agy v%s started on %s%s", Version, hostname, lockStatus)
		err = telegram.SendTelegramMessage(msg)
	}
	if err != nil {
		logger.Errorf("Failed to send startup notification: %v", err)
	}
}

func shouldSendNotification(command string) bool {
	return command == "server"
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
				line = strings.ReplaceAll(line, "/usr/local/google/home/ricc", "~")
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

	printFooter()
	return nil
}

func printFooter() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	railsEnv := os.Getenv("RAILS_ENV")
	if railsEnv == "" {
		railsEnv = "development"
	}
	fmt.Println()
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("🛡️  Emorr-Agy | GH: https://github.com/palladius/emorr-agy | Version: **%s**\n", Version)
	fmt.Printf("👋 Created with ☕ for Riccardo | Host: %s | Env: %s\n", hostname, railsEnv)
}

// --- Server Subcommand Logic ---

var (
	serverUnlocked   = false
	wrongPinAttempts = 0
	logFatalf        = log.Fatalf
)

const HardcodedPin = "4242"

var isTestMode = func() bool {
	return flag.Lookup("test.v") != nil || os.Getenv("TEST_ENV") == "true"
}

func escapeShellArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

var getStatusOutputFunc = getStatusOutput
var getMonitorOutputFunc = getMonitorOutput

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

	// Initialize Logger
	projectID := getEnvWithFallback("PROJECT_ID", "GCP_PROJECT", "GOOGLE_CLOUD_PROJECT")
	if err := logger.Init(projectID); err != nil {
		log.Printf("Warning: failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Infof("Server started (v%s) with PID %d, listening to Telegram...", Version, currentPID)
	sendStartupNotification()

	offset := 0
	for {
		updates, err := telegram.GetTelegramUpdates(botToken, offset)
		if err != nil {
			logger.Errorf("Error getting updates: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			_ = processUpdate(botToken, update)
		}

		time.Sleep(1 * time.Second)
	}
}

func processUpdate(botToken string, update telegram.TelegramUpdate) error {
	if update.CallbackQuery != nil {
		if !serverUnlocked && !isTestMode() {
			_ = telegram.AnswerCallbackQuery(botToken, update.CallbackQuery.ID, "🔒 Server is locked. Enter PIN first.")
			return nil
		}
		return processCallbackQuery(botToken, *update.CallbackQuery)
	}

	if update.Message == nil {
		return nil
	}

	var text string
	chatID := update.Message.Chat.ID

	if update.Message.Voice != nil {
		logger.Infof("Received voice message from chat %d", chatID)
		transcribedText, err := downloadAndTranscribe(botToken, update.Message.Voice.FileID, "audio/ogg", chatID)
		if err != nil {
			logger.Errorf("Failed to process voice message: %v", err)
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("⚠️ Error processing voice message: %v", err))
			return err
		}
		text = transcribedText
	} else if update.Message.Audio != nil {
		logger.Infof("Received audio message from chat %d", chatID)
		mimeType := update.Message.Audio.MimeType
		if mimeType == "" {
			mimeType = "audio/mpeg"
		}
		transcribedText, err := downloadAndTranscribe(botToken, update.Message.Audio.FileID, mimeType, chatID)
		if err != nil {
			logger.Errorf("Failed to process audio message: %v", err)
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("⚠️ Error processing audio message: %v", err))
			return err
		}
		text = transcribedText
	} else {
		text = strings.TrimSpace(update.Message.Text)
	}

	if text == "" {
		return nil
	}

	// PIN Gate Check
	if !serverUnlocked && !isTestMode() {
		pinText := strings.TrimSpace(text)
		if pinText == HardcodedPin {
			serverUnlocked = true
			wrongPinAttempts = 0
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, "🔓 Server unlocked successfully! Welcome back.")
			return nil
		} else {
			wrongPinAttempts++
			if wrongPinAttempts >= 3 {
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, "❌ Too many wrong PIN attempts. Shutting down server for security!")
				time.Sleep(1 * time.Second) // wait for telegram message to send
				logFatalf("Shutting down server: 3 wrong PIN attempts")
			}
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("🔒 Server is locked. Please enter the correct PIN to unlock. (Attempt %d/3)", wrongPinAttempts))
			return nil
		}
	}

	logger.Infof("Routing command message from chat %d: %q", chatID, text)

	switch {
	case strings.HasPrefix(text, "/status") || text == "status":
		statusOutput, err := getStatusOutputFunc()
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error getting status: %v", err))
		} else {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("```\n%s\n```", statusOutput))
		}

	case strings.HasPrefix(text, "/monitor") || text == "monitor":
		monitorOutput, err := getMonitorOutputFunc()
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error getting monitor: %v", err))
		} else {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("```\n%s\n```", monitorOutput))
		}

	case strings.HasPrefix(text, "/listall") || text == "listall":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error finding home dir: %v", err))
			return err
		}
		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		allSessions, err := engine.Classify(nil)
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error listing sessions: %v", err))
			return err
		}
		limit := 5
		if len(allSessions) < limit {
			limit = len(allSessions)
		}
		selected := allSessions[:limit]

		// Format the table of all sessions (including archived)
		os.Setenv("NO_COLOR", "1")
		var buf bytes.Buffer
		_ = sessions.ListSessions(&buf, engine, sessions.ListOptions{Format: "short", All: true})
		os.Unsetenv("NO_COLOR")
		tableStr := buf.String()

		if len(selected) == 0 {
			msg := fmt.Sprintf("📁 *All Sessions:*\n```\n%s\n```\nNo sessions found.", tableStr)
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, msg)
			return nil
		}
		var btnInfos []telegram.SessionButton
		for _, s := range selected {
			btnInfos = append(btnInfos, telegram.SessionButton{
				ID:     s.ID,
				Folder: s.Folder,
			})
		}
		markup, err := telegram.BuildSessionsKeyboard(btnInfos)
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error building keyboard: %v", err))
			return err
		}
		msg := fmt.Sprintf("📁 *Last 5 Sessions:*\n```\n%s\n```", tableStr)
		err = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, msg, markup)
		if err != nil {
			logger.Errorf("Failed to send sessions list: %v", err)
		}

	case strings.HasPrefix(text, "/list") || text == "list":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error finding home dir: %v", err))
			return err
		}
		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		allSessions, err := engine.Classify(nil)
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error listing sessions: %v", err))
			return err
		}
		var filtered []sessions.Session
		for _, s := range allSessions {
			trimmedID := s.ID
			if idx := strings.Index(s.ID, "-"); idx != -1 {
				if strings.HasPrefix(s.ID, "emagy-") {
					trimmedID = strings.TrimPrefix(s.ID, "emagy-")
				} else if strings.HasPrefix(s.ID, "emgem-") {
					trimmedID = strings.TrimPrefix(s.ID, "emgem-")
				} else if strings.HasPrefix(s.ID, "emcld-") {
					trimmedID = strings.TrimPrefix(s.ID, "emcld-")
				}
			}
			activeConvs := engine.FindActiveConvs()
			pid := activeConvs[trimmedID]
			if pid == 0 {
				pid = activeConvs[s.ID]
			}
			detailedState := sessions.InferDetailedState(homeDir, s.ID, s.State, pid)
			if strings.Contains(detailedState, "Waiting on User") {
				filtered = append(filtered, s)
			}
		}

		limit := 5
		if len(filtered) < limit {
			limit = len(filtered)
		}
		selected := filtered[:limit]

		// Format the table of active sessions (excluding archived)
		os.Setenv("NO_COLOR", "1")
		var buf bytes.Buffer
		_ = sessions.ListSessions(&buf, engine, sessions.ListOptions{Format: "short", All: false})
		os.Unsetenv("NO_COLOR")
		tableStr := buf.String()

		if len(selected) == 0 {
			msg := fmt.Sprintf("💬 *Active Sessions:*\n```\n%s\n```\nNo sessions are currently waiting on human interaction.", tableStr)
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, msg)
			return nil
		}
		var btnInfos []telegram.SessionButton
		for _, s := range selected {
			btnInfos = append(btnInfos, telegram.SessionButton{
				ID:     s.ID,
				Folder: s.Folder,
			})
		}
		markup, err := telegram.BuildSessionsKeyboard(btnInfos)
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error building keyboard: %v", err))
			return err
		}
		msg := fmt.Sprintf("💬 *Sessions Pending Human Interaction:*\n```\n%s\n```", tableStr)
		err = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, msg, markup)
		if err != nil {
			logger.Errorf("Failed to send human-pending list: %v", err)
		}

	case strings.HasPrefix(text, "/new") || strings.HasPrefix(strings.ToLower(text), "new ") || strings.ToLower(text) == "new":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error finding home dir: %v", err))
			return err
		}

		var remainder string
		if strings.HasPrefix(text, "/new") {
			remainder = strings.TrimSpace(strings.TrimPrefix(text, "/new"))
		} else if strings.HasPrefix(strings.ToLower(text), "new ") {
			remainder = strings.TrimSpace(text[4:])
		}

		if remainder == "" {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, "⚠️ Usage: `/new [harness] [query...]` or `new [harness] [query...]`\ne.g., `new agy write a hello world script` or `new check if server is healthy` (defaults to `agy`).")
			return nil
		}

		parts := strings.Fields(remainder)
		harness := "agy"
		query := remainder

		if len(parts) > 0 {
			firstWord := strings.ToLower(parts[0])
			if firstWord == "agy" || firstWord == "gemini" || firstWord == "claude" {
				harness = firstWord
				query = strings.TrimSpace(strings.TrimPrefix(remainder, parts[0]))
				if query == "" {
					_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("⚠️ Please specify an initial query for the `%s` harness.", harness))
					return nil
				}
			}
		}

		prefix := "emagy-"
		if harness == "gemini" {
			prefix = "emgem-"
		} else if harness == "claude" {
			prefix = "emcld-"
		}
		sessionName := prefix + time.Now().Format("060102-150405")

		harnessCmd := fmt.Sprintf("%s -i %s", harness, escapeShellArg(query))
		cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", homeDir, harnessCmd)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("⚠️ Failed to start new tmux session: %v\nstderr: %s", err, stderr.String()))
			return nil
		}

		var btnInfos []telegram.SessionButton
		btnInfos = append(btnInfos, telegram.SessionButton{
			ID:     sessionName,
			Folder: homeDir,
		})
		markup, err := telegram.BuildSessionsKeyboard(btnInfos)
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("🚀 Spawned new tmux session `%s` running harness `%s` with query: %q", sessionName, harness, query))
			return nil
		}

		msg := fmt.Sprintf("🚀 Spawned new tmux session `%s` running harness `%s` with query: %q", sessionName, harness, query)
		err = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, msg, markup)
		if err != nil {
			logger.Errorf("Failed to send new session message: %v", err)
		}

	case strings.HasPrefix(text, "/help") || strings.HasPrefix(text, "/start") || strings.HasPrefix(text, "/menu") || text == "help" || text == "menu":
		helpMsg := "📡 *Emorr-Agy Bot Help*\n\nAvailable commands:\n• `/status` - Show system, tmux, and thread status\n• `/monitor` - Show detailed active threads\n• `/list` - Show active sessions waiting on user interaction\n• `/listall` - Show the last 5 sessions of any state\n• `/new [harness] [query...]` - Spawn a new tmux session running a harness\n• `/restart` - Restart the background bot server"
		replyMarkup, err := telegram.BuildReplyKeyboard()
		if err == nil {
			_ = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, "⌨️ Persistent keyboard registered.", replyMarkup)
		}
		markup, err := telegram.BuildMenuKeyboard()
		if err != nil {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, helpMsg)
		} else {
			_ = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, helpMsg, markup)
		}

	case strings.HasPrefix(text, "/restart") || text == "restart":
		_ = telegram.SendTelegramMessageToChat(botToken, chatID, "🔄 *Restarting the emorr-agy server...*")
		restartServer()

	case strings.ToLower(text) == "ping" || text == "/ping":
		_ = telegram.SendTelegramMessageToChat(botToken, chatID, "🏓 Pong!")

	default:
		// Unknown text/command
		helpMsg := "❌ *Command not recognized.*\n\n📡 *Emorr-Agy Bot Help*\n\nAvailable commands:\n• `/status` - Show system, tmux, and thread status\n• `/monitor` - Show detailed active threads\n• `/list` - Show active sessions waiting on user interaction\n• `/listall` - Show the last 5 sessions of any state\n• `/new [harness] [query...]` - Spawn a new tmux session running a harness\n• `/restart` - Restart the background bot server"
		replyMarkup, err := telegram.BuildReplyKeyboard()
		if err == nil {
			_ = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, "⌨️ Persistent keyboard restored.", replyMarkup)
		}
		markup, err := telegram.BuildMenuKeyboard()
		if err == nil {
			_ = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, helpMsg, markup)
		} else {
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, helpMsg)
		}
	}

	return nil
}

func restartServer() {
	pid := os.Getpid()
	execPath, err := os.Executable()
	if err != nil {
		execPath = "./bin/emorr-agy"
	}

	cmd := exec.Command("tmux", "has-session", "-t", "emorr-agy-server")
	if err := cmd.Run(); err == nil {
		restartCmd := fmt.Sprintf("sleep 1 && tmux send-keys -t emorr-agy-server C-c && sleep 1 && tmux send-keys -t emorr-agy-server '%s server' C-m", execPath)
		go func() {
			_ = exec.Command("bash", "-c", restartCmd).Start()
		}()
	} else {
		restartCmd := fmt.Sprintf("sleep 1 && %s server & sleep 0.5 && kill %d", execPath, pid)
		go func() {
			_ = exec.Command("bash", "-c", restartCmd).Start()
		}()
	}
}


func processCallbackQuery(botToken string, cb telegram.TelegramCallbackQuery) error {
	data := cb.Data
	logger.Infof("Received callback query: ID=%s, Data=%q", cb.ID, data)

	// Answer callback query immediately to stop spinner
	_ = telegram.AnswerCallbackQuery(botToken, cb.ID, "")

	if cb.Message == nil {
		return nil
	}
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if strings.HasPrefix(data, "menu:") {
		action := strings.TrimPrefix(data, "menu:")
		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		switch action {
		case "list_active":
			allSessions, err := engine.Classify(nil)
			if err != nil {
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error listing sessions: %v", err))
				return err
			}
			var filtered []sessions.Session
			for _, s := range allSessions {
				trimmedID := s.ID
				if idx := strings.Index(s.ID, "-"); idx != -1 {
					if strings.HasPrefix(s.ID, "emagy-") {
						trimmedID = strings.TrimPrefix(s.ID, "emagy-")
					} else if strings.HasPrefix(s.ID, "emgem-") {
						trimmedID = strings.TrimPrefix(s.ID, "emgem-")
					} else if strings.HasPrefix(s.ID, "emcld-") {
						trimmedID = strings.TrimPrefix(s.ID, "emcld-")
					}
				}
				activeConvs := engine.FindActiveConvs()
				pid := activeConvs[trimmedID]
				if pid == 0 {
					pid = activeConvs[s.ID]
				}
				detailedState := sessions.InferDetailedState(homeDir, s.ID, s.State, pid)
				if strings.Contains(detailedState, "Waiting on User") {
					filtered = append(filtered, s)
				}
			}
			limit := 5
			if len(filtered) < limit {
				limit = len(filtered)
			}
			selected := filtered[:limit]

			// Format the table of active sessions (excluding archived)
			os.Setenv("NO_COLOR", "1")
			var buf bytes.Buffer
			_ = sessions.ListSessions(&buf, engine, sessions.ListOptions{Format: "short", All: false})
			os.Unsetenv("NO_COLOR")
			tableStr := buf.String()

			if len(selected) == 0 {
				msg := fmt.Sprintf("💬 *Active Sessions:*\n```\n%s\n```\nNo sessions are currently waiting on human interaction.", tableStr)
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, msg)
				return nil
			}
			var btnInfos []telegram.SessionButton
			for _, s := range selected {
				btnInfos = append(btnInfos, telegram.SessionButton{
					ID:     s.ID,
					Folder: s.Folder,
				})
			}
			markup, err := telegram.BuildSessionsKeyboard(btnInfos)
			if err != nil {
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error building keyboard: %v", err))
				return err
			}
			msg := fmt.Sprintf("💬 *Sessions Pending Human Interaction:*\n```\n%s\n```", tableStr)
			_ = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, msg, markup)

		case "list_all":
			allSessions, err := engine.Classify(nil)
			if err != nil {
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error listing sessions: %v", err))
				return err
			}
			limit := 5
			if len(allSessions) < limit {
				limit = len(allSessions)
			}
			selected := allSessions[:limit]

			// Format the table of all sessions (including archived)
			os.Setenv("NO_COLOR", "1")
			var buf bytes.Buffer
			_ = sessions.ListSessions(&buf, engine, sessions.ListOptions{Format: "short", All: true})
			os.Unsetenv("NO_COLOR")
			tableStr := buf.String()

			if len(selected) == 0 {
				msg := fmt.Sprintf("📁 *All Sessions:*\n```\n%s\n```\nNo sessions found.", tableStr)
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, msg)
				return nil
			}
			var btnInfos []telegram.SessionButton
			for _, s := range selected {
				btnInfos = append(btnInfos, telegram.SessionButton{
					ID:     s.ID,
					Folder: s.Folder,
				})
			}
			markup, err := telegram.BuildSessionsKeyboard(btnInfos)
			if err != nil {
				_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("Error building keyboard: %v", err))
				return err
			}
			msg := fmt.Sprintf("📁 *Last 5 Sessions:*\n```\n%s\n```", tableStr)
			_ = telegram.SendTelegramMessageToChatWithMarkup(botToken, chatID, msg, markup)

		case "restart_server":
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, "🔄 *Restarting the emorr-agy server...*")
			restartServer()
		}
		return nil
	}

	if strings.HasPrefix(data, "show:") {
		sessionID := strings.TrimPrefix(data, "show:")
		details, opts, isDead, err := sessions.GetSessionDetailsAndOptions(homeDir, sessionID)
		if err != nil {
			_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, fmt.Sprintf("Error retrieving details for %s: %v", sessionID, err), "")
			return err
		}

		var optButtons []telegram.OptionButton
		for _, o := range opts {
			optButtons = append(optButtons, telegram.OptionButton{
				ID:   o.ID,
				Text: o.Text,
			})
		}

		markup, _ := telegram.BuildOptionsAndActionsKeyboard(sessionID, optButtons, isDead)

		_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, details, markup)
		return nil
	}

	if strings.HasPrefix(data, "exec:") {
		parts := strings.Split(data, ":")
		if len(parts) < 3 {
			return fmt.Errorf("invalid callback data: %s", data)
		}
		sessionID := parts[1]
		optionID := parts[2]

		// Execute tmux send-keys
		cmd := exec.Command("tmux", "send-keys", "-t", sessionID, optionID, "Enter")
		if err := cmd.Run(); err != nil {
			logger.Errorf("Failed to run tmux send-keys: %v", err)
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("⚠️ Failed to send key %s to tmux session %s", optionID, sessionID))
		}

		// Wait a small delay for terminal to react/draw next screen
		time.Sleep(300 * time.Millisecond)

		// Refresh the message details and options
		details, opts, isDead, err := sessions.GetSessionDetailsAndOptions(homeDir, sessionID)
		if err != nil {
			// If session finished, update message
			_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, fmt.Sprintf("Session %s finished or became unavailable.", sessionID), "")
			return nil
		}

		var optButtons []telegram.OptionButton
		for _, o := range opts {
			optButtons = append(optButtons, telegram.OptionButton{
				ID:   o.ID,
				Text: o.Text,
			})
		}

		markup, _ := telegram.BuildOptionsAndActionsKeyboard(sessionID, optButtons, isDead)

		_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, details, markup)
		return nil
	}

	if strings.HasPrefix(data, "revive:") {
		sessionID := strings.TrimPrefix(data, "revive:")
		
		_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, fmt.Sprintf("🔄 Resuscitating session %s in background...", sessionID), "")
		
		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		err := sessions.ResuscitateSession(engine, sessionID)
		if err != nil {
			logger.Errorf("Failed to resuscitate session %s: %v", sessionID, err)
			_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, fmt.Sprintf("⚠️ Failed to resuscitate session %s: %v", sessionID, err), "")
			return err
		}

		time.Sleep(500 * time.Millisecond)

		details, opts, isDead, err := sessions.GetSessionDetailsAndOptions(homeDir, sessionID)
		if err == nil {
			var optButtons []telegram.OptionButton
			for _, o := range opts {
				optButtons = append(optButtons, telegram.OptionButton{
					ID:   o.ID,
					Text: o.Text,
				})
			}
			markup, _ := telegram.BuildOptionsAndActionsKeyboard(sessionID, optButtons, isDead)
			_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, details, markup)
		} else {
			_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, fmt.Sprintf("Session %s resuscitated but details currently unavailable.", sessionID), "")
		}
		return nil
	}

	if strings.HasPrefix(data, "archive:") {
		sessionID := strings.TrimPrefix(data, "archive:")
		engine := sessions.NewClassificationEngine(sessions.RealTmuxRunner{}, sessions.OSFileSystem{}, homeDir)
		
		_ = exec.Command("tmux", "kill-session", "-t", sessionID).Run()
		if !strings.HasPrefix(sessionID, "emagy-") && !strings.HasPrefix(sessionID, "emgem-") && !strings.HasPrefix(sessionID, "emcld-") {
			_ = exec.Command("tmux", "kill-session", "-t", "emagy-"+sessionID).Run()
			_ = exec.Command("tmux", "kill-session", "-t", "emgem-"+sessionID).Run()
			_ = exec.Command("tmux", "kill-session", "-t", "emcld-"+sessionID).Run()
		}

		err := sessions.ArchiveSession(engine, sessionID)
		if err != nil {
			logger.Errorf("Failed to archive session %s: %v", sessionID, err)
			_ = telegram.SendTelegramMessageToChat(botToken, chatID, fmt.Sprintf("⚠️ Failed to archive session %s: %v", sessionID, err))
			return err
		}

		_ = telegram.EditTelegramMessageText(botToken, chatID, messageID, fmt.Sprintf("🗄️ Session %s has been successfully archived.", sessionID), "")
		return nil
	}

	return nil
}

func downloadAndTranscribe(botToken, fileID, mimeType string, chatID int64) (string, error) {
	filePath, err := telegram.GetFilePath(botToken, fileID)
	if err != nil {
		return "", fmt.Errorf("failed to get file path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home: %w", err)
	}

	ext := filepath.Ext(filePath)
	if ext == "" {
		if mimeType == "audio/ogg" {
			ext = ".ogg"
		} else {
			ext = ".mp3"
		}
	}

	localFileName := fmt.Sprintf("voice_%d_%s%s", time.Now().UnixNano(), fileID, ext)
	localPath := filepath.Join(homeDir, ".gemini/antigravity-cli/tmp", localFileName)

	err = telegram.DownloadFile(botToken, filePath, localPath)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer os.Remove(localPath)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable is empty")
	}

	transcriber := gemini.NewGeminiTranscriber(apiKey)
	res, err := transcriber.Transcribe(localPath, mimeType)
	if err != nil {
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	flag := gemini.MapLanguageToFlag(res.Language)
	replyMsg := fmt.Sprintf("%s _%s_", flag, res.Text)
	_ = telegram.SendTelegramMessageToChat(botToken, chatID, replyMsg)

	return res.Text, nil
}

func getStatusOutput() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, "status")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
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
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
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
