# Implementation Plan: Telegram Server Interactive Commands

## Phase 1: tmux Pane Option Parser
- [x] Task 1.1: Write unit tests for parsing choices/options from a list of tmux terminal output lines.
- [x] Task 1.2: Implement `ParsePaneOptions(lines []string) []string` in a helper package to return detected numeric options (e.g. "1", "2", "3", etc.).
- [x] Task 1.3: Conductor - User Manual Verification 'Phase 1: tmux Pane Option Parser' (Protocol in workflow.md)

## Phase 2: Telegram Server Command Handlers (`/list` & `/listall`)
- [x] Task 2.1: Write unit tests for generating inline keyboard markup from session details.
- [x] Task 2.2: Implement `/list` and `/listall` handler in the Telegram server update processor.
- [x] Task 2.3: Implement callback query handler in the Telegram server to print detail output and draw option buttons.
- [x] Task 2.4: Implement callback query handler for option selection, calling `tmux send-keys` and updating message content.
- [x] Task 2.5: Conductor - User Manual Verification 'Phase 2: Telegram Server Command Handlers (/list & /listall)' (Protocol in workflow.md)

## Phase 3: Validation & Verification
- [x] Task 3.1: Verify compile parity (`just build`), run all unit tests, ensure >80% code coverage.
- [x] Task 3.2: Manually run bot server and verify behavior with active session `emagy-260604-1117`.
- [x] Task 3.3: Conductor - User Manual Verification 'Phase 3: Validation & Verification' (Protocol in workflow.md)
