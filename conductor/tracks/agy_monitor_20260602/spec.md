# Track Spec: `agy-monitor` Status Tool

## Overview
A high-performance, blazing-fast Go utility (`agy-monitor.go`) designed to monitor all local Antigravity (`agy`) threads and report their active execution states with intuitive emojis. The utility will be highly optimized for `watch` mode.

## Functional Requirements
1. **Thread Discovery**:
   - Reads `~/.gemini/antigravity-cli/cache/last_conversations.json` to map project directories to active conversation IDs.
   - Discovers all conversation databases (`.db` files) under `~/.gemini/antigravity-cli/conversations/`.
2. **Process Association & Open State (🟢 / ⚫)**:
   - Traverses the Linux `/proc` filesystem to check PIDs for open file descriptors pointing to any conversation `.db` or `.pb` files.
   - If a running `agy` process holds the file open, it is marked as **Open/Active** (`🟢`). Otherwise, it is **Closed/Inactive** (`⚫`).
3. **Execution State Inference**:
   - **Tool Calling / IO (🛠️)**: Detected if the parent `agy` process has running child processes (like `bash`, `git`, etc.) in its process tree.
   - **Gemini Writing (✍️)**: Detected if the process is actively running and generating output without child processes.
   - **Waiting on User (💬)**: Detected if the process is idle, sleeping, or waiting for input on standard input.
4. **Output Format**:
   - Displays a clean, compact terminal output list: `<Emoji> <Conversation ID (Short)> - <Project Directory> [<Detailed State>]`.
   - Fully optimized for near-instant execution (sub-10ms) suitable for standard `watch -n 1` loops.
