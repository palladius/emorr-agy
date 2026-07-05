# Spec: Multi-Harness Session Detection (Antigravity 2.0 & Claude Code)

## Overview

The session classification engine currently only scans `~/.gemini/antigravity-cli/` for conversation databases, brain directories, and cached conversations. This means Antigravity 2.0 (UI) sessions stored under `~/.gemini/antigravity/` are completely invisible to `emorr-agy sessions list`, `emorr-agy check`, and the Telegram bot.

This track extends the `ClassificationEngine` to scan **multiple session source directories**, correctly assigning harness types and surfacing all sessions from:
1. **agy** (Antigravity CLI) — `~/.gemini/antigravity-cli/` (already working)
2. **antigravity-ui** (Antigravity 2.0) — `~/.gemini/antigravity/` (NEW)
3. **claude** (Claude Code) — tmux prefix `emcld-` (already working via tmux)

## Functional Requirements

### FR-1: Scan Antigravity 2.0 session directory
- The `ClassificationEngine.Classify()` method must scan `~/.gemini/antigravity/conversations/` for `.db` files, the same way it currently scans `~/.gemini/antigravity-cli/conversations/`.
- Sessions discovered from `~/.gemini/antigravity/` must be assigned `Harness: "antigravity-ui"`.
- The `brain/` directory under `~/.gemini/antigravity/brain/<convID>/` must be used for transcript-based description extraction, mirroring the existing CLI logic.
- **Titles**: Parse `~/.gemini/antigravity/annotations/<convID>.pbtxt` for conversation titles (plaintext protobuf format: `title:"..."`).
- **Folder mapping**: Extract workspace URIs from `~/.gemini/antigravity/agyhub_summaries_proto.pb` by parsing `file:///...` paths paired with convIDs.

### FR-2: Configurable session source registry
- Introduce a data structure (e.g., `SessionSource`) that maps harness name → base directory path (e.g., `{"agy": "~/.gemini/antigravity-cli", "antigravity-ui": "~/.gemini/antigravity/"}`).
- The classify loop should iterate over all registered sources instead of hardcoding a single path.
- This makes it trivial to add future harness types (e.g., Claude Code if it stores local sessions).

### FR-3: Deduplicate cross-source sessions
- If a session ID appears in both a tmux session AND a filesystem source, deduplicate: the tmux entry takes precedence (since it's live), and the filesystem entry is skipped.
- The existing `runningMap` deduplication logic should be extended to cover multi-source scenarios.

### FR-4: Display harness type for Antigravity 2.0
- The `HARNESS` column in `sessions list` must show `antigravity-ui` for Antigravity 2.0 sessions.
- Add an appropriate emoji for the new harness in `getEmojiForHarness()` (e.g., 🖥️ for UI).
- The `--harness antigravity-ui` filter must work to filter only Antigravity 2.0 sessions.

### FR-5: Active session detection for Antigravity 2.0
- Extend `FindActiveConvs()` to also scan `~/.gemini/antigravity/conversations/` for DB lock files, so active Antigravity 2.0 sessions show as 🟢 open_agy (or a new `open_antigravity_ui` state) rather than dead.

### FR-6: Add `--folder` filter to `sessions list`
- Currently only `emorr-agy monitor` has `--folder/-f` filtering. The `sessions list` command is missing this flag entirely.
- Port the `isPathMatch()` logic from `main.go` to `internal/sessions/list.go`.
- Add `Folder string` field to `ListOptions` and apply folder prefix matching to filter sessions.
- Support `--folder .` (current directory), `--folder ~/git/emorr-agy` (absolute), and `--folder ~` (home).
- The folder filter must work across all harness types (agy, antigravity-ui, claude).

## Non-Functional Requirements

- **Backward Compatibility**: Existing `agy`, `gemini`, and `claude` harness detection must remain unchanged.
- **Performance**: Scanning an additional directory should add negligible overhead.
- **Testability**: All new paths must be mockable via the existing `FileSystem` interface.

## Acceptance Criteria

1. `emorr-agy sessions list` shows Antigravity 2.0 sessions with harness `antigravity-ui` and the correct emoji.
2. `emorr-agy sessions list --harness antigravity-ui` filters to only Antigravity 2.0 sessions.
3. Active Antigravity 2.0 sessions (with open DB locks) show as 🟢 instead of 💀.
4. Duplicate session IDs across sources are correctly deduplicated.
5. All existing tests pass without modification.
6. New unit tests cover the multi-source scanning logic with >80% coverage.

## Out of Scope

- Claude Code local session scanning (no `~/.claude/` directory exists; Claude sessions are detected via tmux prefix only).
- Antigravity IDE sessions (`~/.gemini/antigravity-ide/`) — can be added as a follow-up track.
- Any changes to the Telegram bot commands (the bot already uses `ClassificationEngine` and will automatically benefit).
