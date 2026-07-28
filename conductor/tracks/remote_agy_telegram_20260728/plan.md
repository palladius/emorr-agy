# Plan: Remote AGY Sessions via Telegram

**Track ID**: `remote_agy_telegram_20260728`
**Type**: Feature
**Status**: new

---

## Phase 1: PIN Security Hardening & Config Foundation

> Lay the groundwork: secure the PIN, build the YAML config loader, and create the new internal packages.

- [ ] Task: Move PIN to `.env` and remove hardcoded constant
  - [ ] Read `EMORR_AGY_PIN` from `.env` (with `getEnvWithFallback`)
  - [ ] Remove `const HardcodedPin = "4242"` from `main.go`
  - [ ] Add `EMORR_AGY_PIN` to `.env.dist` with documentation
  - [ ] Server refuses to start if PIN is not set (clear error message)
  - [ ] Remove PIN from `/status` output (line ~1039 in `main.go`)
- [ ] Task: Write tests for PIN from env
  - [ ] Test: server reads PIN from env var
  - [ ] Test: server refuses to start without PIN
  - [ ] Test: `/status` output does not contain PIN
  - [ ] Test: existing PIN gate logic works with env-sourced PIN
- [ ] Task: Create `internal/remoteagy/` package
  - [ ] Create package skeleton with doc comment
  - [ ] Define core types: `FolderConfig`, `RemoteSession`, `SessionState`
- [ ] Task: Implement YAML config loader (`internal/remoteagy/config.go`)
  - [ ] Parse YAML file with `folders` array
  - [ ] Support `name`, `path`, `agy_flags` fields
  - [ ] Expand `~` in paths to home directory
  - [ ] Default path: `~/.emorr-agy-folders.yaml`, overridable via `EMORR_AGY_FOLDERS_YAML` env var
  - [ ] Validate: paths exist, names are unique
  - [ ] Graceful degradation: missing/invalid YAML disables remote agy feature (no crash)
- [ ] Task: Write tests for YAML config loader
  - [ ] Test: valid YAML parsing
  - [ ] Test: `~` expansion
  - [ ] Test: missing file → warning, feature disabled
  - [ ] Test: invalid YAML → warning, feature disabled
  - [ ] Test: env var override for config path
  - [ ] Test: duplicate folder names rejected
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## Phase 2: Session Manager (tmux lifecycle)

> Build the core session management: create, list, attach, detach, kill tmux sessions.

- [ ] Task: Implement session manager (`internal/remoteagy/sessions.go`)
  - [ ] `StartSession(folder FolderConfig) (*RemoteSession, error)` — creates tmux session `remote-agy-<name>`, runs `cd <path> && agy <flags>`
  - [ ] `ListSessions() []*RemoteSession` — scans `tmux ls` for `remote-agy-*` sessions
  - [ ] `KillSession(name string) error` — `tmux kill-session -t <name>`
  - [ ] `IsSessionAlive(name string) bool` — check if tmux session exists
  - [ ] `SendInput(name string, text string) error` — `tmux send-keys -t <name> '<text>' Enter`
  - [ ] `CapturePane(name string) (string, error)` — `tmux capture-pane -t <name> -p`
  - [ ] Max 3 concurrent sessions enforced
- [ ] Task: Inject marker string on session start
  - [ ] Generate unique marker: `emorr-agy-marker-<md5-of-timestamp+folder>`
  - [ ] Send marker as first input to the agy session after a brief startup delay
  - [ ] Store marker in `RemoteSession` struct for later transcript lookup
- [ ] Task: Write tests for session manager
  - [ ] Test: StartSession creates tmux session with correct name
  - [ ] Test: ListSessions discovers `remote-agy-*` sessions
  - [ ] Test: KillSession terminates session
  - [ ] Test: SendInput sends keys correctly
  - [ ] Test: CapturePane returns pane content
  - [ ] Test: max 3 sessions enforced (4th start fails)
  - [ ] Test: marker injection on start
- [ ] Task: Implement startup session discovery
  - [ ] On server boot, call `ListSessions()` to find orphaned `remote-agy-*` sessions
  - [ ] Log found sessions and make them available for resume
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## Phase 3: Concierge Menu & Modal Routing

> Build the Telegram UI: concierge menu with folder buttons, session resume buttons, and modal mode state machine.

- [ ] Task: Implement concierge state machine (`internal/remoteagy/concierge.go`)
  - [ ] States: `CONCIERGE` (main menu), `ATTACHED` (to a session)
  - [ ] Track current state and attached session name
  - [ ] `/menu` command always returns to `CONCIERGE` state
- [ ] Task: Build concierge menu Telegram UI
  - [ ] Greeting message with server hostname, version, session count
  - [ ] "🚀 Start New Session" section: inline buttons from YAML folder list
  - [ ] "📎 Resume Session" section: inline buttons for existing `remote-agy-*` sessions
  - [ ] If 3 sessions running: disable "Start" buttons, show warning
  - [ ] Fuzzy text/voice matching: "start gic" → match to "GIC" folder config
- [ ] Task: Build attached-session control button row
  - [ ] 🏠 Menu — detach, return to concierge
  - [ ] 🔄 Refresh — force capture-pane and send
  - [ ] ☠️ Kill — confirmation dialog, then kill session
  - [ ] ⏸️ Pause — toggle output polling on/off
  - [ ] 🔍 Info — show tmux name, PID, uptime, folder, agy_flags
- [ ] Task: Wire concierge into existing processUpdate flow
  - [ ] After PIN unlock, set state to `CONCIERGE`
  - [ ] Route messages/callbacks based on current state
  - [ ] Integrate with existing callback query handling
