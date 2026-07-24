package sessions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/palladius/emorr-agy/internal/color"
)

// folderColor returns Cyan for symlink paths, Blue for normal dirs.
func folderColor(path string) string {
	if fi, err := os.Lstat(path); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return color.Cyan
	}
	return color.Blue
}

type ListOptions struct {
	Harness       []string
	Format        string // "short", "long", "json"
	All           bool   // If false, exclude archived sessions
	ActiveOnly    bool   // If true, show only active/running sessions
	Folder        string // If set, filter sessions by folder prefix match
	GroupByFolder bool   // If true, group output by directory/folder with emoji header
	Limit         int    // If > 0, limit output to max N sessions
}



// ListSessions retrieves and prints the classified sessions to the provided writer.
func ListSessions(w io.Writer, engine *ClassificationEngine, opts ListOptions) error {
	sessions, err := engine.Classify(opts.Harness)
	if err != nil {
		return err
	}

	// Resolve home dir for tilde substitution in folder display
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = engine.homeDir
	}

	// Classify dead sessions using transcript analysis (must happen before filtering)
	now := time.Now()
	var metas []SessionMetadata
	for _, s := range sessions {
		if s.State == StateDeadResuscitatable || s.State == StateDeadArchived {
			metas = append(metas, BuildSessionMetadata(s, engine.homeDir))
		}
	}
	superseded := DetectSupersession(metas)
	for i, s := range sessions {
		if s.State == StateDeadResuscitatable || s.State == StateDeadArchived {
			for _, m := range metas {
				if m.ID == s.ID {
					sessions[i].Classification = ClassifySession(m, superseded[s.ID], now)
					break
				}
			}
		}
	}

	// Apply visibility filters
	if opts.ActiveOnly {
		var filtered []Session
		for _, s := range sessions {
			if s.State == StateOpenTmux || s.State == StateOpenAgy || s.State == StateOpenPrivate {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	} else if !opts.All {
		// Default: show live sessions + NEEDS_RESUME + OBSOLETE (hide archived + finished)
		var filtered []Session
		for _, s := range sessions {
			if s.State == StateDeadArchived {
				continue // always hide archived in default view
			}
			if s.Classification == ClassFinished {
				continue // hide finished dead sessions in default view
			}
			filtered = append(filtered, s)
		}
		sessions = filtered
	}

	// Apply folder filter if set
	if opts.Folder != "" {
		var filtered []Session
		for _, s := range sessions {
			if IsPathMatch(s.Folder, opts.Folder) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Apply limit if set
	if opts.Limit > 0 && len(sessions) > opts.Limit {
		sessions = sessions[:opts.Limit]
	}



	format := opts.Format
	if format == "" {
		format = "short"
	}

	if opts.GroupByFolder {
		return renderGroupedSessions(w, sessions, homeDir, format)
	}

	switch format {

	case "json":
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		_, err = w.Write(data)
		if err != nil {
			return err
		}
		_, _ = w.Write([]byte("\n"))

	case "long":
		tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			color.Colorize("ST T", color.Plain),
			color.Colorize("  SESSION ID", color.Plain),
			color.Colorize("AGE", color.Plain),
			color.Colorize("DIR", color.Plain),
			color.Colorize("HARNESS", color.Plain),
			color.Colorize("WINDOWS", color.Plain),
			color.Colorize("RESUME COMMAND", color.Plain),
			color.Colorize("DESCRIPTION", color.Plain),
		)
		hasPrintedSeparator := false
		for _, s := range sessions {
			if s.State == StateDeadArchived && !hasPrintedSeparator {
				fmt.Fprintln(tw, "---\t---\t---\t---\t---\t---\t---\t---")
				hasPrintedSeparator = true
			}
			emoji := formatStatus(s)
			harnessEmoji := getEmojiForHarness(s.Harness)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, homeDir, "~")
			if folder == "" {
				folder = "-"
			}
			if len(folder) > 30 {
				folder = folder[:27] + "..."
			}

			ageColor := color.LightGray
			if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
				ageColor = color.Green
			} else if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			coloredDesc := FormatTitleAndDescription(s.Title, s.Description, s.WorktreeBranch, s.IsCron)
			statusAndHarness := fmt.Sprintf("%s %s  ", emoji, harnessEmoji)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				color.Colorize(statusAndHarness, color.Plain),
				color.Colorize(ShortenSessionID(s.ID), color.BoldWhite),
				color.Colorize(fmt.Sprintf("%3s", age), ageColor),
				color.Colorize(folder, folderColor(s.Folder)),
				color.Colorize(s.Harness, color.Plain),
				color.Colorize(strconv.Itoa(s.ProcessCount), color.Plain),
				color.Colorize(s.ResumeCommand, color.Plain),
				coloredDesc,
			)
		}
		tw.Flush()

	default: // "short" or fallback
		tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
		fmt.Fprintf(tw, "ST T  \t%s\t%s\t%s\t%s\n",
			color.Colorize("SESSION ID", color.Plain),
			color.Colorize("AGE", color.Plain),
			color.Colorize("DIR", color.Plain),
			color.Colorize("DESCRIPTION", color.Plain),
		)
		hasPrintedSeparator := false
		for _, s := range sessions {
			if s.State == StateDeadArchived && !hasPrintedSeparator {
				fmt.Fprintln(tw, "---   \t---\t---\t---\t---")
				hasPrintedSeparator = true
			}
			emoji := formatStatus(s)
			harnessEmoji := getEmojiForHarness(s.Harness)
			age := FormatAge(s.LastActivity)
			folder := strings.ReplaceAll(s.Folder, homeDir, "~")
			if folder == "" {
				folder = "-"
			}
			if len(folder) > 30 {
				folder = folder[:27] + "..."
			}

			ageColor := color.LightGray
			if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
				ageColor = color.Green
			} else if strings.Contains(age, "d") || age == "n/a" {
				ageColor = color.DarkGray
			}

			coloredDesc := FormatTitleAndDescription(s.Title, s.Description, s.WorktreeBranch, s.IsCron)

			fmt.Fprintf(tw, "%s %s  \t%s\t%s\t%s\t%s\n",
				emoji, harnessEmoji,
				color.Colorize(ShortenSessionID(s.ID), color.BoldWhite),
				color.Colorize(fmt.Sprintf("%3s", age), ageColor),
				color.Colorize(folder, folderColor(s.Folder)),
				coloredDesc,
			)
		}
		tw.Flush()
	}


	return nil
}

