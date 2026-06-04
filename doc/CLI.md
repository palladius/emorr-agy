# Emorr-Agy CLI User Manual

`emorr-agy` is a Go-based command-line interface (CLI) and orchestrator daemon that monitors active process sessions, manages `tmux` clusters, and integrates real-time alerts into Telegram.

---

## Installation & Compilation

Building the binary requires Go 1.26+. The project includes a `justfile` that configures environment variables (like `GOROOT`) for development machine compatibility.

To compile:
```bash
just build
```

This creates a self-contained binary at `bin/emorr-agy`.

---

## Global CLI Commands

### 1. `status`
Displays a comprehensive diagnostic report of the server environment, server status, active `tmux` sessions, and current Antigravity conversation threads.

* **Usage**:
  ```bash
  ./bin/emorr-agy status
  ```
* **Sample Output**:
  ```text
  📡 Emorr-Agy Status:
  --------------------------------------------------------------------------------
  Version:      v0.1.1
  Hostname:     derek.zrh.corp.google.com
  Telegram:     ✅ Configured
  Server:       ❌ Not Running
  
  Active tmux Sessions:
  --------------------
    No active tmux sessions found (or tmux server not running).
  
  Antigravity Threads:
  -------------------
    🟢 3 Active Threads (monitoring)
    ⚫ 6 Closed Threads (history)
  ```

---

### 2. `monitor`
Performs a deep-dive scan on current Antigravity agent process trees and displays their states (active, idle, tool calling, etc.) dynamically.

* **Usage**:
  ```bash
  ./bin/emorr-agy monitor
  ```
* **Sample Output**:
  ```text
  📡 Antigravity (agy) Thread Monitor:
  --------------------------------------------------------------------------------
  🟢 ff02fea2 - /usr/local/google/home/ricc/git/emorr-agy [Waiting on User 💬]
  🟢 9be55c1a - /usr/local/google/home/ricc/git/emorr-agy [Tool Calling / IO 🛠️]
  🟢 3853afe8 - /usr/local/google/home/ricc/git/pvt-ricc-sre-agent-o11y [Waiting on User 💬]
  ⚫ e43572f5 - /usr/local/google/home/ricc [Closed]
  ⚫ 38d7ad41 - /usr/local/google/home/ricc/bin [Closed]
  ```

---

### 3. `telegram send`
Dispatches a markdown-formatted message to the designated Telegram channels or chat IDs.

* **Usage**:
  ```bash
  ./bin/emorr-agy telegram send "<message_content>"
  ```
* **Example**:
  ```bash
  ./bin/emorr-agy telegram send "🟢 Build success for staging environment!"
  ```
* **Sample Output**:
  ```text
  🎉 Message sent successfully to Telegram!
  ```

---

### 4. `server`
Runs the background daemon that long-polls Telegram for incoming bot messages and dispatches commands (like `/status`, `/monitor`, `/help`) back to the originating chat.

It implements safety locks using a PID file (`~/.emorr-agy-server.pid`) to ensure only one instance of the server runs at a time.

* **Usage**:
  ```bash
  ./bin/emorr-agy server
  ```
* **Sample Output**:
  ```text
  Server started with PID 183204, listening to Telegram...
  ```

---

### 5. `check`
Verifies that `tmux` is installed on the system and inspects the user's `~/.tmux.conf` configuration file to ensure mouse and scroll support (`set -g mouse on`) is correctly configured.

* **Usage**:
  ```bash
  ./bin/emorr-agy check
  ```
* **Sample Output (Success)**:
  ```text
  🔍 Emorr-Agy System Check:
  --------------------------------------------------------------------------------
  ✅ tmux: Installed at /usr/bin/tmux (tmux 3.6)
  ✅ ~/.tmux.conf: Mouse and scrolling support is enabled ('set -g mouse on')
  ```
* **Sample Output (Failure/Missing config)**:
  ```text
  🔍 Emorr-Agy System Check:
  --------------------------------------------------------------------------------
  ✅ tmux: Installed at /usr/bin/tmux (tmux 3.6)
  ❌ ~/.tmux.conf: File does not exist
     👉 Tip: Create ~/.tmux.conf and add 'set -g mouse on' to enable scrolling & mouse clicks.
  ```

---

## Configuration & Environment Variables

The CLI reads configuration from the environment, supporting standard `.env` file loading on startup.

| Environment Variable | Description | Fallbacks / Default |
| :--- | :--- | :--- |
| `TELEGRAM_BOT_ID` | Telegram API Token for your bot | `TELEGRAM_BOT_TOKEN`, `TELEGRAM_APITOKEN` |
| `TELEGRAM_CHAT_ID` | Target Channel or User Chat ID | `TELEGRAM_CHANNEL_ID` (Defaults to Riccardo's direct ID `605724096`) |
