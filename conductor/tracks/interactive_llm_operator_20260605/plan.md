# Implementation Plan: Telegram Buttons, LLM Operator, and MCP Sandboxing

## Phase 1: Telegram Button Controls & Option Extraction
- [x] Task 1.1: Enhance `ParsePaneOptions` in `internal/sessions/options.go` to detect yes/no prompts (`(y/n)`, `[y/n]`, `proceed?`) in tmux terminal lines and yield `y` and `n` as options. (247868b)
- [x] Task 1.2: Write unit tests verifying yes/no option detection and standard session buttons layout. (247868b)
- [x] Task 1.3: Update Telegram message details response in `main.go` to display standard action buttons (Archive for active, Revive for dead) alongside option choices. (247868b)
- [ ] Task 1.4: Conductor - User Manual Verification 'Phase 1: Telegram Button Controls & Option Extraction' (Protocol in workflow.md)

## Phase 2: Session Revival & Callback Routing
- [x] Task 2.1: Implement callback routing in `main.go` for the `revive:<session_id>` action, which invokes `ResumeSession` to spawn the session in its correct folder. (247868b)
- [x] Task 2.2: Implement the `archive:<session_id>` callback action to archive active/dead sessions. (247868b)
- [x] Task 2.3: Write unit tests for callback action execution and directory routing during resuscitation. (247868b)
- [ ] Task 2.4: Conductor - User Manual Verification 'Phase 2: Session Revival & Callback Routing' (Protocol in workflow.md)

## Phase 3: LLM Operator ("Cindy"), New Harness Command & PIN Gate
- [ ] Task 3.1: Implement PIN gate check (`4242`) in the Telegram bot update handler in `main.go` (lock server on startup, terminate on 3 wrong attempts).
- [ ] Task 3.2: Implement `/new [harness] [query...]` command in `main.go` to spawn a new tmux session and run the harness with the interactive prompt query.
- [ ] Task 3.3: Write unit tests verifying `/ask` parsing, loading the system prompt from `etc/prompts/cindy.txt`, and resolving intents.
- [ ] Task 3.4: Implement the `/ask` command handler and template loader in `internal/operator/cindy.go`.
- [ ] Task 3.5: Conductor - User Manual Verification 'Phase 3: LLM Operator ("Cindy"), New Harness Command & PIN Gate' (Protocol in workflow.md)

## Phase 4: MCP Sandboxing Layer & Final Verification
- [ ] Task 4.1: Write unit tests verifying the MCP client connection, shell command execution routing, and error handling.
- [ ] Task 4.2: Implement the MCP client connector in `internal/mcp/client.go` with fallback mock capabilities.
- [ ] Task 4.3: Integrate all components, verify compilation (`just build`), run the full test suite, and verify >80% coverage.
- [ ] Task 4.4: Conductor - User Manual Verification 'Phase 4: MCP Sandboxing Layer & Final Verification' (Protocol in workflow.md)
