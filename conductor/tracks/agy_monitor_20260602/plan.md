# Implementation Plan: `agy-monitor` Status Tool

## Phase 1: Database & Process Parsing Core
- [x] Task 1.1: Write failing unit tests for the discovery of conversation databases (reading `last_conversations.json` and mapping `.db` files).
- [x] Task 1.2: Implement database directory scanner and mapping in `main.go`.
- [x] Task 1.3: Write failing unit tests for process descriptor inspection (checking `/proc/*/fd/` for open locks).
- [x] Task 1.4: Implement the `/proc` filesystem traversal to map running `agy` PIDs to conversation IDs.
- [x] Task 1.5: Conductor - User Manual Verification 'Phase 1: Database & Process Parsing Core' (Protocol in workflow.md)

## Phase 2: Execution State Inference & Display
- [x] Task 2.1: Write failing unit tests for process state inference heuristics (discovering child processes or blocked state).
- [x] Task 2.2: Implement child process inspection and stdin block state parsing.
- [x] Task 2.3: Implement output formatting matching the requested emoji specs:
  - `🟢` (Open/Active)
  - `⚫` (Closed/Inactive)
  - `🛠️` (Tool Calling / IO)
  - `✍️` (Gemini Generating)
  - `💬` (Waiting for User)
- [x] Task 2.4: Conductor - User Manual Verification 'Phase 2: Execution State Inference & Display' (Protocol in workflow.md)

## Phase 3: Validation & Optimization
- [x] Task 3.1: Implement profiling tests to guarantee execution times remain well below 10ms.
- [x] Task 3.2: Verify no lint errors and all quality gate checks.
- [x] Task 3.3: Conductor - User Manual Verification 'Phase 3: Validation & Optimization' (Protocol in workflow.md)
