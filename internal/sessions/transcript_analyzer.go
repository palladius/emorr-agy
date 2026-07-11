package sessions

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

// repoPathRe matches git repo paths on both Linux and macOS.
var repoPathRe = regexp.MustCompile(`(?:/Users/[^/]+|/home/[^/]+)/git/([^/\s"'\)]+)`)

// systemNames are directory names that should be filtered out as repo names.
var systemNames = map[string]bool{
	"bin": true, "lib": true, "usr": true, "opt": true,
	"sbin": true, "etc": true, "var": true, "tmp": true,
}

// readTranscriptSteps reads all steps from a transcript.jsonl file.
func readTranscriptSteps(transcriptPath string) ([]TranscriptStep, error) {
	f, err := os.Open(transcriptPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var steps []TranscriptStep
	scanner := bufio.NewScanner(f)
	// Increase buffer for potentially large lines
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var step TranscriptStep
		if err := json.Unmarshal([]byte(line), &step); err != nil {
			continue // skip malformed lines
		}
		steps = append(steps, step)
	}
	return steps, scanner.Err()
}

// cleanUserContent extracts the user's actual request from XML-wrapped content.
// It strips tags like <USER_REQUEST>...</USER_REQUEST> and <ADDITIONAL_METADATA>...</ADDITIONAL_METADATA>.
func cleanUserContent(content string) string {
	// Remove ADDITIONAL_METADATA block entirely
	reMetadata := regexp.MustCompile(`(?s)<ADDITIONAL_METADATA>.*?</ADDITIONAL_METADATA>`)
	content = reMetadata.ReplaceAllString(content, "")

	// Extract content from USER_REQUEST tags if present
	reUserReq := regexp.MustCompile(`(?s)<USER_REQUEST>(.*?)</USER_REQUEST>`)
	if matches := reUserReq.FindStringSubmatch(content); len(matches) > 1 {
		content = matches[1]
	}

	return strings.TrimSpace(content)
}

// IsStructurallyInterrupted checks if the last step of the transcript was a
// PLANNER_RESPONSE with non-empty tool_calls, indicating the session was
// interrupted mid-execution.
func IsStructurallyInterrupted(transcriptPath string) bool {
	steps, err := readTranscriptSteps(transcriptPath)
	if err != nil || len(steps) == 0 {
		return false
	}

	last := steps[len(steps)-1]
	return last.Type == "PLANNER_RESPONSE" && len(last.ToolCalls) > 0
}

// IsExplicitQuit checks if the last USER_INPUT in the transcript was an
// explicit quit command (/quit, /exit, exit, quit, /q).
func IsExplicitQuit(transcriptPath string) bool {
	steps, err := readTranscriptSteps(transcriptPath)
	if err != nil || len(steps) == 0 {
		return false
	}

	// Find the last USER_INPUT step
	var lastUserInput *TranscriptStep
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].Type == "USER_INPUT" {
			lastUserInput = &steps[i]
			break
		}
	}
	if lastUserInput == nil {
		return false
	}

	content := cleanUserContent(lastUserInput.Content)
	content = strings.ToLower(content)

	quitCommands := map[string]bool{
		"/quit": true,
		"/exit": true,
		"exit":  true,
		"quit":  true,
		"/q":    true,
	}

	return quitCommands[content]
}

// ExtractRepoName extracts the most common repository name from file paths
// in tool call arguments and content throughout the transcript.
func ExtractRepoName(transcriptPath string) string {
	steps, err := readTranscriptSteps(transcriptPath)
	if err != nil || len(steps) == 0 {
		return ""
	}

	counts := make(map[string]int)

	for _, step := range steps {
		// Search in tool call args
		for _, tc := range step.ToolCalls {
			argsStr := string(tc.Args)
			for _, match := range repoPathRe.FindAllStringSubmatch(argsStr, -1) {
				name := match[1]
				if !systemNames[name] {
					counts[name]++
				}
			}
		}

		// Search in content
		for _, match := range repoPathRe.FindAllStringSubmatch(step.Content, -1) {
			name := match[1]
			if !systemNames[name] {
				counts[name]++
			}
		}
	}

	if len(counts) == 0 {
		return ""
	}

	// Find the most common repo name
	var bestName string
	var bestCount int
	for name, count := range counts {
		if count > bestCount {
			bestName = name
			bestCount = count
		}
	}
	return bestName
}
