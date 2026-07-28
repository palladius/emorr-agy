# Track: Remote AGY Sessions via Telegram

**Track ID**: `remote_agy_telegram_20260728`
**Type**: Feature
**Status**: new

## Documents

- [Specification](./spec.md) — The *What* and the *How*
- [Implementation Plan](./plan.md) — The execution roadmap (7 phases, TDD)
- [Metadata](./metadata.json)

## Summary

Enable Riccardo to launch, interact with, and manage Antigravity CLI (`agy`) sessions
remotely from Telegram. Turns his phone into a full agentic coding terminal while on
vacation. Features include:

- PIN from `.env` (security hardening)
- YAML-based folder whitelist with per-folder agy flags
- Concierge menu with folder buttons and session resume
- tmux-based session management (max 3 concurrent)
- Output polling (2-3s) with emoji pagination
- ask_question detection from JSONL transcripts → Telegram inline buttons
- Modal mode: attached to one session at a time
- Control buttons: Menu, Refresh, Kill, Pause, Info
