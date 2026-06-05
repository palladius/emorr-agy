# Implementation Plan: Telegram Buttons, LLM Operator, and MCP Sandboxing

## Phase 1: MCP Sandboxing Layer
- [ ] Task 1.1: Write unit tests verifying the MCP client connection, shell command execution routing, and error handling.
- [ ] Task 1.2: Implement the MCP client connector in `internal/mcp/client.go` with fallback mock capabilities.
- [ ] Task 1.3: Conductor - User Manual Verification 'Phase 1: MCP Sandboxing Layer' (Protocol in workflow.md)

## Phase 2: Telegram Button Controls
- [ ] Task 2.1: Write unit tests for the button layout helper and callback routing for Show, Resume, Revive, and Archive.
- [ ] Task 2.2: Implement inline button generation in `/list` and `/listall` outputs, and implement callback routing in `main.go`.
- [ ] Task 2.3: Conductor - User Manual Verification 'Phase 2: Telegram Button Controls' (Protocol in workflow.md)

## Phase 3: LLM Operator ("Cindy") Command
- [ ] Task 3.1: Write unit tests verifying `/ask` parsing, loading the system prompt from `etc/prompts/cindy.txt`, and resolving intents.
- [ ] Task 3.2: Implement the `/ask` command handler and template loader in `internal/operator/cindy.go`.
- [ ] Task 3.3: Conductor - User Manual Verification 'Phase 3: LLM Operator ("Cindy") Command' (Protocol in workflow.md)

## Phase 4: Integration & Verification
- [ ] Task 4.1: Integrate all components into the root `main.go` bot server daemon.
- [ ] Task 4.2: Verify compilation (`just build`), run the full test suite, and verify >80% coverage.
- [ ] Task 4.3: Conductor - User Manual Verification 'Phase 4: Integration & Verification' (Protocol in workflow.md)
