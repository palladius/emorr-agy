# Implementation Plan: Codebase Modularization

## Phase 1: Environment & Telegram Packages
- [x] Task 1.1: Create `internal/env/` package, write tests, and implement env-lookup and quote-cleaning logic. 9a40446
- [x] Task 1.2: Create `internal/telegram/` package, write tests, and implement Telegram message wrappers. 813f78c
- [ ] Task 1.3: Conductor - User Manual Verification 'Phase 1: Environment & Telegram Packages' (Protocol in workflow.md)

## Phase 2: Core Subcommand Packages
- [ ] Task 2.1: Create `internal/check/` package, write tests, and implement tmux checking.
- [ ] Task 2.2: Create `internal/status/` package, write tests, and implement status command gathering.
- [ ] Task 2.3: Create `internal/harnesses/` package, write tests, and implement `internal/harnesses/agy.go` for Antigravity-specific scanning logic.
- [ ] Task 2.4: Create `internal/monitor/` package, write tests, and implement thread process scanning using the modular harnesses package.
- [ ] Task 2.5: Create `internal/server/` package, write tests, and implement background bot receiver.
- [ ] Task 2.6: Conductor - User Manual Verification 'Phase 2: Core Subcommand Packages' (Protocol in workflow.md)

## Phase 3: Root Entrypoint & Validation
- [ ] Task 3.1: Refactor root `main.go` to import and call `internal/` packages.
- [ ] Task 3.2: Verify compile via `just build` and run all unit tests (`go test ./...`) to ensure 100% parity and passing tests.
- [ ] Task 3.3: Conductor - User Manual Verification 'Phase 3: Root Entrypoint & Validation' (Protocol in workflow.md)
