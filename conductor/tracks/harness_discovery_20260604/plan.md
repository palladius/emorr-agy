# Implementation Plan: Harness Discovery Integration

## Phase 1: Harness Discovery Logic & Tests (v1)
- [ ] Task 1.1: Write failing unit tests for the harness discovery logic (mocking system PATH and local directories).
- [ ] Task 1.2: Implement the discovery helper `discoverHarnesses()`, checking system PATH and `~/.local/bin/`, `~/bin/`, etc.
- [ ] Task 1.3: Conductor - User Manual Verification 'Phase 1: Harness Discovery Logic & Tests (v1)' (Protocol in workflow.md)

## Phase 2: Check Command & Monitor Integration (v2)
- [ ] Task 2.1: Write failing unit tests for the updated `runCheck()` behavior showing discovered tools or warning messages.
- [ ] Task 2.2: Update the `check` subcommand logic in `main.go` to invoke `discoverHarnesses()` and print status.
- [ ] Task 2.3: Integrate `discoverHarnesses()` into `status` and `monitor` commands so they dynamically scan only active/discovered harnesses.
- [ ] Task 2.4: Conductor - User Manual Verification 'Phase 2: Check Command & Monitor Integration (v2)' (Protocol in workflow.md)

## Phase 3: Validation & Quality Gates
- [ ] Task 3.1: Run full unit tests to confirm discovery logic works correctly and covers 80%+ of statements in discovery-related code.
- [ ] Task 3.2: Verify linting and compile clean with no warnings.
- [ ] Task 3.3: Conductor - User Manual Verification 'Phase 3: Validation & Quality Gates' (Protocol in workflow.md)
