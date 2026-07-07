# Specification: Session PID Display and Child Process Count

## Overview
Enhance `emorr-agy sessions list` to show the PID of alive sessions and a compact child process count in both normal and verbose modes. This complements the existing `verbose_process_tree_20260707` track (which adds the `--verbose` child tree) by making process liveness visible at a glance in **all** output modes.

## Functional Requirements

### FR1: Compact child process count (normal mode)
- After each active session row that has non-sidecar child processes, append an inline indicator: `⚙️N` where N is the count, colored in purple.
- Example: `💻 ⬆️   emagy-88ad590d-...  <1m  ~/TurboSeby  ok i removed the blobs ⚙️3`
- Sessions with zero work children show no indicator.
- No PIDs are shown in normal mode — just the count.

### FR2: PIDs for every live process (verbose mode)
- In `--verbose` mode, every live process shows its PID in a greppable format: `[PID]`.
- **Session row**: prepend `[PID]` in purple before the session ID.
  - Example: `💻 ⬆️   [745915] emagy-88ad590d-...  <1m  ~/TurboSeby  ok i removed the blobs`
- **Child process rows**: each child tree line shows `[PID]` in purple.
  - Example: `├── ⚙️  [1541570] ollama pull dolphin3:latest  7m`
- Dead sessions show no PID bracket.
- The PID format `[DIGITS]` is consistent and greppable (e.g. `grep '\[' | grep PID`).

### FR3: Sidecar exclusion
- Reuse the sidecar detection logic (`IsKnownSidecar()`).
- Only non-sidecar children are counted (normal mode) or shown (verbose mode, unless `--sidecars` flag).

### FR4: Color
- All PIDs and counts use purple (`\033[00;35m`) for consistency.
- Add `Purple` color constant to `internal/color/color.go` if not already present.

## Non-Functional Requirements
- No new dependencies. Uses existing `/proc` scanning from `FindActiveConvs()` and child process walking.
- Adds `Purple` color constant if not already present (may already be added by verbose_process_tree track).
- Must not break existing JSON output format (`--json`). PID and child count should be added to the JSON struct.

## Acceptance Criteria
- [ ] Normal mode shows `⚙️N` in purple for sessions with N > 0 work children.
- [ ] Normal mode does NOT show any PIDs.
- [ ] Verbose mode shows `[PID]` in purple on the session row for alive sessions.
- [ ] Verbose mode shows `[PID]` in purple on each child process tree line.
- [ ] Dead sessions show no PID bracket in verbose mode.
- [ ] JSON output includes `pid` and `work_child_count` fields.
- [ ] Existing tests pass; new tests cover count rendering and PID display.

## Out of Scope
- Full child tree rendering (covered by `verbose_process_tree_20260707`).
- Subagent identification from transcripts.
- Telegram `/status` enhancement.
