# Emorr-Agy Interactive Features & Testing Guide

This guide describes how to test and run the three newly implemented interactive features on the `emorr-agy` bot server:
1. **Startup PIN Gate Security** (unlocked via `4242`).
2. **Microphone-friendly Harness Spawner** (`/new` or `new`).
3. **Local CLI Script Wrappers** (`emagy-run`).

---

## 1. Testing the Startup PIN Gate

The server daemon now starts in a locked state by default. All commands and callback button clicks are blocked until the correct PIN is provided.

### Step-by-Step Test:
1. **Compile the server**:
   ```bash
   just build
   ```
2. **Start the server under tmux**:
   ```bash
   just run-server-under-tmux
   ```
3. **Verify Startup Notification**:
   Open Telegram. You should receive a startup notification:
   > 🟢 **Emorr-Agy v0.1.4 started on <hostname>**
   > 🔒 **Server is locked.** Please enter the PIN (`4242`) to unlock.

4. **Test Command Blocking**:
   Send `/status` or `/list` to the bot.
   * **Expected Result**: The bot replies with a lock warning:
     > 🔒 **Server is locked.** Please enter the correct PIN to unlock. (Attempt 1/3)

5. **Test Fatal Autokill (3 wrong PINs)**:
   Send three incorrect PINs (e.g., `1111`, `1111`, `1111`).
   * **Expected Result**: On the 3rd wrong attempt, the bot sends:
     > ❌ **Too many wrong PIN attempts. Shutting down server for security!**
   * The server process immediately terminates itself. You can verify this by checking `tmux ls` (the `emorr-agy-server` session will have closed).

6. **Test Successful Unlock**:
   Restart the server:
   ```bash
   just run-server-under-tmux
   ```
   Type the correct PIN:
   ```text
   4242
   ```
   * **Expected Result**: The bot replies:
     > 🔓 **Server unlocked successfully! Welcome back.**
   * All commands (`/status`, `/list`, etc.) are now fully operational.

---

## 2. Spawning a Harness (`/new` or `new`)

You can now request a new harness session either by typing or by using your microphone/voice transcription.

### Testing via Telegram Chat:
1. Type and send:
   ```text
   /new gemini write a go test
   ```
   * *Or use the microphone/voice transcription to say*:
     > new gemini write a go test
2. **Expected Result**: The bot runs the tmux session in the background and replies with:
   > 🚀 **Spawned new tmux session `emgem-YYMMDD-HHMMSS` running harness `gemini` with query: "write a go test"**
   * Below the message, an inline button will appear showing the session name and directory (e.g., `emgem-260607-135736 (home)`).
3. Click the inline button.
   * **Expected Result**: The message updates to show the live stdout/options of that newly spawned session.

---

## 3. Testing Local CLI Script Wrappers

If you prefer to start a session directly from your local terminal and have the Telegram bot automatically register and track it, use the wrapper scripts.

### Step-by-Step Test:
1. **Start a session**:
   ```bash
   # Run the short script alias:
   ./bin/emagy-run gemini write a quick bash script
   ```
2. **Expected Result**:
   * The terminal launches a new tmux session named `emgem-YYMMDD-HHMMSS`.
   * It immediately attaches you to the running `gemini` harness process.
   * On Telegram, run `/list` or `/listall`. You will see the new session listed under **Active Sessions** with options to interact or archive it.
3. Detach from the session by pressing `Ctrl+B` then `D` (or let the harness finish and exit).

---

## 4. Troubleshooting
* **Check bot logs**:
  To view the real-time logs of the server, use:
  ```bash
  just show-logs
  ```
