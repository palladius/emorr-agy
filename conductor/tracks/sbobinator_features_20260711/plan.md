# Implementation Plan: Port agy-sbobinator Features

## Phase 1: 3-State Session Classification (P0 — GitHub Issue #6)

- [x] Task: Add SessionClassification type and constants `f2b2124`
    - [ ] Write tests for classification constants and string representation in `internal/sessions/classification_test.go`
    - [ ] Create `internal/sessions/classification.go` with `SessionClassification` type (`Finished`, `NeedsResume`, `Obsolete`)
    - [ ] Add emoji and label helpers: `ClassificationEmoji()`, `ClassificationLabel()`

- [ ] Task: Implement transcript-based structural interruption detection
    - [ ] Write tests for `IsStructurallyInterrupted()` covering: last step is PLANNER_RESPONSE with tool_calls → true, last step is PLANNER_RESPONSE without tool_calls → false, last step is USER_INPUT → false, empty transcript → false
    - [ ] Create `internal/sessions/transcript_analyzer.go` with `IsStructurallyInterrupted(transcriptPath string) bool`
    - [ ] Parse last step of transcript.jsonl, check if type is PLANNER_RESPONSE and tool_calls is non-empty

- [ ] Task: Implement explicit quit detection
    - [ ] Write tests for `IsExplicitQuit()` covering: `/quit`, `/exit`, `exit`, `quit`, `/q`, normal prompts, prompts with XML metadata wrapping
    - [ ] Add `IsExplicitQuit(transcriptPath string) bool` to transcript_analyzer.go
    - [ ] Extract last USER_INPUT content, clean XML metadata tags, check against quit commands

- [ ] Task: Implement repo name extraction from transcripts
    - [ ] Write tests for `ExtractRepoName()` covering: tool calls with file paths, multiple repos (returns most common), no paths → empty, system paths filtered out
    - [ ] Add `ExtractRepoName(transcriptPath string) string` to transcript_analyzer.go
    - [ ] Regex scan tool call args and content for `/home/.../git/<repo>/...` pattern

- [ ] Task: Implement supersession detection
    - [ ] Write tests for `DetectSupersession()` covering: newer session on same repo → true, only session for repo → false, different repos → false
    - [ ] Add `DetectSupersession(sessions []SessionMetadata) map[string]bool` — maps session ID to superseded status
    - [ ] Track latest session per repo, mark older ones as superseded

- [ ] Task: Implement 7-day recency window classification
    - [ ] Write tests for `ClassifySession()` combining all factors: recent + not superseded + not quit → NeedsResume, recent + superseded → Obsolete, old + structurally interrupted → Obsolete, old + clean → Finished, explicit quit → Finished
    - [ ] Add `ClassifySession(meta SessionMetadata, isSuperseded bool, now time.Time) SessionClassification`
    - [ ] Wire all detection functions together

- [ ] Task: Integrate classification into ListSessions output
    - [ ] Write tests verifying ListSessions shows 3-state emojis (🟢/🔴/⚠️) instead of just Open/Closed
    - [ ] Modify `ListSessions()` in list.go to call `ClassifySession()` for dead sessions
    - [ ] Update short/long/JSON formats with classification field

- [ ] Task: Conductor - User Manual Verification 'Phase 1: 3-State Classification' (Protocol in workflow.md)

## Phase 2: Transcript Timeline Viewer (P1 — GitHub Issue #7)

- [ ] Task: Add timeline subcommand CLI wiring
    - [ ] Write test for CLI argument parsing of `timeline <id>` subcommand
    - [ ] Add `timeline` case to main.go switch, wire to `sessions.ShowTimeline()`

- [ ] Task: Implement transcript timeline renderer
    - [ ] Write tests for `ShowTimeline()` covering: USER_INPUT rendering with XML cleanup, PLANNER_RESPONSE text rendering, run_command tool calls showing CommandLine, view_file tool calls showing AbsolutePath, other tool calls showing name+args, date separator on day change
    - [ ] Create `internal/sessions/timeline.go` with `ShowTimeline(w io.Writer, transcriptPath string) error`
    - [ ] Parse transcript.jsonl line by line, render formatted output with emojis and ANSI colors

- [ ] Task: Add session stats at bottom of timeline
    - [ ] Write tests for stats display (duration, step count, start/end times)
    - [ ] Add stats section after timeline rendering in ShowTimeline()

- [ ] Task: Add resume command hints at bottom
    - [ ] Write tests for resume command hint formatting
    - [ ] Show `cd ~/git/<repo> && agy --conversation <uuid>` and tmux variant

- [ ] Task: Conductor - User Manual Verification 'Phase 2: Timeline Viewer' (Protocol in workflow.md)

## Phase 3: Auto-Resume and CLI Archive (P2 — GitHub Issue #8)

- [ ] Task: Implement auto-resume logic
    - [ ] Write tests for `FindAutoResumeTarget()` covering: picks latest NeedsResume session, excludes current session, falls back to most recent if no NeedsResume, returns nil if no sessions
    - [ ] Add `FindAutoResumeTarget(sessions []Session, currentUUID string) *Session` to resume.go
    - [ ] Add `--auto` / `--latest` flag to `resume` command in main.go

- [ ] Task: Implement CLI archive command
    - [ ] Write tests for `ArchiveSessionCLI()` covering: archives specific session by ID, archives all FINISHED+OBSOLETE without ID, creates archive dir, skips NeedsResume sessions
    - [ ] Add `ArchiveSessionCLI(engine *ClassificationEngine, sessionID string) error` to resume.go
    - [ ] Add `sessions archive [id]` subcommand to main.go

- [ ] Task: Add long mode with resume commands
    - [ ] Write tests for `--long` output format showing reasons and resume commands
    - [ ] Add `Long` field to `ListOptions`
    - [ ] Update ListSessions to show explanation lines when Long is true: why status was assigned, cd + agy --conversation command, tmux command variant

- [ ] Task: Conductor - User Manual Verification 'Phase 3: Auto-Resume and CLI Archive' (Protocol in workflow.md)

## Phase 4: Integration and Polish

- [ ] Task: Update Telegram /list to use 3-state classification
    - [ ] Write integration test verifying /list filters by NeedsResume (not just "Waiting on User")
    - [ ] Modify processUpdate /list handler to use classification-based filtering

- [ ] Task: Update help text and version
    - [ ] Update printUsage() with new subcommands (timeline, sessions archive)
    - [ ] Bump VERSION to 0.1.8
    - [ ] Update CHANGELOG.md with new features

- [ ] Task: Ensure test coverage > 80% for new code
    - [ ] Run `go test -coverprofile=coverage.out ./...`
    - [ ] Verify all new files in internal/sessions/ have > 80% coverage

- [ ] Task: Commit and comment on GitHub Issues
    - [ ] Stage all changes and commit with `feat(sessions): Port 3-state classification, timeline, auto-resume from sbobinator`
    - [ ] Comment on GitHub Issues #6, #7, #8 with Conductor track ID and implementation status

- [ ] Task: Conductor - User Manual Verification 'Phase 4: Integration and Polish' (Protocol in workflow.md)
