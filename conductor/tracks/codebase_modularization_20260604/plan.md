# Implementation Plan: Codebase Modularization

## Phase 1: Environment & Telegram Packages
- [ ] Task 1.1: Create `internal/env/` package, write tests, and implement env-lookup and quote-cleaning logic.
- [ ] Task 1.2: Create `internal/telegram/` package, write tests, and implement Telegram message wrappers.
- [ ] Task 1.3: Conductor - User Manual Verification 'Phase 1: Environment & Telegram Packages' (Protocol in workflow.md)

## Phase 2: Core Subcommand Packages
- [ ] Task 2.1: Create `internal/check/` package, write tests, and implement tmux checking.
- [ ] Task 2.2: Create `internal/status/` package, write tests, and implement status command gathering.
- [ ] Task 2.3: Create `internal/monitor/` package, write tests, and implement thread process scanning.
- [ ] Task 2.4: Create `internal/server/` package, write tests, and implement background bot receiver.
- [ ] Task 2.5: Conductor - User Manual Verification 'Phase 2: Core Subcommand Packages' (Protocol in workflow.md)

## Phase 3: Root Entrypoint & Validation
- [ ] Task 3.1: Refactor root `main.go` to import and call `internal/` packages.
- [ ] Task 3.2: Verify compile via `just build` and run all unit tests (`go test ./...`) to ensure 100% parity and passing tests.
- [ ] Task 3.3: Conductor - User Manual Verification 'Phase 3: Root Entrypoint & Validation' (Protocol in workflow.md)
