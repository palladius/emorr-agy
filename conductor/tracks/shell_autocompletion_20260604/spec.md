# Specification: Shell Autocompletion

## Overview
This track introduces shell autocompletion for `emorr-agy` to enhance user experience when using the CLI. It will support both Bash and Zsh shells. The autocompletion script will be dynamically generated via a new subcommand: `emorr-agy completion <shell>` (supporting `bash` and `zsh`).
It will support static completion of commands and flags, as well as dynamic completion of session IDs for `sessions show` by querying the running tmux sessions/database files on the system.

## Functional Requirements
1. **Completion Subcommand**:
   - `emorr-agy completion bash`: Outputs the Bash autocompletion script to stdout.
   - `emorr-agy completion zsh`: Outputs the Zsh autocompletion script to stdout.
2. **Static Completions**:
   - Complete first-level subcommands: `telegram`, `monitor`, `status`, `server`, `check`, `sessions`, `completion`.
   - Complete second-level subcommands under `sessions`: `list`, `show`.
   - Complete flags for commands:
     - For `sessions list`: `--harness`, `--json`, `--long`, `--short`.
     - For `sessions show`: `--classify`, `--llm`.
3. **Dynamic Completions**:
   - For `emorr-agy sessions show`, typing `[TAB][TAB]` will dynamically list active session IDs parsed from the system.
4. **Shell Compatibility**:
   - Ensure the scripts are compatible with modern Bash (version 4+) and Zsh shell environments.

## Non-Functional Requirements
- **Performance**: Autocompletion resolution must run in <15ms to prevent lag during tab completion.

## Acceptance Criteria
- Running `source <(emorr-agy completion bash)` in Bash enables tab-completion.
- Running `source <(emorr-agy completion zsh)` in Zsh enables tab-completion.
- Typing `emorr-agy [TAB][TAB]` lists all first-level subcommands.
- Typing `emorr-agy sessions [TAB][TAB]` lists `list` and `show`.
- Typing `emorr-agy sessions show [TAB][TAB]` lists available session IDs parsed dynamically from the system.

## Out of Scope
- Support for Fish, PowerShell, or command-line shells other than Bash and Zsh.
- Dynamic completion of other flags/argument values (like harness names) unless easily queried.