- [ ] Task: Write tests for concierge state machine
  - [ ] Test: initial state is CONCIERGE after PIN unlock
  - [ ] Test: starting a session transitions to ATTACHED
  - [ ] Test: `/menu` transitions back to CONCIERGE
  - [ ] Test: fuzzy folder matching
  - [ ] Test: max sessions enforcement in UI
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## Phase 4: Output Relay (agy → Telegram)

> Poll tmux output and relay it to Telegram with proper formatting and pagination.

- [ ] Task: Implement output poller (`internal/remoteagy/poller.go`)
  - [ ] Goroutine that runs every 2-3 seconds for the attached session
  - [ ] Calls `CapturePane()` and diffs against last snapshot
  - [ ] Sends only new/changed content to Telegram
  - [ ] Stops polling when session is detached or paused
  - [ ] Detects session exit (tmux session gone) → triggers auto-detach notification
- [ ] Task: Implement message splitting with emoji pagination
  - [ ] Split messages at 4096 char Telegram limit
  - [ ] Use emoji numbers (1️⃣, 2️⃣, ... 9️⃣) as page indicators
  - [ ] Append ✂️ at end of each split (except last) to indicate continuation
  - [ ] Preserve code block boundaries when splitting (don't split mid-block)
- [ ] Task: Write tests for output poller
  - [ ] Test: diff detection (new content only)
  - [ ] Test: no message sent when pane unchanged
  - [ ] Test: session exit detection
  - [ ] Test: pause/resume polling
- [ ] Task: Write tests for message splitting
  - [ ] Test: short message (no split)
  - [ ] Test: long message split with emoji numbers
  - [ ] Test: code block boundary preservation
  - [ ] Test: ✂️ continuation marker
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## Phase 5: Input Relay (Telegram → agy)

> Wire text, voice, and button inputs into the tmux session.

- [ ] Task: Implement input relay in attached mode
  - [ ] Text messages: `SendInput(sessionName, text)`
  - [ ] Voice messages: use existing transcription → `SendInput(sessionName, transcribedText)`
  - [ ] Escape special characters for tmux send-keys safety
- [ ] Task: Write tests for input relay
  - [ ] Test: text message forwarded to tmux
  - [ ] Test: special characters escaped properly
  - [ ] Test: input rejected when not attached (returns to concierge prompt)
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## Phase 6: ask_question Detection & Interactive Buttons

> Parse agy's JSONL transcript to detect interactive choices and render as Telegram buttons.

- [ ] Task: Implement transcript locator (`internal/remoteagy/transcript.go`)
  - [ ] Given marker string, search `~/.gemini/*/brain/*/. system_generated/logs/transcript.jsonl` files
  - [ ] Match the marker in USER_INPUT steps to find the correct conversation ID
  - [ ] Cache the found path for subsequent reads
  - [ ] Use `emorr-agy sessions list` output as a hint for narrowing the search
- [ ] Task: Implement ask_question parser
  - [ ] Tail/poll the transcript JSONL file for new lines
  - [ ] Detect `ask_question` tool calls in PLANNER_RESPONSE steps
  - [ ] Parse question text and options array from the tool call arguments
  - [ ] Handle multi-select vs single-select questions
- [ ] Task: Render ask_question as Telegram inline buttons
  - [ ] Map each option to an inline keyboard button
  - [ ] Include option index (1️⃣, 2️⃣, etc.) in button label
  - [ ] On button tap: send the option text/index back to the agy session via tmux send-keys
  - [ ] Handle "Other" / write-in option (prompt user for text input)
- [ ] Task: Write tests for transcript locator
  - [ ] Test: marker found in transcript
  - [ ] Test: marker not found → graceful degradation
  - [ ] Test: caching works
- [ ] Task: Write tests for ask_question parser
  - [ ] Test: single-select question parsed correctly
  - [ ] Test: multi-select question parsed correctly
  - [ ] Test: non-ask_question tool calls ignored
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## Phase 7: Integration, Polish & YAML Config File

> Wire everything together, create the config file, and polish the UX.

- [ ] Task: Create the YAML config file
  - [ ] Create `$GIC/etc/emorr-agy-folders.yaml` with GIC, GIC Unsafe, Casalborsetti YOLO entries
  - [ ] Symlink `~/.emorr-agy-folders.yaml` → `$GIC/etc/emorr-agy-folders.yaml`
  - [ ] Document in README.md
- [ ] Task: End-to-end integration
  - [ ] Full flow: PIN unlock → concierge → start session → send input → receive output → kill session
  - [ ] Test session resume after simulated server restart
  - [ ] Test max 3 sessions enforcement
- [ ] Task: Update VERSION and CHANGELOG.md
  - [ ] Bump version for this feature
  - [ ] Document all changes in CHANGELOG.md
- [ ] Task: Update README.md
  - [ ] Document remote agy feature
  - [ ] Document YAML config format
  - [ ] Document new env vars (`EMORR_AGY_PIN`, `EMORR_AGY_FOLDERS_YAML`)
- [ ] Task: Phase Verification & Checkpoint (Refer to workflow.md)

---

## V2 TODOs (Out of Scope)

- [ ] LLM-powered concierge (Gemini) for natural language intent recognition
- [ ] Multi-server support (multiple machines, one Telegram bot)
- [ ] Per-user PINs / Telegram user ID whitelisting
- [ ] Auto-suggest folders based on usage frequency
- [ ] File/image relay from agy sessions
- [ ] One-time rotating PIN generation
