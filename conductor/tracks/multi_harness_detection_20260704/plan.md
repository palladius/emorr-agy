# Implementation Plan: Multi-Harness Session Detection

## Phase 1: Session Source Registry & Multi-Directory Scanning

- [x] Task: Write Tests — Define `SessionSource` struct and multi-source Classify logic `2d9e8a8`
    - [x] Add test for `SessionSource` struct creation with harness name and base path
    - [x] Add test for `Classify()` scanning multiple source directories and assigning correct harness per source
    - [x] Add test verifying that sessions from `antigravity/` get `Harness: "antigravity-ui"`
    - [x] Run tests and confirm they fail (Red Phase)

- [x] Task: Implement — `SessionSource` struct and refactor `Classify()` to use source registry `2d9e8a8`
    - [x] Define `SessionSource` struct with `Harness string` and `BasePath string` fields in `classify.go`
    - [x] Add a default source list: `[{"agy", "antigravity-cli"}, {"antigravity-ui", "antigravity"}]`
    - [x] Refactor the conversation DB scanning loop (step 2 in `Classify()`) to iterate over all sources instead of hardcoding `antigravity-cli`
    - [x] Refactor `getTranscriptDescription()` to accept a base path parameter or look up from the source registry
    - [x] Refactor `FindActiveConvs()` to scan lock files across all source directories
    - [x] Run tests and confirm they pass (Green Phase)

- [x] Task: Conductor - User Manual Verification 'Phase 1: Session Source Registry & Multi-Directory Scanning' (Protocol in workflow.md) `2d9e8a8`

## Phase 2: Emoji, Display & Filtering

- [x] Task: Write Tests — Emoji and harness filter for `antigravity-ui` `2d9e8a8`
    - [x] Add test for `getEmojiForHarness("antigravity-ui")` returning the correct emoji (🖥️)
    - [x] Add test for `ListSessions` with `Harness: ["antigravity-ui"]` filter returning only Antigravity 2.0 sessions
    - [x] Run tests and confirm they fail (Red Phase)

- [x] Task: Implement — Add emoji and ensure filtering works `2d9e8a8`
    - [x] Add `"antigravity-ui"` case to `getEmojiForHarness()` in `list.go` returning 🖥️
    - [x] Verify the existing `--harness` filter works for the new harness type (it should, since it's string-based)
    - [x] Run tests and confirm they pass (Green Phase)

- [x] Task: Conductor - User Manual Verification 'Phase 2: Emoji, Display & Filtering' (Protocol in workflow.md) `2d9e8a8`

## Phase 3: Cross-Source Deduplication & Edge Cases

- [x] Task: Write Tests — Deduplication when same convID appears in tmux AND filesystem source `fbf6edf`
    - [x] Add test: session in tmux with `emagy-` prefix AND in `antigravity-cli` conversations → only one entry, tmux wins
    - [x] Add test: session in tmux with `emgem-` prefix AND in `antigravity` conversations → only one entry, tmux wins
    - [x] Add test: session in `antigravity-cli` AND `antigravity` (unlikely but defensive) → no duplicates
    - [x] Run tests and confirm they fail (Red Phase)

- [x] Task: Implement — Extend `runningMap` for multi-source deduplication `fbf6edf`
    - [x] Ensure the `runningMap` is populated by tmux sessions first, then checked before adding filesystem sessions from any source
    - [x] Add cross-source deduplication: if a convID was already added by a previous source, skip it
    - [x] Run tests and confirm they pass (Green Phase)

- [x] Task: Conductor - User Manual Verification 'Phase 3: Cross-Source Deduplication & Edge Cases' (Protocol in workflow.md) `fbf6edf`

## Phase 4: Add `--folder` Filter to `sessions list`

- [~] Task: Write Tests — Folder filter for `ListSessions`
    - [ ] Add test for `ListOptions.Folder` filtering sessions by folder prefix match
    - [ ] Add test for `--folder .` resolving to current working directory
    - [ ] Add test for folder filter working across multiple harness types
    - [ ] Run tests and confirm they fail (Red Phase)

- [ ] Task: Implement — Port `isPathMatch()` and add `--folder` flag
    - [ ] Move or copy `isPathMatch()` from `main.go` to `internal/sessions/list.go` (or a shared util)
    - [ ] Add `Folder string` field to `ListOptions` struct
    - [ ] Apply folder prefix filtering in `ListSessions()` after classify
    - [ ] Add `--folder` and `-f` flags to the `sessions list` FlagSet in `main.go`
    - [ ] Handle `~` expansion and relative path resolution (`filepath.Abs`)
    - [ ] Run tests and confirm they pass (Green Phase)

- [ ] Task: Conductor - User Manual Verification 'Phase 4: Add `--folder` Filter to `sessions list`' (Protocol in workflow.md)

