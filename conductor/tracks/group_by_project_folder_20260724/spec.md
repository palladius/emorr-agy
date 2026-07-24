# Specification: Group Sessions by Project/Folder

## Overview
Currently, `sessions list` displays all sessions in a flat, date-sorted list. For users managing multiple project directories, grouping active and resuscitatable sessions by their working directory (folder) with an emoji header makes it significantly easier to navigate sessions related to specific projects.

Refer to [GitHub Issue #9](https://github.com/palladius/emorr-agy/issues/9).

## Functional Requirements
1. **ListOptions & CLI Flag Support**:
   - Add `GroupByFolder bool` field to `sessions.ListOptions`.
   - Support CLI flags `--group-by-folder`, `--grouped`, and shorthand `-g` for `emorr-agy sessions list`.
2. **Folder Grouping & Output Formatting**:
   - Group sessions by `Session.Folder` (cleaned path, tildified for display, e.g. `~/git/emorr-agy`).
   - Sessions with missing or empty folder paths should be grouped under `📁 (Unknown folder)`.
   - Each group header starts with a folder emoji (e.g. `📂` or `📁`) followed by the tildified folder path in bold/colored style (e.g. `📂 ~/path/to/adk-sre-benjamin/`).
   - Print sessions under their respective group headers with appropriate indentation (e.g., 2 spaces or tab-aligned).
   - In grouped text view ("short" or "long" formats), the `DIR` column can be omitted from individual rows since the directory is already specified in the header, or aligned cleanly.
3. **Ordering**:
   - Sort groups by the `LastActivity` of their most recent session (descending).
   - Sort sessions within each group by `LastActivity` descending.
4. **Format Compatibility**:
   - Short and Long formats support grouped output with header lines and indented session rows.
   - JSON format when `GroupByFolder: true` outputs a structured map/dictionary or array of folder groups (e.g., `[{"folder": "...", "sessions": [...]}]`).

## Non-Functional Requirements
- Maintain backward compatibility when `--group-by-folder` / `-g` is not specified (default remains flat date-sorted list).
- Code modularity and full test coverage for grouping logic.

## Acceptance Criteria
- Running `emorr-agy sessions list --group-by-folder` (or `-g` / `--grouped`) displays output grouped by folder with folder emojis.
- Unit tests in `internal/sessions/list_test.go` cover `ListSessions` with `GroupByFolder: true` for short, long, and JSON formats.
- All unit tests pass via `just test`.
