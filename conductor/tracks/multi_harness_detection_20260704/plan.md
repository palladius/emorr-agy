# Implementation Plan: Multi-Harness Session Detection

## Phase 1: Session Source Registry & Multi-Directory Scanning

- [ ] Task: Write Tests — Define `SessionSource` struct and multi-source Classify logic
    - [ ] Add test for `SessionSource` struct creation with harness name and base path
    - [ ] Add test for `Classify()` scanning multiple source directories and assigning correct harness per source
    - [ ] Add test verifying that sessions from `antigravity/` get `Harness: "antigravity-ui"`
    - [ ] Run tests and confirm they fail (Red Phase)

- [ ] Task: Implement — `SessionSource` struct and refactor `Classify()` to use source registry
    - [ ] Define `SessionSource` struct with `Harness string` and `BasePath string` fields in `classify.go`
    - [ ] Add a default source list: `[{"agy", "antigravity-cli"}, {"antigravity-ui", "antigravity"}]`
    - [ ] Refactor the conversation DB scanning loop (step 2 in `Classify()`) to iterate over all sources instead of hardcoding `antigravity-cli`
    - [ ] Refactor `getTranscriptDescription()` to accept a base path parameter or look up from the source registry
    - [ ] Refactor `FindActiveConvs()` to scan lock files across all source directories
    - [ ] Run tests and confirm they pass (Green Phase)

- [ ] Task: Conductor - User Manual Verification 'Phase 1: Session Source Registry & Multi-Directory Scanning' (Protocol in workflow.md)

## Phase 2: Emoji, Display & Filtering

- [ ] Task: Write Tests — Emoji and harness filter for `antigravity-ui`
    - [ ] Add test for `getEmojiForHarness("antigravity-ui")` returning the correct emoji (🖥️)
    - [ ] Add test for `ListSessions` with `Harness: ["antigravity-ui"]` filter returning only Antigravity 2.0 sessions
    - [ ] Run tests and confirm they fail (Red Phase)

- [ ] Task: Implement — Add emoji and ensure filtering works
    - [ ] Add `"antigravity-ui"` case to `getEmojiForHarness()` in `list.go` returning 🖥️
    - [ ] Verify the existing `--harness` filter works for the new harness type (it should, since it's string-based)
    - [ ] Run tests and confirm they pass (Green Phase)

- [ ] Task: Conductor - User Manual Verification 'Phase 2: Emoji, Display & Filtering' (Protocol in workflow.md)

## Phase 3: Cross-Source Deduplication & Edge Cases

- [ ] Task: Write Tests — Deduplication when same convID appears in tmux AND filesystem source
    - [ ] Add test: session in tmux with `emagy-` prefix AND in `antigravity-cli` conversations → only one entry, tmux wins
    - [ ] Add test: session in tmux with `emgem-` prefix AND in `antigravity` conversations → only one entry, tmux wins
    - [ ] Add test: session in `antigravity-cli` AND `antigravity` (unlikely but defensive) → no duplicates
    - [ ] Run tests and confirm they fail (Red Phase)

- [ ] Task: Implement — Extend `runningMap` for multi-source deduplication
    - [ ] Ensure the `runningMap` is populated by tmux sessions first, then checked before adding filesystem sessions from any source
    - [ ] Add cross-source deduplication: if a convID was already added by a previous source, skip it
    - [ ] Run tests and confirm they pass (Green Phase)

- [ ] Task: Conductor - User Manual Verification 'Phase 3: Cross-Source Deduplication & Edge Cases' (Protocol in workflow.md)

## Phase 4: Add `--folder` Filter to `sessions list`

- [ ] Task: Write Tests — Folder filter for `ListSessions`
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

