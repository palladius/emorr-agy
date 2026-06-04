# Implementation Plan: Telegram Audio/Voice Messages Support

## Phase 1: Telegram Audio Download & Helper [checkpoint: a9de854]
- [x] Task 1.1: Write unit tests verifying audio/voice message detection, file path retrieval, and local downloading/saving. (a9de854)
- [x] Task 1.2: Implement audio download logic in `internal/telegram/audio.go` using Bot API `GetFile` and downloading the audio payload to `~/.gemini/antigravity-cli/tmp/`. (a9de854)
- [x] Task 1.3: Conductor - User Manual Verification 'Phase 1: Telegram Audio Download & Helper' (Protocol in workflow.md) (a9de854)

## Phase 2: Gemini Transcription & Language Detection [checkpoint: b1f1a4c]
- [x] Task 2.1: Write unit tests for the Gemini transcription service, verifying mock audio payloads, transcription text retrieval, and language code identification. (b1f1a4c)
- [x] Task 2.2: Implement `internal/gemini/transcribe.go` using Gemini API to transcribe files and return the text and language code. (b1f1a4c)
- [x] Task 2.3: Implement the 10-flag mapping helper (`it` -> `🇮🇹`, etc.). (b1f1a4c)
- [x] Task 2.4: Conductor - User Manual Verification 'Phase 2: Gemini Transcription & Language Detection' (Protocol in workflow.md) (b1f1a4c)

## Phase 3: Bot Integration & Command Execution [checkpoint: ffe2740]
- [x] Task 3.1: Write unit tests verifying that a voice message triggers a transcription reply in italics with emoji and runs the command. (ffe2740)
- [x] Task 3.2: Update the Telegram bot update handler in `main.go` to orchestrate the download, transcription, echo reply, and command execution. (ffe2740)
- [x] Task 3.3: Verify compile parity (`just build`), run all unit tests, ensure >80% code coverage. (ffe2740)
- [x] Task 3.4: Conductor - User Manual Verification 'Phase 3: Bot Integration & Command Execution' (Protocol in workflow.md) (ffe2740)
