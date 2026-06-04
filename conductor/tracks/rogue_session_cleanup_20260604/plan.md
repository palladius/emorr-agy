# Implementation Plan: Sync and Cleanup Rogue Session Changes

## Phase 1: Automated Test & Coverage Verification [checkpoint: 7bd9e96]
- [x] Task 1.1: Write/verify unit tests for daemon startup notification filter and server/status commands. c732089
- [x] Task 1.2: Run tests to verify they pass and measure coverage for `main.go`. c732089
- [x] Task 1.3: Conductor - User Manual Verification 'Phase 1: Automated Test & Coverage Verification' (Protocol in workflow.md) 7bd9e96

## Phase 2: Compilation & Execution Validation
- [ ] Task 2.1: Build the project using `just` and verify `bin/emorr-agy` is produced.
- [ ] Task 2.2: Verify running `bin/emorr-agy status` and `bin/emorr-agy check`.
- [ ] Task 2.3: Verify starting `bin/emorr-agy server` in the background and checking that the PID file `~/.emorr-agy-server.pid` is created.
- [ ] Task 2.4: Conductor - User Manual Verification 'Phase 2: Compilation & Execution Validation' (Protocol in workflow.md)

## Phase 3: Documentation Link Integration
- [ ] Task 3.1: Add links to `doc/CLI.md` and `doc/img/emorragi_architecture.png` in `README.md`.
- [ ] Task 3.2: Link the architecture diagram in the Conductor index.md registry.
- [ ] Task 3.3: Conductor - User Manual Verification 'Phase 3: Documentation Link Integration' (Protocol in workflow.md)
