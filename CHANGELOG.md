# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2026-06-04

### Added
- Colorized session lists and thread monitor outputs (respecting the `NO_COLOR` standard).
- Display of the last 5 transcript lines in `sessions show` command (configurable via `DefaultLastLinesCount`).
- Printed application version on server startup stdout.
- Added `just version` recipe to display the current application version.
- Added `just show-logs` recipe to fetch the latest N logs from Cloud Logging for the resolved project (with support for `GCLOUD_IDENTITY` and auto-bypassing active service account impersonation).

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
