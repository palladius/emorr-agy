# Specification: Harness Discovery Integration

## Overview
This specification covers the implementation of harness discovery logic in `emorr-agy`. The `check` subcommand will verify if `agy` (Antigravity), `gemini` (Gemini CLI), and `claude` (Claude Code) are installed. The `monitor` and `status` subcommands will dynamically adapt to show sessions only from the discovered harnesses.

## Functional Requirements
- **Harness Detection Engine:**
  - Search for three command binaries: `agy`, `gemini`, and `claude`.
  - Scan the system `$PATH` first (using `exec.LookPath`).
  - Fallback check in standard directories: `~/.local/bin/`, `~/bin/`, and `~/.npm-global/bin/` (or standard global npm binary path).
- **Check Subcommand Update:**
  - Check each harness tool and output status:
    - `✅ <harness>: Installed at <path>`
    - `⚠️ <harness>: Not found. (Tip: Advice on installation)`
  - Keep check command passing even if a harness is missing.
- **Dynamic Daemon / Monitor Integration:**
  - Ensure the `status` and `monitor` commands only query the harnesses that are discovered on the host system.

## Acceptance Criteria
- `emorr-agy check` displays status for `agy`, `gemini`, and `claude`.
- Missing tools display warning emojis (`⚠️`) and install advice.
- Discovered paths are correct.
- Test coverage for discovery logic exceeds 80%.
