# Specification: Codebase Modularization

## Overview
This specification details the structural refactoring of the `emorr-agy` codebase. The goal is to migrate from a single monolithic `main.go` in the root directory to a highly modular structure. The primary entrypoint `main.go` will remain at the project root but will act purely as a CLI arguments parser that imports and dispatches to granular package components residing within the `internal/` directory.

## Functional Requirements
- **Directory Restructuring:**
  - Create the `internal/` package layout.
  - Sub-packages will be created under `internal/`:
    - `internal/env/`: Handles environment variable fetching, fallback parsing, and value cleaning.
    - `internal/telegram/`: Wraps the Telegram Bot API communication, sending messages, and fetching updates.
    - `internal/monitor/`: Implements Antigravity thread process scanning, sqlite step inspection, and formatting.
    - `internal/status/`: Collects and formats system, tmux, and thread statuses.
    - `internal/server/`: Implements the Telegram bot background daemon receiver loop.
    - `internal/check/`: Verifies local installations of tmux and its configuration.
- **Entrypoint (`main.go`) Refactoring:**
  - Reduce `main.go` to a thin wrapper loading `.env` using `godotenv` and executing the appropriate subcommand packages.
- **Build Configurations:**
  - Update `justfile` build script if necessary to compile correctly.

## Non-Functional Requirements
- **Parity & Correctness:** The restructuring must preserve 100% of the existing functionalities, CLI flags, outputs, and behaviors.
- **Improved Testability:** Each sub-package must have its own unit tests, allowing isolation of business logic from the main system wrapper.

## Acceptance Criteria
- Codebase builds successfully yielding `bin/emorr-agy`.
- All subcommands (`telegram`, `monitor`, `status`, `server`, `check`) execute with the same output, options, and behaviors as before.
- Statement coverage for package components under `internal/` is verified via unit tests.
