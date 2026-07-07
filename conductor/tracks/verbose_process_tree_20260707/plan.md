# Implementation Plan: Verbose Process Tree

## Phase 1: Color Support
- [ ] Task: Add Purple color constant
    - [ ] Add `Purple = "\033[00;35m"` to `internal/color/color.go`

## Phase 2: Child Process Discovery
- [ ] Task: Implement `GetChildProcesses()` function in `internal/sessions/ps.go`
    - [ ] Walk `/proc`, read `/proc/<pid>/stat` to find PPID matches
    - [ ] Read `/proc/<pid>/cmdline` for command line
    - [ ] Calculate elapsed time from process start time
    - [ ] Return `[]ChildProcess{PID, Cmdline, ElapsedTime, IsSidecar}`
- [ ] Task: Implement sidecar detection
    - [ ] Match cmdline against known MCP extension paths
    - [ ] Add `IsKnownSidecar(cmdline string) bool` helper

## Phase 3: CLI Flag Wiring
- [ ] Task: Add `--verbose` / `-v` and `--sidecars` flags
    - [ ] Add flags to `sessions list` FlagSet in `main.go`
    - [ ] Add `Verbose` and `ShowSidecars` fields to `ListOptions` struct

## Phase 4: Tree Rendering
- [ ] Task: Implement tree output in `ListSessions()`
    - [ ] After each active session row, if `Verbose`, call `GetChildProcesses(pid)`
    - [ ] Filter out sidecars unless `ShowSidecars` is set
    - [ ] Render `├──` / `└──` tree lines with `⚙️ [PID] cmdline  elapsed`
    - [ ] PID in purple via `color.Colorize(pid, color.Purple)`
    - [ ] Truncate cmdline to ~60 chars

## Phase 5: Testing
- [ ] Task: Unit tests for child process discovery
    - [ ] Test `GetChildProcesses()` with mock `/proc` filesystem
    - [ ] Test `IsKnownSidecar()` with known and unknown cmdlines
- [ ] Task: Unit tests for tree rendering
    - [ ] Test verbose output format includes tree characters
    - [ ] Test non-verbose output is unchanged
    - [ ] Test sidecar filtering
- [ ] Task: Conductor - User Manual Verification 'Testing' (Protocol in workflow.md)
