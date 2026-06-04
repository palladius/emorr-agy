# Specification: Sync and Cleanup Rogue Session Changes

## Overview
This specification details the tasks required to verify, document, and officially integrate the changes made during the rogue session `SESSION-1285437-LAST_WISHES.md`. This includes verifying compilation, testing the CLI and background server daemon, and registering documentation assets.

## Functional Requirements
- **Verification of Build & GOROOT:**
  - Verify that the binary compiles successfully using the `just` tool.
  - Verify the binary is named `emorr-agy` (not `emorragy`) and is output to `bin/`.
- **Command & Daemon Verification:**
  - Verify the `status` command output displays version, hostname, daemon state, etc.
  - Verify the `check` command scans for `tmux` mouse/scroll configurations.
  - Verify the `server` daemon initializes, creates `~/.emorr-agy-server.pid`, and parses commands.
  - Verify startup notifications are restricted to the `server` daemon start.
- **Documentation Integration:**
  - Ensure the architecture infographic `doc/img/emorragi_architecture.png` is linked in `README.md` and `conductor/index.md`.
  - Ensure the user manual `doc/CLI.md` is linked in `README.md`.

## Non-Functional Requirements
- **No Residual Duplicates:** Remove any temporary or redundant session artifacts.

## Acceptance Criteria
- Clean compile via `just` yielding `bin/emorr-agy`.
- Successful runtime verification of `status`, `check`, and `server`.
- Links to `doc/CLI.md` and `doc/img/emorragi_architecture.png` are present in `README.md`.
- Conductor tracks registry updated to link the architecture diagram.