// IsPathMatch returns true if dir matches the filter path. An empty filter matches
// everything. The match is prefix-based: /workspace/proj1 matches /workspace/proj1/sub
// but not /workspace/proj10 (full path component matching via trailing separator).
func IsPathMatch(dir, filter string) bool {
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

func getEmojiForHarness(harness string) string {
	switch harness {
	case "gemini":
		return "♊️"
	case "agy":
		return "⬆️"
	case "ag2ui":
		return "🖥️"
	case "claude":
		return "🇫🇷"
	default:
		return "❓"
	}
}

func getEmojiForState(state SessionState) string {
	switch state {
	case StateOpenTmux:
		return "💻"
	case StateOpenAgy:
		return "🟢"
	case StateOpenPrivate:
		return "🔒"
	case StateDeadResuscitatable:
		return "💤"
	case StateDeadArchived:
		return "⚫"
	default:
		return "❓"
	}
}

func formatStatus(s Session) string {
	// For dead sessions, use 3-state classification if available
	if (s.State == StateDeadResuscitatable || s.State == StateDeadArchived) && s.Classification != "" {
		return s.Classification.ClassificationEmoji()
	}

	emoji := getEmojiForState(s.State)
	if s.State == StateOpenTmux {
		if s.AttachedClients == 0 {
			return "💤"
		}
		if s.AttachedClients > 1 {
			return fmt.Sprintf("💻(%d)", s.AttachedClients)
		}
		return "💻"
	}
	if s.State == StateOpenPrivate {
		if s.AttachedClients > 1 {
			return fmt.Sprintf("🔒(%d)", s.AttachedClients)
		}
		return "🔒"
	}
	return emoji
}

// sessionIDColor returns the ANSI color for the session ID column.
// Yellow for live sessions, BoldWhite for NEEDS_RESUME, DarkGray for finished/obsolete/archived.
func sessionIDColor(s Session) string {
	switch {
	case s.State == StateOpenTmux || s.State == StateOpenAgy || s.State == StateOpenPrivate:
		return color.Yellow
	case s.Classification == ClassNeedsResume:
		return color.BoldWhite
	default:
		return color.DarkGray
	}
}

// sessionRowColor returns the ANSI color for all non-ID columns in a row.
// DarkGray for finished/obsolete/archived, LightGray otherwise.
func sessionRowColor(s Session) string {
	if s.Classification == ClassFinished || s.Classification == ClassObsolete || s.State == StateDeadArchived {
		return color.DarkGray
	}
	return ""
}

func FormatAge(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	d := time.Since(t)
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd", days)
}

// ShortenSessionID chops UUIDs (36-char 5-part hyphenated strings) to their first two components (e.g. 1dec76de-548e).
// Non-UUID session IDs are returned as-is.
func ShortenSessionID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) == 5 && len(id) == 36 {
		return parts[0] + "-" + parts[1]
	}
	return id
}


