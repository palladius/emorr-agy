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
- [logging_20260604](tracks/logging_20260604/index.md) (planned)
- [rogue_session_cleanup_20260604](tracks/rogue_session_cleanup_20260604/index.md) (completed)
- [harness_discovery_20260604](tracks/harness_discovery_20260604/index.md) (planned)
- [codebase_modularization_20260604](tracks/codebase_modularization_20260604/index.md) (planned)
- [session_classification_20260604](tracks/session_classification_20260604/index.md) (planned)
- [telegram_audio_messages_20260604](tracks/telegram_audio_messages_20260604/index.md) (planned)
- [audio_transcription_model_fix_20260604](tracks/audio_transcription_model_fix_20260604/index.md) (planned)

## References
- [Product Guide](product.md)
- [Product Guidelines](product-guidelines.md)
- [Tech Stack](tech-stack.md)
- [Workflow](workflow.md)

---

- [ ] **Track: Implement Logging**
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

- [ ] **Track: lets have a GOOD think at this: https://github.com/palladius/emorr-agy/issues/4**
*Link: [./tracks/session_classification_20260604/](./tracks/session_classification_20260604/)*

---

- [ ] **Track: Telegram should support audio messgaes**
*Link: [./tracks/telegram_audio_messages_20260604/](./tracks/telegram_audio_messages_20260604/)*

---

- [ ] **Track: Fix audio transcription model and print server version in stdout**
*Link: [./tracks/audio_transcription_model_fix_20260604/](./tracks/audio_transcription_model_fix_20260604/)*
