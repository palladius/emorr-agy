# Implementation Plan: Implement Logging

## Phase 1: Local Logging Foundation (v1)
- [ ] Task 1.1: Add `log/` folder to `.gitignore` and ensure directory creation on startup.
- [ ] Task 1.2: Write failing unit tests for the text logger (daily rotation, format verification).
- [ ] Task 1.3: Implement the local text logger with daily rotation (`log/emorr-agy-YYYYMMDD.log`).
- [ ] Task 1.4: Write failing unit tests for Telegram message interaction logging.
- [ ] Task 1.5: Integrate local logging for incoming and outgoing Telegram messages/replies.
- [ ] Task 1.6: Write failing unit tests for command/shell execution logging.
- [ ] Task 1.7: Integrate local logging for all local command/shell executions.
- [ ] Task 1.8: Conductor - User Manual Verification 'Phase 1: Local Logging Foundation (v1)' (Protocol in workflow.md)

## Phase 2: Cloud Logging & OTel Integration (v2)
- [ ] Task 2.1: Add Google Cloud Logging SDK dependency to `go.mod`.
- [ ] Task 2.2: Write failing unit tests for Cloud Logger initialization and authentication fallback behavior.
- [ ] Task 2.3: Implement the Cloud Logger using the GCP SDK (triggered when `PROJECT_ID` is set) with standard OTel metadata configuration.
- [ ] Task 2.4: Write integration test verifying log streaming to GCP (with simulated or mocked client).
- [ ] Task 2.5: Conductor - User Manual Verification 'Phase 2: Cloud Logging & OTel Integration (v2)' (Protocol in workflow.md)

## Phase 3: Validation, Linting, & Code Review
- [ ] Task 3.1: Run all unit and integration tests to verify logging works under both local and cloud modes.
- [ ] Task 3.2: Verify code coverage for logging files is >= 80%.
- [ ] Task 3.3: Verify linting and static analysis are passing cleanly.
- [ ] Task 3.4: Conductor - User Manual Verification 'Phase 3: Validation, Linting, & Code Review' (Protocol in workflow.md)
