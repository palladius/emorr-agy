# Plan: Support Multiple Harnesses (Gemini CLI and Claude Code)

## Phase 1: Research and Core Harness Discovery
- [ ] Task: Design and implement harness detection logic (TDD)
    - [ ] Write failing test for detecting installed harnesses (`agy`, `gemini`, `claude`)
    - [ ] Implement path-lookup/discovery logic for the binaries
    - [ ] Verify test passes and refactor code
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Research and Core Harness Discovery' (Protocol in workflow.md)

## Phase 2: Gemini CLI Session Discovery & State Inference
- [ ] Task: Parse and scan Gemini CLI sessions (TDD)
    - [ ] Write failing test for scanning `~/.gemini/tmp/` chats directory structure
    - [ ] Implement scanner for `~/.gemini/tmp/*/chats/*.json` and parse session details
    - [ ] Verify test passes and refactor code
- [ ] Task: Infer Gemini CLI active process states (TDD)
    - [ ] Write failing test for identifying active node processes running the `gemini` CLI
    - [ ] Implement active process detection using `/proc` or `pgrep` matching workspace paths
    - [ ] Verify test passes and refactor code
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Gemini CLI Session Discovery & State Inference' (Protocol in workflow.md)

## Phase 3: Claude Code Session Discovery
- [ ] Task: Detect Claude Code sessions/processes (TDD)
    - [ ] Write failing test for detecting running `claude` CLI processes
    - [ ] Implement scanner for running `claude` processes and histories
    - [ ] Verify test passes and refactor code
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Claude Code Session Discovery' (Protocol in workflow.md)

## Phase 4: Unified CLI and Server Integration
- [ ] Task: Unify monitor output (TDD)
    - [ ] Write failing test for merging multiple harnesses' sessions into a single sorted list
    - [ ] Update `runMonitor()` and `runStatus()` to gather threads from all detected harnesses and format them
    - [ ] Verify test passes and refactor code
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Unified CLI and Server Integration' (Protocol in workflow.md)
