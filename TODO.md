
* Add install instructions to README. is there some magic thing like `go install github/repo/link`?
- [x] Why is description empty for some sessions? → Fixed: scan full transcript for first USER_INPUT (commit 4f5ab4f)
- [x] Gemini CLI deep discovery via SHA256 path hashing → 81% folder resolution (commit ca3ba10)
- [ ] 9 Gemini CLI sessions still unresolved (paths no longer exist on disk)
- [ ] Performance: reading full transcript files for descriptions (could use bufio.Scanner for early exit)
- [ ] ♊ emoji renders 1 col narrower than 🖥️/⬆️ (terminal font issue, no Go fix possible)
