# Specification: Interactive Telegram Buttons, LLM Operator (Cindy), and MCP Sandboxing

## Overview
This specification details the implementation of interactive Telegram buttons, an LLM-driven operator command (`/ask`) using a prompt configuration file, and a secure execution sandboxing layer that routes external system operations through an MCP (Model Context Protocol) client.

## Functional Requirements
1. **Interactive Telegram Buttons**:
   - **Session Actions**: Provide inline buttons for listed sessions to support:
     - `Show`: Display session transcript and status.
     - `Resume`: Re-attach or start the session.
     - `Revive`: Resuscitate a dead session.
     - `Archive`: Exclude or archive the session.
   - **System Actions**: Provide inline buttons for bot control including `/status` and `/restart`.
   - **Harness Compatibility**: Buttons must support unified abstract actions that work across all harnesses (`agy`, `gemini`, `claude`).

2. **LLM Operator ("Cindy")**:
   - **Command `/ask`**: Implement a Telegram command `/ask <query>` to speak to the LLM assistant.
   - **Prompt Template**: Load Cindy's system prompt from `etc/prompts/cindy.txt`.
   - **Action Routing**: Cindy uses the LLM to understand the user's intent and execute actions such as listing sessions, showing a specific session, or checking status.

3. **MCP Sandboxing**:
   - **Execution Restriction**: All shell command executions (e.g. starting tmux, attaching) and filesystem writes must route through a mockable/configurable MCP client (OneMCP).
   - **Security Policy**: The bot server daemon is blocked from executing raw local shell commands directly, ensuring safety.

## Acceptance Criteria
- `/list` and `/listall` display inline keyboards with Show/Resume/Revive/Archive buttons for each session.
- General menu includes buttons for Status and Restart.
- `/ask` loads the configuration from `etc/prompts/cindy.txt`, invokes the Gemini model, and routes commands.
- External process execution runs through the MCP client connector.
- All code changes compiles cleanly and has >80% unit test coverage.
