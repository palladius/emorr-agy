# Specification: Verbose Process Tree in `sessions list`

## Overview
Add a `--verbose` (`-v`) flag to `emorr-agy sessions list` that shows child processes of each active agy session as an indented tree under the session row. PIDs are displayed in purple/magenta for alive processes.

## Context / Investigation Results
Live investigation on 2026-07-07 confirmed:
- All `run_command` tool calls from agy (including subagent-spawned ones) become **direct OS child processes** of the agy PID.
- Every agy process spawns 3 **MCP sidecar** children at startup (hotels-mcp, nanobanana, google-workspace). These are infrastructure, not "work".
- Work children (e.g. `ollama pull`, `bash -c`, `git`) spawn later and are the interesting ones.
- The parenthood chain is discoverable via `/proc/<pid>/stat` (PPID field).

## Functional Requirements

### FR1: `--verbose` / `-v` flag
- Add `--verbose` and `-v` boolean flags to `sessions list` subcommand.
- When enabled, after each active session row (states: `open_tmux`, `open_agy`, `open_private`), print indented child process lines.

### FR2: Child process discovery
- For each active session, resolve the agy PID (already available from `FindActiveConvs()`).
- Walk `/proc` to find all processes whose PPID equals the agy PID.
- For each child, read `/proc/<pid>/cmdline` and `/proc/<pid>/stat` (for elapsed time).

### FR3: Sidecar filtering
- By default, hide MCP sidecar processes. Detect sidecars by matching cmdline against known patterns:
  - `hotels-mcp` (or any path under `~/.gemini/extensions/`)
  - `nanobanana`
  - `dist/index.js --use-dot-names` (google-workspace MCP)
- Alternative heuristic: child process started within ~10s of parent = sidecar.
- Add `--sidecars` flag to show them anyway.

### FR4: Tree rendering
- Use box-drawing characters (`├──`, `└──`) for the tree.
- Each child line shows:
  - `⚙️` emoji prefix
  - PID in **purple/magenta** color (`\033[00;35m`)
  - Command line (truncated to ~60 chars)
  - Elapsed time (from `/proc/<pid>/stat` start time)
- Example output:
  ```
  💻 ⬆️   emagy-88ad590d-...  <1m  ~/TurboSeby  ok i removed the blobs
           ├── ⚙️  [1541570] ollama pull dolphin3:latest    7m
           ├── ⚙️  [1541667] ollama pull qwen3:14b          6m
           └── ⚙️  [1541680] ollama pull gemma4:latest      6m
  ```

### FR5: Purple color constant
- Add `Purple = "\033[00;35m"` to `internal/color/color.go`.

## Non-Functional Requirements
- No new dependencies. Use only `/proc` filesystem and existing Go stdlib.
- Must not slow down the default (non-verbose) output path.
- Works on Linux only (which is fine — emorr-agy is Linux-only).

## Acceptance Criteria
- [ ] `emorr-agy sessions list -v` shows child work processes under active sessions.
- [ ] PIDs are rendered in purple.
- [ ] MCP sidecars are hidden by default.
- [ ] `emorr-agy sessions list -v --sidecars` shows sidecar processes too.
- [ ] Sessions with no work children show no extra lines.
- [ ] Default `sessions list` (no `-v`) output is unchanged.
- [ ] Unit tests cover child process discovery and sidecar filtering.

## Out of Scope
- Subagent-level identification (mapping which subagent spawned which command — requires transcript parsing).
- Progress bars for individual commands.
- Telegram `/status` enhancement (separate track).
