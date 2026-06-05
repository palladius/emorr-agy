# Antigravity Archival Technical Details

## Overview
In the Antigravity UI and CLI framework, archiving a conversation/session does **not** physically move the SQLite database or the brain transcript files to a separate directory (e.g., no `mv` operations are executed). 

Instead, archiving is a **logical/metadata-driven state change**.

## How It Works

1. **Physical Files Retained**:
   - SQLite Database: `~/.gemini/antigravity-cli/conversations/<session-id>.db`
   - Brain Transcript Folder: `~/.gemini/antigravity-cli/brain/<session-id>/`
   - The above files and directories remain in their original paths on disk.

2. **Logical State Transition**:
   - When a session is archived by the UI, the session's metadata is updated to mark it as closed (e.g., setting `worth_resuscitate: false` in the cache).
   - Alternatively, a session concluding with an exclusion keyword (e.g., `Chumbia` in the `EXCLUDE_RESUSCITATE` configuration) is automatically treated as archived.

3. **Classification**:
   - The `emorr-agy` classification engine reads these flags and assigns the state **`dead_archived`** (rendered with the `⚫` black circle emoji in `monitor` or `sessions list` outputs).
   - Once marked as archived, the session is excluded from active/interactive command lists (like `/list`).