// FormatTitleAndDescription builds colored description:
// title (with [ and ] brackets removed) in Yellow prefixed with 🏷️ (no space after emoji),
// followed by description in Cyan separated by a single space (no │ character).
func FormatTitleAndDescription(rawTitle, desc, worktreeBranch string, isCron bool) string {
	desc = strings.TrimSpace(desc)
	desc = strings.ReplaceAll(desc, "\n", " ")
	if isCron {
		desc = "🔁 " + desc
	}
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}

	cleanTitle := strings.ReplaceAll(rawTitle, "[", "")
	cleanTitle = strings.ReplaceAll(cleanTitle, "]", "")
	cleanTitle = strings.TrimSpace(cleanTitle)

	var coloredDesc string
	if cleanTitle != "" {
		coloredDesc = color.Colorize(cleanTitle, color.Yellow) + "🏷️"
		if desc != "" {
			coloredDesc += " " + color.Colorize(desc, color.Cyan)
		}
	} else {
		coloredDesc = color.Colorize(desc, color.Cyan)
	}


	if worktreeBranch != "" {
		coloredDesc = "🌳 " + color.Colorize(worktreeBranch, color.Green) + " " + coloredDesc
	}

	return coloredDesc
}

// FindGitRoot returns (repoRoot, worktreeBranch, relSubfolder) for a given directory path.
// If the path is inside a git worktree (.worktrees/<branch>/... or worktrees/<branch>/...),
// repoRoot is the base git repository root, worktreeBranch is <branch>, and relSubfolder is any path inside the worktree.
func FindGitRoot(dir string) (repoRoot string, worktreeBranch string, relSubfolder string) {
	if dir == "" {
		return "", "", ""
	}
	cleanDir := filepath.Clean(dir)
	curr := cleanDir
	var firstGit string
	var firstRel string

	for {
		gitPath := filepath.Join(curr, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			if firstGit == "" {
				firstGit = curr
				rel, err := filepath.Rel(curr, cleanDir)
				if err == nil && rel != "." {
					firstRel = rel
				} else {
					firstRel = ""
				}
			}
		}
		parent := filepath.Dir(curr)
		if parent == curr || parent == "." || parent == "/" {
			break
		}
		curr = parent
	}

	if firstGit == "" {
		return cleanDir, "", ""
	}

	parts := strings.Split(firstGit, string(filepath.Separator))
	for i, p := range parts {
		if (p == ".worktrees" || p == "worktrees") && i > 0 && i+1 < len(parts) {
			repo := strings.Join(parts[:i], string(filepath.Separator))
			branch := parts[i+1]
			rel := firstRel
			if i+2 < len(parts) {
				subInWt := strings.Join(parts[i+2:], "/")
				if firstRel != "" {
					rel = subInWt + "/" + firstRel
				} else {
					rel = subInWt
				}
			}
			return repo, branch, rel
		}
	}

	if firstRel != "" {
		relParts := strings.Split(firstRel, string(filepath.Separator))
		if len(relParts) >= 2 && (relParts[0] == ".worktrees" || relParts[0] == "worktrees") {
			branch := relParts[1]
			rel := ""
			if len(relParts) > 2 {
				rel = strings.Join(relParts[2:], "/")
			}
			return firstGit, branch, rel
		}
	}

	return firstGit, "", firstRel
}

// FormatSubfolder cleans and formats a subfolder path, capping it to max 25 chars followed by .../
func FormatSubfolder(subfolder string) string {
	if subfolder == "" {
		return ""
	}
	clean := filepath.Clean(subfolder) + "/"
	if len(clean) > 25 {
		clean = clean[:22] + ".../"
	}
	return "📁 " + color.Colorize(clean, color.Blue)
}

type FolderGroup struct {
	Folder        string    `json:"folder"`
	DisplayFolder string    `json:"display_folder"`
	LastActivity  time.Time `json:"last_activity"`
	Sessions      []Session `json:"sessions"`
}

