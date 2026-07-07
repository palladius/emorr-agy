# Specification: Session PID Display and Child Process Count

## Overview
Enhance `emorr-agy sessions list` to show the PID of alive sessions and a compact child process count in both normal and verbose modes. This complements the existing `verbose_process_tree_20260707` track (which adds the `--verbose` child tree) by making process liveness visible at a glance in **all** output modes.

## Functional Requirements

### FR1: Session PID in row
- When a session's agy process is alive (PID resolvable from `FindActiveConvs()`), prepend the PID in purple/magenta before the session ID.
- Format: `🟢 ♊️  [12345] ba1934d4-daef-490c-...`
- When the process is NOT alive (dead/sleeping sessions), show no PID bracket — the row stays unchanged.
- Applies to **both** normal (`short`) and `long` formats.
- PID is colored in purple (`\033[00;35m`), consistent with verbose mode.

### FR2: Compact child process count (normal mode)
- After each active session row that has non-sidecar child processes, append an inline indicator: `⚙️ N` where N is the count, colored in purple.
- Example: `💻 ⬆️  [745915] emagy-88ad590d-...  <1m  ~/TurboSeby  ok i removed the blobs ⚙️3`
- Sessions with zero work children show no indicator.
- This appears in **normal mode only** — in verbose mode, the full tree is shown instead.

### FR3: Sidecar exclusion for count
- Reuse the sidecar detection logic from the `verbose_process_tree` track (`IsKnownSidecar()`).
- Only non-sidecar children are counted.

### FR4: Integration with verbose mode
- In `--verbose` mode, the session PID is still shown (FR1), but the compact count (FR2) is replaced by the full child tree from the `verbose_process_tree` track.

## Non-Functional Requirements
- No new dependencies. Uses existing `/proc` scanning from `FindActiveConvs()` and child process walking.
- Adds `Purple` color constant if not already present (may already be added by verbose_process_tree track).
- Must not break existing JSON output format (`--json`). PID and child count should be added to the JSON struct.

## Acceptance Criteria
- [ ] Active sessions show `[PID]` in purple before the session ID.
- [ ] Dead sessions show no PID bracket.
- [ ] Normal mode shows `⚙️ N` in purple for sessions with N > 0 work children.
- [ ] Verbose mode shows PID but replaces count with full tree.
- [ ] JSON output includes `pid` and `work_child_count` fields.
- [ ] Existing tests pass; new tests cover PID display and count rendering.

## Out of Scope
- Full child tree rendering (covered by `verbose_process_tree_20260707`).
- Subagent identification from transcripts.
- Telegram `/status` enhancement.
