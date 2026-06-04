# Specification: Telegram Audio/Voice Messages Support

## Overview
This specification details the implementation of audio/voice message support in `emorr-agy`. When a user sends a voice or audio message to the Telegram bot, the bot will download the audio file, transcribe and identify the language using the Gemini API, reply to the user with the transcription in *Italic* prefixed by a language flag emoji, and execute the transcription text as a standard command.

## Functional Requirements
1. **Telegram Audio Download:**
   - Detect incoming voice (`Message.Voice`) and audio (`Message.Audio`) files in the Telegram update handler.
   - Request the file download path using the Telegram Bot API `GetFile` method.
   - Download the file (`.ogg`/`.mp3` format) and save it temporarily inside `~/.gemini/antigravity-cli/tmp/`.

2. **Gemini Transcription & Language Identification:**
   - Upload/send the audio file to the Gemini API (using standard multimodal audio integration).
   - Instruct the model to:
     - Transcribe the audio text verbatim.
     - Identify the primary language of the speech (returning an ISO language code like `it`, `en`, `es`, etc.).

3. **Flag Emoji Mapping:**
   - Map the detected language code to a flag emoji:
     - `it` -> `🇮🇹`
     - `en`/`gb` -> `🇬🇧`
     - `us` -> `🇺🇸`
     - `es` -> `🇪🇸`
     - `fr` -> `🇫🇷`
     - `de` -> `🇩🇪`
     - `pt` -> `🇵🇹`
     - `ja` -> `🇯🇵`
     - `zh` -> `🇨🇳`
     - `ru` -> `🇷🇺`
     - Fallback flag: `🌐`

4. **Reply & Execution Loop:**
   - Format: Reply to the voice message with the transcription text in *Italic* prefixed by the flag emoji (e.g., `🇮🇹 _status_`).
   - Route: Pass the transcribed text to the bot's command router to execute the command as if it had been typed directly.

## Acceptance Criteria
- Sending a voice message triggers file retrieval and download.
- Verbatim text is retrieved from the transcription.
- The transcription is sent back in italics with the matching emoji.
- The command is executed successfully.
- Test coverage for audio download, translation, mapping, and routing is >80%.
- Compilation is clean without warnings.
