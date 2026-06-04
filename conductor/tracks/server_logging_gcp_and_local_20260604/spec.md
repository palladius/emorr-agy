# Specification: Server Logging (GCP & Local)

## Overview
This track implements a dedicated logging system in `emorr-agy` to track bot activity, API calls (Telegram, Gemini), and command execution.

## Functional Requirements
1. **Local File Logging**:
   - Write all server logs to `log/server.log`.
   - The logger must create the `log` directory and the log file if they do not exist.
   - The `log/` directory must be added to `.gitignore` to prevent committing logs.
   - Format: `[YYYY-MM-DD HH:MM:SS] [LEVEL] MESSAGE` (e.g., `[2026-06-04 17:15:00] [INFO] Server started`).

2. **GCP Cloud Logging**:
   - If the environment variable `PROJECT_ID` is set, the logger must send logs to GCP Cloud Logging.
   - To keep dependencies minimal, the integration will use the GCP Logging REST API (`entries:write`) with Application Default Credentials (ADC) or token fetching via `gcloud` command.
   - Log name: `projects/<PROJECT_ID>/logs/emorr-agy-server`.
   - Log payload must include severity (`INFO`, `WARNING`, `ERROR`), log message, timestamp, and optional context fields (e.g., caller details).
   - Sending logs to GCP must be done asynchronously (non-blocking) so that network latency does not slow down the server's response time to Telegram.

## Non-Functional Requirements
- **Performance**: Asynchronous/buffered writing for GCP logging to avoid blocking update polling loop.
- **Resilience**: If GCP Logging API calls fail or auth token is missing, fail silently (write warning to `log/server.log`) without interrupting the bot server execution.
