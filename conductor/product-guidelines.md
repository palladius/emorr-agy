# Product Guidelines - E. Morricone Ag

## Tone & Voice
- **Clarity and Precision**: Keep communications concise, helpful, and clean. Wrap technical errors in user-friendly context instead of dumping raw stack traces.
- **Action-Oriented**: Guide the user to the next logical steps or actions using clear verbs and options.

## Telegram Formatting & UI Design
- **Structured Layouts**: Use clean Markdown formatting for status logs.
- **Typography and Highlights**:
  - Monospace for commands (e.g., `/status`, `/projects`) and session identifiers.
  - Bold tags for status labels (e.g., **[BUSY]**, **[IDLE]**).
- **Inline Keyboard Interactions**: Use custom Telegram keyboard buttons (e.g., `Focus`, `Terminate`, `Send Command`) to make interactions fast and mobile-friendly without requiring users to type out long commands.

## UX & Micro-interactions
- **Safety Prompts**: Require confirmation buttons before executing destructive actions such as terminating an active tmux session.
- **Feedback Loop**: Provide immediate visual confirmation (like a temporary Telegram alert or notification checkmark) when a command is executed.
