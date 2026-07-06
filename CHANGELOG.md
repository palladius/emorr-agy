# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.6] - 2026-07-05

### Added
- Added folder filtering using `--folder/-f` to the `monitor` command.
- Integrated `Description` column in the `monitor` command using initial conversation requests.
- Formatted `SESSION ID` as clickable terminal hyperlinks (OSC 8) pointing to the conversation brain directory.
- Added `--llm` and `--classify` flags to the `monitor inspect <convID>` subcommand to optionally run LLM classification.
- Fixed unit tests to match new agy harness emoji definition.

## [0.1.5] - 2026-07-05

### Added
- Multi-harness session detection: scans both `~/.gemini/antigravity-cli/` (agy) and `~/.gemini/antigravity/` (Antigravity 2.0/UI) for session discovery.
- New `SessionSource` registry for extensible harness type scanning.
- Emoji 🖥️ for `ag2ui` (Antigravity 2.0) harness type in session listings.
- Cross-source session deduplication: tmux sessions take precedence over filesystem-only entries.
- Added `--folder` / `-f` flag to `sessions list` command to filter sessions by directory path.
  - Supports `~` expansion, relative paths, and `--folder .` for the current working directory.
  - Works across all harness types (agy, ag2ui, gemini, claude).
- Exported `IsPathMatch()` function in sessions package for reuse across commands.

## [0.1.4] - 2026-06-05

### Added
- Implemented `emorr-agy resume <id>` subcommand to allow resuming or resuscitating sessions directly from the CLI.
- Automatically handles running sessions (attaching to existing tmux session) and dead sessions (spawning new tmux session and running the correct harness command for `agy` or `gemini`).
- Added robust unit testing for the `ResumeSession` command covering all execution paths.

## [0.1.3] - 2026-06-05

### Added
- Parsed and displayed the number of attached clients to tmux sessions.
- Visual status indicators: detached tmux sessions show as `💤` (asleep/detached) and attached sessions show as `💻` (or `💻(N)` if multiple clients are attached).
- Added unit tests for the tmux session attachment formatting logic.
- Loaded session descriptions (`about` field) from the cache and surfaced a `DESCRIPTION` column in the session list output.
- Filtered out archived sessions (`⚫`) from the default list to keep it clean.
- Added `-a` / `--all` flags to include archived sessions in the list output, positioned at the bottom of the list under a visual dashed divider.
- Added a fallback that automatically extracts and cleans the first user input/request from the session's `transcript.jsonl` log when no LLM classification cache is available.
- Shortened session list column headers (STATUS -> ST, TYPE -> T, DIRECTORY -> DIR) and capped directory paths to a maximum of 15 characters for a more compact tabular CLI layout.

## [0.1.2] - 2026-06-04

### Added
- Colorized session lists and thread monitor outputs (respecting the `NO_COLOR` standard).
- Display of the last 5 transcript lines in `sessions show` command (configurable via `DefaultLastLinesCount`).
- Printed application version on server startup stdout.
- Added `just version` recipe to display the current application version.
- Added `just show-logs` recipe to fetch the latest N logs from Cloud Logging for the resolved project (with support for `GCLOUD_IDENTITY` and auto-bypassing active service account impersonation).
- Added an interactive inline menu keyboard for `/start`/`/help` (with quick-select buttons to list active, list all, or restart the server).
- Added a `/restart` command and callback handler to allow clean, non-blocking restarts of the background bot server.
- Added `doc/antigravity/TECH_DETAILS.md` explaining the logical archiving mechanism.

### Changed
- Formatted `emorr-agy monitor` output as a clean table with AGE, STATE, and right-aligned DIRECTORY.
- Simplified thread monitor states to clean text tags (`TOOL`, `USER`, `CLOSED`, `WRITING`) for perfect tabular alignment.
- Corrected TMUX emoji padding to prevent tabwriter column shifts.
- Switched default Gemini model to `gemini-3.1-flash-lite` to resolve v1beta 404 errors, with support for overrides via `AUDIO_TRANSCRIPTION_GEMINI_MODEL` and `GEMINI_MODEL` env vars.
- Switched the harness emoji for `agy` from an apple (🍏) to a balloon (🎈) to represent floating/air.

## [0.1.1] - 2026-06-04

### Added
- Created `main_test.go` with unit tests for environment helpers and startup notification filters.
- Implemented `shouldSendNotification` helper function to filter Telegram startup notifications to the `server` daemon only.
- Added comprehensive manual verification plan for compilation and daemon execution.
- Added a `VERSION` file and `CHANGELOG.md` to track project version history.

### Changed
- Refactored `main.go` startup notifications to run only when starting the `server` daemon command.
- Compiled binary output path moved to `bin/emorr-agy`.
- Re-configured build environment in `justfile` to override default stale `GOROOT`.

## [0.1.0] - 2026-06-02

### Added
- Initial implementation of the `status` command displaying Hostname, Telegram configs, background daemon PIDs, and active tmux/Antigravity threads.
- Initial implementation of the `monitor` command using process/db scans and emojis to represent state.
- Initial implementation of the `server` background daemon polling Telegram for chat commands.
- Initial implementation of the `check` command validating local tmux installations and configurations.
