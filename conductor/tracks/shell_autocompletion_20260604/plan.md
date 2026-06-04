# Implementation Plan: Shell Autocompletion

## Phase 1: Completion Command & Static Completion Scripts
- [ ] Task 1.1: Write unit tests for the static completion generation (checking bash/zsh output formats).
- [ ] Task 1.2: Implement completion template generators in `internal/completion/` package.
- [ ] Task 1.3: Conductor - User Manual Verification 'Phase 1: Completion Command & Static Completion Scripts' (Protocol in workflow.md)

## Phase 2: Dynamic Session ID Completion Support
- [ ] Task 2.1: Write unit tests for session ID autocomplete candidate generation (mocking tmux and JSON dirs).
- [ ] Task 2.2: Implement query candidate function in `internal/completion/` or `internal/sessions/` to return matching active session IDs.
- [ ] Task 2.3: Integrate candidate querying into the completion scripts (using a hidden command like `emorr-agy completion complete-session-ids`).
- [ ] Task 2.4: Conductor - User Manual Verification 'Phase 2: Dynamic Session ID Completion Support' (Protocol in workflow.md)

## Phase 3: Root CLI Integration & Verification
- [ ] Task 3.1: Register completion subcommand in `main.go`.
- [ ] Task 3.2: Verify compile parity (`just build`), run all unit tests, ensure >80% code coverage.
- [ ] Task 3.3: Conductor - User Manual Verification 'Phase 3: Root CLI Integration & Verification' (Protocol in workflow.md)
