# Specification: Implement Logging

## Overview
This specification details the design and implementation of a robust logging system for `emorr-agy`. The logger will trace all interactions with external systems (such as Telegram API requests/responses, MCP actions, and local command execution). It supports two targets depending on the environment:
1. **Local Development (v1):** Writes to daily rotating log files in the `log/` folder in simple text format.
2. **Cloud/Production Deployment (v2):** Exports structured logs to Google Cloud Logging (using Google Cloud Logging SDK) and supports OpenTelemetry (OTel) compatibility (via standard OTel env vars) if `PROJECT_ID` is configured.

## Functional Requirements
- **Local Logging (v1):**
  - Appends logs to `log/emorr-agy-YYYYMMDD.log` (git-ignored directory).
  - Format: `[TIMESTAMP] [LEVEL] [COMPONENT] Message`.
  - Captures:
    - Telegram incoming messages/events and outgoing message payloads.
    - MCP tool execution calls, inputs, and outcomes.
    - Local command/shell executions.
- **Cloud/OTel Logging (v2):**
  - Triggered if `PROJECT_ID` is defined.
  - Initializes Google Cloud Logging SDK and streams structured JSON logs.
  - Integrates OTel environment variables (e.g., standard metadata, trace propagation where applicable).
- **Graceful Fallback:**
  - Fallbacks gracefully to local/stdout logging if GCP cloud logging initialization fails.

## Non-Functional Requirements
- **Low Overhead:** Extremely low latency overhead.
- **Configurability:** Controlled via environment variables (`PROJECT_ID`, `LOG_LEVEL`, `LOG_DIR`).

## Acceptance Criteria
- A daily log file is generated in the `log/` directory during local runs.
- Log entries correctly document incoming/outgoing Telegram messages and commands executed.
- When `PROJECT_ID` is set and valid credentials exist, logs stream to Google Cloud Logging.
- Unit and integration tests for logging coverage exceed 80%.

## Out of Scope
- In-bot interactive log viewer.
- Full distributed tracing integration for external systems outside `emorr-agy`'s direct calls.
