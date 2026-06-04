# Specification: Fix Audio Transcription Wrong Model

## Overview
This track addresses two issues:
1. The Telegram server's audio transcription and session classification commands currently fail with a 404 error because the model `gemini-2.5-flash` is not supported/found in the `v1beta` Google Generative Language API.
2. The server's startup message in stdout does not print the version of the application.

## Functional Requirements
1. **Configurable / Dynamic Gemini Model**:
   - Check the environment variable `AUDIO_TRANSCRIPTION_GEMINI_MODEL` or `GEMINI_MODEL`. If set, use it.
   - If not set, fallback to a default model defined in a constants file. Currently, the default model is `gemini-3.5-flash` (since older models like `gemini-2.0-flash` are deprecated and `gemini-2.5-flash` is not supported in `v1beta`).
   - Apply this dynamic model selection to both audio transcription in `internal/gemini/transcribe.go` and session classification in `internal/sessions/gemini.go`.
2. **Server stdout version printing**:
   - Update the server startup message in stdout to print the current version (`Version` constant, which is `0.1.1`).

## Acceptance Criteria
- Running `./bin/emorr-agy server` prints: `Server started (v0.1.1) with PID <pid>, listening to Telegram...`
- Session classification (`sessions show --classify <id>`) and voice/audio message transcription commands run successfully without `404 Not Found` API errors.
- Unit tests verify the model selection logic.
