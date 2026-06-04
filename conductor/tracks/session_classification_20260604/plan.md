# Implementation Plan: Harness Session Classification & Management

## Phase 1: Session Classification Engine & Exclusions [checkpoint: 61fae95]
- [x] Task 1.1: Write unit tests verifying session classification logic (mocking tmux sessions and config json dirs) and exclusion rules. (6239ec9)
- [x] Task 1.2: Implement classification logic in `internal/sessions/classify.go` classifying active tmux sessions and config JSON files into the 4 states, including standard harness prefixes (`emagy-`, `emgem-`, `emcld-`). (6e2592f)
- [x] Task 1.3: Support resuscitation exclusions via `EXCLUDE_RESUSCITATE` and defaults (`*Chumbia*`, `*chumbawamba*`). (6e2592f)
- [x] Task 1.4: Conductor - User Manual Verification 'Phase 1: Session Classification Engine & Exclusions' (Protocol in workflow.md) (61fae95)

## Phase 2: Sessions List Subcommand
- [~] Task 2.1: Write unit tests for `sessions list` command outputs, formats (`--short`, `--long`, `--json`), and filtering by `--harness`.
- [ ] Task 2.2: Implement `sessions list` command integration in `internal/sessions/list.go` with tabular output styling, emojis, and JSON format option.
- [ ] Task 2.3: Conductor - User Manual Verification 'Phase 2: Sessions List Subcommand' (Protocol in workflow.md)

## Phase 3: Sessions Show & LLM Classification
- [ ] Task 3.1: Write unit tests verifying `sessions show` command, file locking (PID locking), and LLM classification caching.
- [ ] Task 3.2: Implement `sessions show` command detail display, including resuscitation/resume commands.
- [ ] Task 3.3: Implement the `--classify` / `--llm` flag logic with caching (JSON cache file) and locking (PID lockfile under `~/.emorr-agy/locks/`).
- [ ] Task 3.4: Conductor - User Manual Verification 'Phase 3: Sessions Show & LLM Classification' (Protocol in workflow.md)

## Phase 4: CLI Integration & Validation
- [ ] Task 4.1: Integrate `sessions` subcommand into root `main.go` using a clean CLI structure.
- [ ] Task 4.2: Verify compile parity (`just build`), run all unit tests, ensure >80% code coverage.
- [ ] Task 4.3: Conductor - User Manual Verification 'Phase 4: CLI Integration & Validation' (Protocol in workflow.md)
