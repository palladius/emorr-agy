# Spec: Remote AGY Sessions via Telegram

## Overview

Enable Riccardo to launch, interact with, and manage **Antigravity CLI (`agy`)** sessions
remotely from Telegram, turning his phone into a full agentic coding terminal while on
vacation. The server (running on PupuBooks or any host) acts as a relay between Telegram
and `agy` processes running inside tmux sessions.

### Motivation

When away from a workstation, Riccardo needs the power of `agy` running server-side.
The existing emorr-agy server already handles Telegram connectivity, PIN-based
authentication, voice transcription, and session monitoring — this feature extends it
into an **active control plane** for spawning and interacting with agy sessions.

---

## Functional Requirements

### FR-1: PIN Security Hardening

- **FR-1.1**: Move PIN from hardcoded constant (`HardcodedPin = "4242"`) to `.env` variable `EMORR_AGY_PIN`.
- **FR-1.2**: Remove PIN leak from `/status` output (currently displayed on line 1039 of `main.go`).
- **FR-1.3**: If `EMORR_AGY_PIN` is not set in `.env`, the server MUST refuse to start and print an error.
- **FR-1.4**: Retain existing behavior: 3 wrong attempts → `log.Fatalf` (server dies).

### FR-2: Folder Whitelist Configuration (YAML)

- **FR-2.1**: A YAML configuration file defines the list of folders where agy may be launched.
- **FR-2.2**: Default path: `~/.emorr-agy-folders.yaml`, overridable via env var `EMORR_AGY_FOLDERS_YAML`.
  - In practice, the canonical location will be `$GIC/etc/emorr-agy-folders.yaml`, symlinked to `~/.emorr-agy-folders.yaml`.
- **FR-2.3**: YAML schema:

```yaml
# Whitelisted folders for remote agy sessions via Telegram.
# Each entry defines a Telegram button with a display name, folder path,
# and optional agy launch flags.
# SECURITY: A rogue agy agent can read files in its folder.
#   Do NOT whitelist folders containing secrets (e.g., emorr-agy itself).
folders:
  - name: "GIC"
    path: "~/git/gic"
    agy_flags: ""
  - name: "🔓 GIC Unsafe"
    path: "~/git/gic"
    agy_flags: "--dangerously-skip-permissions"
  - name: "🏖️ Casalborsetti YOLO"
    path: "~/git/ricc-personal-life/casalborsetti-cre-app"
    agy_flags: "--dangerously-skip-permissions"
  - name: "emorr-agy"          # INTENTIONALLY ABSENT — security risk!
    path: "~/git/emorr-agy"    # ❌ Do NOT add this! Agent could read .env
    agy_flags: ""
```

- **FR-2.4**: The `emorr-agy` folder itself MUST NOT be whitelisted (security: agent could read `.env` containing Telegram token and PIN).
- **FR-2.5**: If YAML file is missing or unparseable, server should log a warning and disable the remote agy feature (not crash).
- **FR-2.6**: `agy_flags` are decided at config time (in the YAML). Telegram users CANNOT escalate permissions at runtime.

### FR-3: Concierge Menu (Main Menu)

- **FR-3.1**: After PIN unlock, the user enters the **concierge** (main menu).
- **FR-3.2**: The concierge displays:
  - A greeting with server info (hostname, version).
  - **"Start New Session"** section: One inline button per whitelisted folder from YAML.
  - **"Resume Session"** section: One button per existing `remote-agy-*` tmux session (auto-discovered).
  - If 3 sessions are already running, the "Start New Session" buttons are disabled/hidden with a message: "⚠️ Max 3 concurrent sessions. Kill one first."
- **FR-3.3**: V1 concierge is button-driven with basic fuzzy string matching for voice/text commands (e.g., "start gic" → matches "GIC" folder).
- **FR-3.4**: V2 (out of scope): LLM-powered concierge that understands natural language hints (e.g., "the geek one" → `~/git/gic`).

### FR-4: Session Lifecycle

#### FR-4.1: Starting a Session

