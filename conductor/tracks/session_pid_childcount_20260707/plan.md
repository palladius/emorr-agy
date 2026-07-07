# Implementation Plan: Session PID Display and Child Process Count

## Phase 1: Color and Sidecar Infrastructure

- [ ] Task: Add Purple color constant to `internal/color/color.go`
- [ ] Task: Write tests for sidecar detection
    - [ ] Test `IsKnownSidecar()` correctly identifies MCP extension paths (nanobanana, hotels-mcp, google-workspace)
    - [ ] Test `IsKnownSidecar()` returns false for work commands (ollama, bash, git)
- [ ] Task: Implement `IsKnownSidecar()` helper in `internal/sessions/ps.go`
    - [ ] Match against known patterns: `hotels-mcp`, `nanobanana`, paths under `~/.gemini/extensions/`
    - [ ] Match `dist/index.js --use-dot-names` (google-workspace MCP)
- [ ] Task: Write tests for `GetWorkChildCount()`
    - [ ] Test returns correct count excluding sidecars
    - [ ] Test returns 0 for sessions with only sidecars
    - [ ] Test returns 0 for dead/non-existent PIDs
- [ ] Task: Implement `GetWorkChildCount()` in `internal/sessions/ps.go`
    - [ ] Walk `/proc` to find children of a given PID
    - [ ] Read `/proc/<child>/cmdline` for each child
    - [ ] Filter out sidecars using `IsKnownSidecar()`
    - [ ] Return count of non-sidecar children
- [ ] Task: Conductor - User Manual Verification 'Color and Sidecar Infrastructure' (Protocol in workflow.md)

## Phase 2: Compact Count in Normal Mode

- [ ] Task: Write tests for compact count rendering
    - [ ] Test that `⚙️N` appears in purple after session row when N > 0
    - [ ] Test that no indicator appears when N == 0
    - [ ] Test that no PIDs are shown in normal mode
- [ ] Task: Integrate compact count into `ListSessions()` short and long formats
    - [ ] After rendering each active session row, if childCount > 0, append `⚙️N` colored in purple
    - [ ] Ensure tabwriter alignment is preserved
- [ ] Task: Conductor - User Manual Verification 'Compact Count in Normal Mode' (Protocol in workflow.md)

## Phase 3: Verbose Mode PID Display

- [ ] Task: Add `--verbose` / `-v` flag to `sessions list` CLI (if not already added by verbose_process_tree track)
    - [ ] Add `Verbose` field to `ListOptions`
- [ ] Task: Write tests for verbose PID display
    - [ ] Test that session row shows `[PID]` in purple before session ID when alive
    - [ ] Test that dead sessions show no PID bracket
    - [ ] Test that child process tree lines show `[PID]` in purple
    - [ ] Test that compact count is NOT shown in verbose mode
- [ ] Task: Implement verbose PID rendering in `ListSessions()`
    - [ ] When verbose and session is alive, prepend `[PID]` in purple before session ID
    - [ ] When verbose, render child tree lines with `[PID]` in purple (leveraging verbose_process_tree track)
    - [ ] Suppress compact count `⚙️N` when in verbose mode
- [ ] Task: Conductor - User Manual Verification 'Verbose Mode PID Display' (Protocol in workflow.md)

## Phase 4: JSON Output and Final Polish

- [ ] Task: Write tests for JSON output including PID and child count
    - [ ] Test JSON struct includes `pid` (int) and `work_child_count` (int) fields
    - [ ] Test values are correct for active and dead sessions
- [ ] Task: Add `PID` and `WorkChildCount` fields to `Session` struct
    - [ ] Populate during `Classify()` or during `ListSessions()`
    - [ ] Ensure JSON marshaling uses `pid` and `work_child_count` tags
- [ ] Task: Verify all existing tests still pass
- [ ] Task: Conductor - User Manual Verification 'JSON Output and Final Polish' (Protocol in workflow.md)
