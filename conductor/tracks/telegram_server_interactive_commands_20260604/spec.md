# Specification: Telegram Server Interactive Commands

## Overview
This track implements interactive Telegram command handlers `/list` and `/listall` in the `emorr-agy` bot server. 
- `/list`: Fetches all active sessions on the system, filters for those pending human interaction (e.g., `Waiting on User`), and displays them as inline keyboard buttons.
- `/listall`: Fetches all sessions on the system (both active and historical/dead) and displays the last 5 as inline keyboard buttons.
- Selecting a session button prints the session details along with any captured options/choices as inline buttons (e.g., option buttons parsed from the last 10 lines of the tmux pane). Senders can tap these option buttons to send the key input directly to the active tmux session.

## Functional Requirements
1. **Interactive Commands**:
   - `/list`: Lists up to 5 active sessions with state `Waiting on User`. Displays each session as an Inline Keyboard Button with label format: `<session_id> (<directory_basename>)`.
   - `/listall`: Lists up to 5 sessions of any kind (active, dead, private), sorted by last activity, as Inline Keyboard Buttons.
2. **Session Detail & Action Buttons**:
   - Tapping a session button triggers a callback query. The bot answers by sending the session details and the last 10 lines from the pane (captured using `tmux capture-pane` or logs).
   - If the session is `Waiting on User` and running in `tmux`, the bot parses the last 10 lines to extract the available options (e.g. lines starting with numbers like `1.`, `2.`, `3.`, `4.` under the prompt).
   - The bot renders these choices as inline keyboard buttons (e.g. `Option 1`, `Option 2`, `Option 3`, `Option 4` or with the actual text if brief).
3. **Response Sourcing**:
   - When a user taps a choice button in Telegram, the bot sends the corresponding keystroke (e.g. `"1\n"`, `"2\n"`) directly to the target tmux session pane using `tmux send-keys -t <session_id> "key"`.
   - The bot edits/refreshes the message to show the updated pane content.

## Non-Functional Requirements
- **Robustness**: The bot must fail gracefully if the tmux session has exited or is detached.
- **Latency**: Inline response time must be under 1 second.

## Acceptance Criteria
- Sending `/list` to the Telegram bot returns inline buttons for active human-pending sessions.
- Sending `/listall` returns inline buttons for the last 5 sessions.
- Tapping a session button shows the last 10 lines and buttons for each option.
- Tapping an option button sends the keystroke to the tmux pane and updates the Telegram message.
