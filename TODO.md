
* Add install instructions to README. is there some magic thing like `go install github/repo/link`?
- [ ] Why is description empty for some sessions? e.g. 💤 ⬆️ 8be91409-10f6-415d-9d49-03378be7d07f ~/git/banche-pvt (2026-07-05)
- [ ] Gemini CLI sessions use ~/.gemini/tmp/<folder-name>/logs.json with sessionId field, NOT last_conversations.json. Need to scan this directory to properly discover Gemini CLI sessions and show ♊ emoji. Current cwd-matching approach only works when the cache also has the right folder→convID mapping.
