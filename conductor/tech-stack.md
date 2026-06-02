# Technology Stack - E. Morricone Ag

## Core Language & Tooling
- **Language**: Go (1.20+)
  - Selected for its concurrency primitives (goroutines/channels) perfect for running concurrent session loops.
  - Native compilability for easy single-binary distribution.

## Libraries & Frameworks
- **Telegram Bot API**: `github.com/go-telegram-bot-api/telegram-bot-api/v5`
  - Fully featured, idiomatic, and highly popular Go client library for the Telegram Bot API.
- **Session Control Interface**: Native command-line `tmux` execution.
  - Interacts with local `tmux` processes via the `os/exec` package.

## Persistence
- **Local State Serialization**: Structured JSON storage.
  - A lightweight local database representation saved at `~/.emorragi_state.json` to serialize and restore session configuration.

## Configuration & Environment
- **Environment Variables**: `github.com/joho/godotenv`
  - Loads `.env` file containing configuration variables (e.g., Telegram tokens).
