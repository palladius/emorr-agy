# Implementation Plan: Fix Audio Transcription Wrong Model

## Phase 1: Model Configuration & Verification
- [x] Task 1.1: Define default Gemini model name config helper in a shared package or locally.
- [x] Task 1.2: Update `internal/gemini/transcribe.go` to construct API URL using dynamic model selection (defaulting to `gemini-3.1-flash-lite`).
- [x] Task 1.3: Update `internal/sessions/gemini.go` to construct API URL using dynamic model selection (defaulting to `gemini-3.1-flash-lite`).
- [x] Task 1.4: Update unit tests in `internal/gemini/transcribe_test.go` and `internal/sessions/gemini_test.go` to verify URL construction.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Model Configuration & Verification' (Protocol in workflow.md)

## Phase 2: Server Startup Message
- [x] Task 2.1: Update `main.go` `runServer()` to output version on stdout startup message.
- [x] Task 2.2: Compile, manually run the server, and verify stdout print contains the version matching `v0.1.1`.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Server Startup Message' (Protocol in workflow.md)
