# Implementation Plan: Group Sessions by Project/Folder

## Phase 1: Core Grouping Logic & ListOptions
- [x] Task: Update `ListOptions` struct in `internal/sessions/list.go` to include `GroupByFolder bool`.
- [x] Task: Implement folder grouping helper functions in `internal/sessions/list.go` to partition sessions by `Folder`, sort groups by latest activity, and format grouped output for short, long, and JSON formats.

## Phase 2: CLI Flag Integration
- [x] Task: Update `main.go` `sessions list` flag set to parse `--group-by-folder`, `--grouped`, and `-g`.
- [x] Task: Pass `GroupByFolder` option to `sessions.ListSessions()`.

## Phase 3: Unit Testing & Verification
- [x] Task: Write comprehensive unit tests in `internal/sessions/list_test.go` testing `ListSessions` with `GroupByFolder: true` across short, long, and JSON output formats.
- [x] Task: Verify all tests pass with `just test`.
- [x] Task: Update `VERSION` and `CHANGELOG.md`.