func renderGroupedSessions(w io.Writer, sessions []Session, homeDir string, format string) error {
	var groupOrder []string
	groupsMap := make(map[string]*FolderGroup)

	for _, s := range sessions {
		repoRoot, worktreeBranch, subfolder := FindGitRoot(s.Folder)
		folderKey := repoRoot
		if worktreeBranch != "" {
			folderKey = filepath.Join(repoRoot, ".worktrees", worktreeBranch)
		}

		g, exists := groupsMap[folderKey]
		if !exists {
			displayFolder := "-"
			if folderKey != "" {
				displayFolder = strings.ReplaceAll(folderKey, homeDir, "~")
				displayFolder = "📂 " + displayFolder
			} else {
				displayFolder = "📁 (Unknown folder)"
			}
			g = &FolderGroup{
				Folder:        folderKey,
				DisplayFolder: displayFolder,
				Sessions:      []Session{},
			}
			groupsMap[folderKey] = g
			groupOrder = append(groupOrder, folderKey)
		}

		if s.LastActivity.After(g.LastActivity) {
			g.LastActivity = s.LastActivity
		}

		sCopy := s
		sCopy.WorktreeBranch = worktreeBranch
		sCopy.Subfolder = subfolder
		g.Sessions = append(g.Sessions, sCopy)
	}

	sort.Slice(groupOrder, func(i, j int) bool {
		gI := groupsMap[groupOrder[i]]
		gJ := groupsMap[groupOrder[j]]
		return gI.LastActivity.After(gJ.LastActivity)
	})

	if format == "json" {
		var outputGroups []FolderGroup
		for _, key := range groupOrder {
			outputGroups = append(outputGroups, *groupsMap[key])
		}
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(outputGroups)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	hasPrintedSeparator := false

	for _, key := range groupOrder {
		g := groupsMap[key]
		displayHeader := g.DisplayFolder
		if len(g.Sessions) > 0 && g.Sessions[0].WorktreeBranch != "" {
			displayHeader += " 🌳 " + color.Colorize(g.Sessions[0].WorktreeBranch, color.Green)
		}

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, color.Colorize(displayHeader, color.BoldWhite))

		if format == "long" {
			for _, s := range g.Sessions {
				if s.State == StateDeadArchived && !hasPrintedSeparator {
					fmt.Fprintf(tw, "    ---   \t---\t---\t---\t---\t---\t---\t---\n")
					hasPrintedSeparator = true
				}
				emoji := formatStatus(s)
				harnessEmoji := getEmojiForHarness(s.Harness)
				age := FormatAge(s.LastActivity)

				ageColor := color.LightGray
				if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
					ageColor = color.Green
				} else if strings.Contains(age, "d") || age == "n/a" {
					ageColor = color.DarkGray
				}

				subfolderStr := FormatSubfolder(s.Subfolder)

				shortID := ShortenSessionID(s.ID)
				coloredID := color.Colorize(shortID, sessionIDColor(s))
				coloredAge := color.Colorize(fmt.Sprintf("%3s", age), ageColor)
				coloredDesc := FormatTitleAndDescription(s.Title, s.Description, "", s.IsCron)

				statusAndHarness := fmt.Sprintf("    %s %s  ", emoji, harnessEmoji)

				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					statusAndHarness,
					coloredID,
					coloredAge,
					subfolderStr,
					color.Colorize(s.Harness, color.Plain),
					color.Colorize(strconv.Itoa(s.ProcessCount), color.Plain),
					color.Colorize(s.ResumeCommand, color.Plain),
					coloredDesc,
				)
			}
		} else { // short
			for _, s := range g.Sessions {
				if s.State == StateDeadArchived && !hasPrintedSeparator {
					fmt.Fprintf(tw, "    ---   \t---\t---\t---\t---\n")
					hasPrintedSeparator = true
				}
				emoji := formatStatus(s)
				harnessEmoji := getEmojiForHarness(s.Harness)
				age := FormatAge(s.LastActivity)

				ageColor := color.LightGray
				if !s.LastActivity.IsZero() && time.Since(s.LastActivity) < time.Hour {
					ageColor = color.Green
				} else if strings.Contains(age, "d") || age == "n/a" {
					ageColor = color.DarkGray
				}

				subfolderStr := FormatSubfolder(s.Subfolder)

				shortID := ShortenSessionID(s.ID)
				coloredID := color.Colorize(shortID, sessionIDColor(s))
				coloredAge := color.Colorize(fmt.Sprintf("%3s", age), ageColor)
				coloredDesc := FormatTitleAndDescription(s.Title, s.Description, "", s.IsCron)

				statusAndHarness := fmt.Sprintf("    %s %s  ", emoji, harnessEmoji)

				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					statusAndHarness,
					coloredID,
					coloredAge,
					subfolderStr,
					coloredDesc,
				)
			}
		}
		_ = tw.Flush()
	}

	return nil
}



