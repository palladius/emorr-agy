# Implementation Plan: Session PID Display and Child Process Count

## Phase 1: PID Resolution Infrastructure

- [ ] Task: Write tests for PID-to-session mapping in list output
    - [ ] Test that active sessions include PID in formatted output
    - [ ] Test that dead sessions show no PID bracket
    - [ ] Test that PID is correctly resolved from `FindActiveConvs()` map
- [ ] Task: Implement PID resolution in `ListSessions()`
    - [ ] Pass `activeConvs` map (from `FindActiveConvs()`) into `ListSessions()` via `ListOptions` or engine
    - [ ] For each session, look up PID from the active conversations map
    - [ ] Format as `[PID]` in purple when alive, omit when dead
    - [ ] Apply to both `short` and `long` formats
- [ ] Task: Add Purple color constant to `internal/color/color.go` (if not already added by verbose_process_tree track)
- [ ] Task: Conductor - User Manual Verification 'PID Resolution Infrastructure' (Protocol in workflow.md)

## Phase 2: Child Process Counting

- [ ] Task: Write tests for child process counting and sidecar filtering
    - [ ] Test `GetWorkChildCount()` returns correct count excluding sidecars
    - [ ] Test `GetWorkChildCount()` returns 0 for sessions with only sidecars
    - [ ] Test `IsKnownSidecar()` correctly identifies MCP extension paths
- [ ] Task: Implement `GetWorkChildCount()` in `internal/sessions/ps.go`
    - [ ] Walk `/proc` to find children of a given PID
    - [ ] Read `/proc/<child>/cmdline` for each child
    - [ ] Filter out sidecars using `IsKnownSidecar()`
    - [ ] Return count of non-sidecar children
- [ ] Task: Implement `IsKnownSidecar()` helper
    - [ ] Match against known patterns: `hotels-mcp`, `nanobanana`, paths under `~/.gemini/extensions/`
    - [ ] Match `dist/index.js --use-dot-names` (google-workspace MCP)
- [ ] Task: Conductor - User Manual Verification 'Child Process Counting' (Protocol in workflow.md)

## Phase 3: Compact Count Rendering

- [ ] Task: Write tests for compact count rendering in normal mode
    - [ ] Test that `⚙️ N` appears in purple after session row when N > 0
    - [ ] Test that no indicator appears when N == 0
    - [ ] Test that compact count is suppressed in verbose mode
- [ ] Task: Integrate compact count into `ListSessions()` short and long formats
    - [ ] After rendering each active session row, if not verbose and childCount > 0, append `⚙️ N` colored in purple
    - [ ] Ensure tabwriter alignment is preserved
- [ ] Task: Conductor - User Manual Verification 'Compact Count Rendering' (Protocol in workflow.md)

## Phase 4: JSON Output and Final Polish

- [ ] Task: Write tests for JSON output including PID and child count
    - [ ] Test JSON struct includes `pid` (int) and `work_child_count` (int) fields
    - [ ] Test values are correct for active and dead sessions
- [ ] Task: Add `PID` and `WorkChildCount` fields to `Session` struct
    - [ ] Populate during `Classify()` or during `ListSessions()`
    - [ ] Ensure JSON marshaling uses `pid` and `work_child_count` tags
- [ ] Task: Verify all existing tests still pass
- [ ] Task: Conductor - User Manual Verification 'JSON Output and Final Polish' (Protocol in workflow.md)
