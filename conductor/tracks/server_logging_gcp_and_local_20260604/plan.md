# Implementation Plan: Server Logging (GCP & Local)

## Phase 1: Local File Logger & .gitignore Setup
- [ ] Task 1.1: Implement `internal/logger/logger.go` supporting structured local logging into `log/server.log`.
- [ ] Task 1.2: Add `log/` folder to `.gitignore`.
- [ ] Task 1.3: Add unit tests verifying local log directory/file creation and log formatting.

## Phase 2: GCP Cloud Logging Integration
- [ ] Task 2.1: Implement token provider in `internal/logger/auth.go` (checks metadata server and `gcloud auth print-access-token`).
- [ ] Task 2.2: Implement non-blocking, asynchronous GCP logging payload transmitter in `internal/logger/gcp.go`.
- [ ] Task 2.3: Add unit tests using mock servers for auth token fetching and log payload transmission.

## Phase 3: Server Integration & Verification
- [ ] Task 3.1: Integrate logger in `main.go`, logging server startup, incoming Telegram messages, commands routing, and errors.
- [ ] Task 3.2: Verify compile parity (`just build`) and run all tests (`just test`).
