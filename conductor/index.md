# Conductor Project Index

## Product Vision
E. Morricone Ag (emorr-agy) is a Go-based Telegram bot and CLI orchestrator to manage, monitor, and interact with remote tmux sessions.

Refer to the [Architecture Infographic](file:///usr/local/google/home/ricc/git/emorr-agy/doc/img/emorragi_architecture.png) for a visual overview of the end-to-end command flow.

## Project Goals
- Convenient process command and control via Telegram.
- Clean process state transparency and activity tracking.
- Session persistence and automated crash/reboot recovery.
- Multi-project workspace segregation.

## Technology Stack
- **Primary Language**: Go (1.20+)
- **Key Libraries**: `go-telegram-bot-api`, `tmux` CLI, JSON serialization

## Active Tracks
- [000_initial_setup](tracks/000_initial_setup/index.md) (planned)
- [agy_monitor_20260602](tracks/agy_monitor_20260602/index.md) (completed)
- [telegram_channel_assets_20260602](tracks/telegram_channel_assets_20260602/index.md) (planned)
- [logging_20260604](tracks/logging_20260604/index.md) (completed)
- [rogue_session_cleanup_20260604](tracks/rogue_session_cleanup_20260604/index.md) (completed)
- [harness_discovery_20260604](tracks/harness_discovery_20260604/index.md) (planned)
- [codebase_modularization_20260604](tracks/codebase_modularization_20260604/index.md) (planned)
- [session_classification_20260604](tracks/session_classification_20260604/index.md) (completed)
- [telegram_audio_messages_20260604](tracks/telegram_audio_messages_20260604/index.md) (completed)
- [audio_transcription_model_fix_20260604](tracks/audio_transcription_model_fix_20260604/index.md) (completed)
- [shell_autocompletion_20260604](tracks/shell_autocompletion_20260604/index.md) (planned)
- [telegram_server_interactive_commands_20260604](tracks/telegram_server_interactive_commands_20260604/index.md) (planned)
- [interactive_llm_operator_20260605](tracks/interactive_llm_operator_20260605/index.md) (planned)
- [multi_harness_detection_20260704](tracks/multi_harness_detection_20260704/index.md) (new)
- [verbose_process_tree_20260707](tracks/verbose_process_tree_20260707/index.md) (new)
- [session_pid_childcount_20260707](tracks/session_pid_childcount_20260707/index.md) (new)
- [sbobinator_features_20260711](tracks/sbobinator_features_20260711/index.md) (new)

## References
- [Product Guide](product.md)
- [Product Guidelines](product-guidelines.md)
- [Tech Stack](tech-stack.md)
- [Workflow](workflow.md)

---

- [x] **Track: Implement Logging**
*Link: [./tracks/logging_20260604/](./tracks/logging_20260604/)*

---

- [x] **Track: Sync and Cleanup Rogue Session Changes**
*Link: [./tracks/rogue_session_cleanup_20260604/](./tracks/rogue_session_cleanup_20260604/)*

---

- [ ] **Track: Add harness discovery to emorr-agy check**
*Link: [./tracks/harness_discovery_20260604/](./tracks/harness_discovery_20260604/)*

---

- [ ] **Track: Restructure codebase to use multiple modular files in subdirectories**
*Link: [./tracks/codebase_modularization_20260604/](./tracks/codebase_modularization_20260604/)*

---

- [x] **Track: lets have a GOOD think at this: https://github.com/palladius/emorr-agy/issues/4**
*Link: [./tracks/session_classification_20260604/](./tracks/session_classification_20260604/)*

---

- [x] **Track: Telegram should support audio messgaes**
*Link: [./tracks/telegram_audio_messages_20260604/](./tracks/telegram_audio_messages_20260604/)*

---

- [x] **Track: Fix audio transcription model and print server version in stdout**
*Link: [./tracks/audio_transcription_model_fix_20260604/](./tracks/audio_transcription_model_fix_20260604/)*

---

- [ ] **Track: Shell autocompletion for emorr-agy subcommands and dynamic session IDs in Bash/Zsh.**
*Link: [./tracks/shell_autocompletion_20260604/](./tracks/shell_autocompletion_20260604/)*

---

- [x] **Track: Telegram Server interactive commands /list and /listall for session response handling.**
*Link: [./tracks/telegram_server_interactive_commands_20260604/](./tracks/telegram_server_interactive_commands_20260604/)*

---

- [ ] **Track: Add interactive Telegram buttons, LLM operator (Cindy) via /ask, and MCP command sandboxing to the server daemon.**
*Link: [./tracks/interactive_llm_operator_20260605/](./tracks/interactive_llm_operator_20260605/)*

---

- [~] **Track: Detect multiple session types (agy CLI, Antigravity 2.0 UI, Claude Code)**
*Link: [./tracks/multi_harness_detection_20260704/](./tracks/multi_harness_detection_20260704/)*

---

- [ ] **Track: Add --verbose flag to sessions list showing child process tree with PIDs in purple**
*Link: [./tracks/verbose_process_tree_20260707/](./tracks/verbose_process_tree_20260707/)*

---

- [ ] **Track: Show session PID in purple when alive and compact child process count in normal mode**
*Link: [./tracks/session_pid_childcount_20260707/](./tracks/session_pid_childcount_20260707/)*

---

- [~] **Track: Port NEEDS_RESUME classification, timeline viewer, auto-resume, and supersession detection from agy-sbobinator**
*Link: [./tracks/sbobinator_features_20260711/](./tracks/sbobinator_features_20260711/)*


