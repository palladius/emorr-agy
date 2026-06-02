
This is an agy orchestrator for Telegram written in Go.
See BDD in  /home/riccardo/obsidian/PBTPersonalSync/Progetti/E. Morricone Ag.md
Use conductor to mplement this! Make sure this is persisted in github in palladius/emorr-agy

Open points:
* Does Antigravity SDK support "open vs not open threads" and "active vs waiting for user input"?
* If so lets start with a super quick prototype.

Use .env for Telegram stuff, Gemini dflt models, api keys and stuff.

DO NOT EXFILTRATE API KEYS, OR TELEGRAM TOKENS, FOR ANY REASONS!
ENSURE THIS WITH A PRESUBMIT CHECK! (I've been burnt this morning).

**Note** Agy SDK just python: https://antigravity.google/docs/sdk-overview 