- User taps a folder button (e.g., "GIC") from the concierge menu.
- Server creates a tmux session named `remote-agy-<foldername>` (e.g., `remote-agy-gic`).
- Inside the tmux session: `cd <folder_path> && agy <agy_flags>`.
- A unique marker string is injected as the first message to agy: `emorr-agy-marker-<random-md5>`. This marker is used later to locate the agy conversation's JSONL transcript.
- User is auto-attached to the new session (modal mode).
- Confirmation message: "🚀 Started agy session in `GIC`. You are now attached."

#### FR-4.2: Attaching to a Session

- From the concierge, user taps a "Resume" button for an existing session.
- Server switches the user's attached session.
- A catch-up message is sent: the latest `tmux capture-pane` content.
- Confirmation: "📎 Attached to `remote-agy-gic`."

#### FR-4.3: Detaching from a Session

- User taps 🏠 **Back to Menu** button, or sends `/menu`.
- Server stops polling output for this session.
- The tmux session keeps running in the background.
- User returns to the concierge menu.

#### FR-4.4: Killing a Session

- User taps ☠️ **Kill Session** button.
- Confirmation prompt: "⚠️ Are you sure you want to kill `remote-agy-gic`? This will terminate the agy process." with Yes/No buttons.
- On confirm: `tmux kill-session -t remote-agy-gic`.
- User returns to concierge.
- Notification: "💀 Session `remote-agy-gic` terminated."

#### FR-4.5: Session Natural Exit

- When the agy process exits naturally, the polling loop detects it.
- Notification: "✅ agy session `remote-agy-gic` has finished!"
- Auto-detach and return to concierge.
- The tmux session remains alive briefly (e.g., 5 min) so user can re-attach and see final output.

### FR-5: Input Relay (Telegram → agy)

- **FR-5.1**: **Text messages**: When attached to a session, any text message is injected into the tmux session via `tmux send-keys -t remote-agy-<name> '<text>' Enter`.
- **FR-5.2**: **Voice messages**: Already transcribed by existing voice transcription logic. The transcribed text is injected into the tmux session identically to text messages.
- **FR-5.3**: **Button choices** (from ask_question detection): When the user taps a choice button (e.g., 1️⃣), the corresponding text/number is sent to the tmux session.

### FR-6: Output Relay (agy → Telegram)

- **FR-6.1**: **Polling**: Every 2-3 seconds, capture the attached session's tmux pane via `tmux capture-pane -t remote-agy-<name> -p`.
- **FR-6.2**: **Diff**: Compare captured content against the last snapshot. Only send **new content** to Telegram.
- **FR-6.3**: **Markdown formatting**: Send output as Telegram Markdown (code blocks, bold, etc.) preserving agy's formatting as closely as possible.
- **FR-6.4**: **Long message splitting**: Telegram has a 4096-char limit. Split long outputs into multiple messages with emoji page numbers (1️⃣, 2️⃣, 3️⃣, ... 9️⃣) and a ✂️ emoji at the end of each split to indicate continuation.
- **FR-6.5**: **No output for detached sessions**: Only the currently attached session's output is relayed.

### FR-7: ask_question Detection & Telegram Buttons

- **FR-7.1**: When starting an agy session, inject a unique marker (e.g., `emorr-agy-marker-<md5>`) as the first prompt.
- **FR-7.2**: Use the marker to locate the agy conversation's transcript JSONL file in `~/.gemini/*/brain/<conversation-id>/.system_generated/logs/transcript.jsonl`.
  - The `emorr-agy sessions list` command can assist in mapping sessions to conversation IDs.
- **FR-7.3**: Monitor the transcript JSONL for `ask_question` tool calls.
- **FR-7.4**: When an `ask_question` is detected, parse the question text and options, and render them as Telegram inline keyboard buttons.
- **FR-7.5**: When the user taps a button, send the corresponding response back to the agy session via tmux send-keys.

### FR-8: Control Buttons (Persistent Row)

When attached to a session, a persistent row of control buttons is shown:

| Button | Action |
|--------|--------|
| 🏠 Menu | Detach and return to concierge |
| 🔄 Refresh | Force re-capture and re-send current screen |
| ☠️ Kill | Terminate session (with confirmation) |
| ⏸️ Pause | Pause/resume output polling (battery saver) |
| 🔍 Info | Show session metadata: tmux name, PID, uptime, folder |

### FR-9: Startup Session Discovery

