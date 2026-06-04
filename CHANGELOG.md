# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
