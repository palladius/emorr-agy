# Initial Concept

Build E. Morricone Ag, the Go-based Telegram orchestrator for remote tmux sessions specified in our BDD.

---

# Product Guide - E. Morricone Ag

## Vision & Goals
**E. Morricone Ag** (`emorr-agy`) is a Go-based Telegram bot and CLI orchestrator that empowers developers and system administrators to manage, monitor, and interact with multiple remote `tmux` sessions. Acting as an active control plane, it brings full visibility and real-time command orchestration directly to Telegram chat.

The key goals are:
- **Convenient Control**: Command and interact with active processes directly from Telegram.
- **State Transparency**: Provide a clear overview of active vs idle sessions.
- **Resilience**: Ensure session persistence and automatic restore capability upon host reboot.
- **Multi-Project Organization**: Segment and filter active sessions by specific projects.

## Core Features
1. **Multi-Project Workspace Context**:
   - Filter active terminal workspaces dynamically.
   - Command `/projects` retrieves active projects and switches context seamlessly.
2. **Standardized Session Names (tmux session naming)**:
   - Automated naming conventions for newly spawned jobs (e.g. `emorragi-data-analysis-1654172400`) to guarantee cleanliness and query stability.
3. **Active/Idle Process State Tracking**:
   - Continuous inspection of session activities to determine whether a terminal process is actively running (**[BUSY]**) or suspended/waiting for user input (**[IDLE]**).
4. **Boot Persistence & Automated Recovery**:
   - Periodic serialization of current session setups to a persistent local store (e.g., `~/.emorragi_state.json`).
   - Automated systemd or startup daemon checks to restore prior jobs upon host reboot.
5. **Interactive Navigation & Focus Control**:
   - Focus on an idle session to issue commands, inspect console buffers, or terminate processes via button actions inside Telegram.

## Target Audience
- System Administrators looking for a lightweight, secure chat-based terminal controller.
- Software Engineers running long-running jobs (e.g., scraping, machine learning training) who need easy status reports and interactive control on the go.
