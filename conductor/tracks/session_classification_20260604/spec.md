# Specification: Harness Session Classification & Management

## Overview
This specification details the implementation of a new `sessions` subcommand hierarchy in `emorr-agy`. This command classifies tmux/orchestrator sessions into deterministic states (`🟢 Open Tmux`, `🔒 Open Private`, `⚫ Dead - Archived`, `💤 Dead - Resuscitatable`), provides clean short/long/json formatting, supports deterministic resuscitation exclusions, and features cached LLM session classification with concurrency locking.

## Functional Requirements
1. **Session Classification Engine:**
   - **Deterministic State Mapping:**
     - **Open Tmux (🟢):** Active tmux sessions matching the configured harness prefixes (e.g., `emagy-` for Antigravity, `emgem-` for Gemini CLI).
     - **Open Private (🔒):** Other active tmux sessions on the host that do not start with the harness prefixes.
     - **Dead - Resuscitatable (💤):** Sessions found in local config state directories (e.g., `~/.gemini/` or `~/.emorr-agy/`) that are NOT currently running in tmux, and do not match the exclusion patterns.
     - **Dead - Archived (⚫):** Saved sessions that are NOT running and match resuscitation exclusion patterns.
   - **Harness Standard Prefix Mapping:**
     - `emagy-` for `agy` (Antigravity)
     - `emgem-` for `gemini` (Gemini CLI)
     - `emcld-` for `claude` (Claude Code)
   - **Resuscitation Exclusions:**
     - Support the `EXCLUDE_RESUSCITATE` env var containing glob/substring patterns.
     - Default/fallback patterns: `*Chumbia*`, `*chumbawamba*`.

2. **Subcommand CLI Interface (`emorr-agy sessions`):**
   - **`emorr-agy sessions list [options]`**
     - Lists all classified sessions.
     - Filter option: `--harness <harness>` (e.g., only list sessions for `agy`).
     - Output formats:
       - **Default/Short:** A clean, colorful table showing Status Emoji, Session Name, and PWD.
       - **Long (`--long`):** An expanded table adding Session Title/Description, Last Activity timestamp, and Process Count.
       - **JSON (`--json`):** Raw structured JSON array of session objects.
   - **`emorr-agy sessions show <session_id> [options]`**
     - Shows detailed information for a specific session, including the commands required to resume/resuscitate it.
     - **LLM/Classify Flag (`--classify` / `--llm`):**
       - Triggers LLM classification on the last `N` messages of the session's log/convo context.
       - Identifies: (a) what the session is about, (b) if user input is pending, (c) if it is worth resuscitating.
       - **Locking & Concurrency:** Implement file-based PID locking (e.g., lockfile at `~/.emorr-agy/locks/<session_id>.lock`) to ensure only one process classifies a session at a time. If locked, exit with error: `another script PID <PID> is doing this, try again later`.
       - **Caching:** Cache the classification result locally in a JSON file (e.g., `~/.emorr-agy/cache/<session_id>.json`) to ensure it runs blazing fast and avoids duplicate LLM calls.

## Acceptance Criteria
- Running `emorr-agy sessions list` outputs a table with correct emojis.
- Exclusions defined via `EXCLUDE_RESUSCITATE` or fallback default patterns are respected.
- `sessions show <session_id> --classify` executes the mock or real LLM classification, caches results, and respects PID locking.
- Test coverage for classification, listing, and showing logic is >80%.
- Compilation is clean without warnings.
