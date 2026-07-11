# Specification: Port agy-sbobinator Features to emorr-agy

## Overview
Port the key intelligent session classification and management features from the `agy-sbobinator` Ruby skill into emorr-agy's Go codebase, adding capabilities that survive reboots and provide deeper insight into session state.

## Background
The `agy-sbobinator` (located at `~/git/skillume/gemini-cli-palladius-private-goodies/skills/resume-conversations/references/agy-sbobinator`) is a 649-line Ruby script that provides rich session analysis. It has features that emorr-agy (Go) lacks, most notably the 3-state classification system that works even after a reboot.

## Functional Requirements

### FR1: 3-State Session Classification (GitHub Issue #6)
- Add `SessionClassification` type with values: `FINISHED`, `NEEDS_RESUME`, `OBSOLETE`
- Implement **structural interruption detection**: parse transcript.jsonl, check if last step was a PLANNER_RESPONSE with non-empty tool_calls (meaning a tool call started but never completed)
- Implement **explicit quit detection**: if last USER_INPUT content matches `/quit`, `/exit`, `exit`, `quit`, `/q` → classify as FINISHED
- Implement **7-day recency window**: sessions active within 7 days remain in scope
- Integrate with existing `ListSessions()` output to show 🟢/🔴/⚠️ status

### FR2: Supersession Detection (GitHub Issue #8)
- Extract repository name from tool call arguments in transcripts (regex on file paths like `/home/riccardo/git/<repo>/...`)
- Track the latest session per repo folder
- If a newer session exists for the same repo, older sessions become OBSOLETE
- Show supersession reason in `--long` mode

### FR3: Transcript Timeline Viewer (GitHub Issue #7)
- New CLI subcommand: `emorr-agy timeline <id>` (or `sessions timeline <id>`)
- Parse transcript.jsonl line by line
- Render chronological steps with emojis:
  - 👤 USER_INPUT (cleaned of XML metadata)
  - 🚀 PLANNER_RESPONSE text
  - 🐚 run_command tool calls (show CommandLine)
  - 👀 view_file tool calls (show AbsolutePath)
  - 🛠️ other tool calls (show name + args)
- Add date separators when day changes
- Show session stats at bottom (duration, step count)

### FR4: Auto-Resume (GitHub Issue #8)
- New flag: `emorr-agy resume --auto` or `emorr-agy resume --latest`
- Find latest NEEDS_RESUME session (excluding current if detectable)
- Fall back to most recent session if no NEEDS_RESUME found
- Execute `agy --conversation <uuid>` (or appropriate harness command)

### FR5: CLI Archive Command (GitHub Issue #8)
- New subcommand: `emorr-agy sessions archive [id]`
- Without ID: archive all FINISHED and OBSOLETE sessions
- With ID: archive specific session
- Move brain folder to `~/.gemini/antigravity-cli/archive/`

### FR6: Long Mode with Resume Commands
- New flag: `--long` or `-l` for `sessions list`
- Show per-session explanations: WHY the session has its status
- Show copy-pasteable resume commands (cd + agy --conversation)
- Show tmux resume commands if in tmux

## Non-Functional Requirements
- All new features must have unit tests
- The classification must be fast (< 0.5s) without LLM calls
- LLM classification remains optional (`--llm` / `--classify`)
- No breaking changes to existing CLI interface

## Out of Scope
- Multi-harness sbobinator coordination (gemini-sbobinator, claude-sbobinator)
- Cached LLM classification (can be added later)
- Session search functionality

## Acceptance Criteria
- `emorr-agy sessions list` shows 3-state classification (🟢/🔴/⚠️) for all sessions including dead ones
- `emorr-agy timeline <id>` renders a human-readable timeline
- `emorr-agy resume --auto` finds and launches the right session
- `emorr-agy sessions archive` moves finished sessions
- All new code has > 80% test coverage
- Existing tests pass