- **FR-9.1**: On server startup (after PIN unlock), scan for existing `remote-agy-*` tmux sessions.
- **FR-9.2**: Any found sessions are offered as "Resume" options in the concierge menu.
- **FR-9.3**: This ensures sessions survive server restarts/crashes.

### FR-10: Single Server Per Telegram Channel (V1)

- **FR-10.1**: V1 supports only one emorr-agy server instance per Telegram bot/channel.
- **FR-10.2**: (V2/TODO) Multi-server support: multiple machines sharing one Telegram bot, each identifying itself.

### FR-11: Sessions List Telegram Marker

- **FR-11.1**: `emorr-agy sessions list` must display a 📱 emoji next to sessions whose name matches the `remote-agy-*` pattern.
- **FR-11.2**: This visually distinguishes Telegram-spawned sessions from locally-started ones in CLI output.
- **FR-11.3**: Example output:
  ```
  📱 remote-agy-gic       [BUSY]  ~/git/gic
  📱 remote-agy-casal     [IDLE]  ~/git/ricc-personal-life/casalborsetti-cre-app
     my-local-session      [BUSY]  ~/git/some-project
  ```

### FR-12: Repo Sandboxing (cannot escape git repo)

- **FR-12.1**: The YAML config supports an optional `sandbox: true` (default: `true`) flag per folder.
- **FR-12.2**: When `sandbox: true`, the agy launch command includes `--sandbox` flag (or equivalent) to restrict file access to the git repository root.
- **FR-12.3**: The YAML config also supports a top-level `default_sandbox: true` setting.
- **FR-12.4**: Example YAML:
  ```yaml
  default_sandbox: true   # applied to all folders unless overridden
  folders:
    - name: "GIC"
      path: "~/git/gic"
      agy_flags: ""
      sandbox: true        # redundant but explicit — agy stays in repo
    - name: "🔓 GIC Unsafe"
      path: "~/git/gic"
      agy_flags: "--dangerously-skip-permissions"
      sandbox: false       # explicitly allows escape (USE WITH CAUTION)
  ```
- **FR-12.5**: At session start, emorr-agy logs the sandbox status clearly: `"Starting sandboxed agy session in ~/git/gic"` vs `"⚠️ Starting UNSANDBOXED agy session"`.

---

## Non-Functional Requirements

- **NFR-1**: All new code must have >80% test coverage.
- **NFR-2**: New functionality should be in dedicated packages under `internal/` (not crammed into `main.go`).
- **NFR-3**: No secrets (PIN, tokens) may be hardcoded or logged.
- **NFR-4**: Polling must be efficient — no busy-waiting, proper sleep intervals.
- **NFR-5**: tmux commands must handle errors gracefully (session doesn't exist, tmux not installed, etc.).

---

## Acceptance Criteria

1. ✅ Server reads PIN from `.env` (`EMORR_AGY_PIN`) and refuses to start if missing.
2. ✅ `/status` no longer leaks the PIN.
3. ✅ After PIN unlock, concierge menu appears with folder buttons from YAML.
4. ✅ Tapping a folder button creates a `remote-agy-*` tmux session and launches `agy`.
5. ✅ Text/voice messages while attached are relayed to the agy session.
6. ✅ New agy output appears in Telegram within 2-3 seconds.
7. ✅ Long outputs are split with emoji pagination.
8. ✅ `ask_question` modals in agy are rendered as Telegram inline buttons.
9. ✅ Control buttons (Menu, Refresh, Kill, Pause, Info) work correctly.
10. ✅ Max 3 concurrent sessions enforced.
11. ✅ Sessions survive server restart and appear as "Resume" options.
12. ✅ emorr-agy folder is NOT in the default YAML config.
13. ✅ `emorr-agy sessions list` shows 📱 emoji next to `remote-agy-*` sessions.
14. ✅ YAML `sandbox` flag prevents agy from escaping the git repo when enabled.

---

## Out of Scope (V2+)

- LLM-powered concierge with natural language understanding.
- Multi-server support (multiple machines, one Telegram bot).
- Per-user PINs or Telegram user ID whitelisting.
- Automatic folder suggestion based on usage frequency ("top 10 popular").
- File upload/download between Telegram and agy sessions.
- Screenshot/image relay from agy's generate_image tool.
