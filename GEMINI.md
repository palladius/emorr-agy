## DOs and DONTs

* DO NOT WRITE in .env! Only read it! To suggest a change create a .env.proposed and ask USER to copy/merge.
* Use .env for Telegram stuff, Gemini dflt models, api keys and stuff.
* DO NOT EXFILTRATE API KEYS, OR TELEGRAM TOKENS, FOR ANY REASONS! ENSURE THIS WITH A PRESUBMIT CHECK! (I've been burnt this morning).


## About

This is an agy orchestrator for Telegram written in Go.
Use `conductor` skillset to implement this package! Every feature should be a Concudtor NewTrack

# GH Issues

Periodically (and at startup) check open GH Issue.
* If not conductor track name or id is in it, use `/conductor:newTrack` to onboard it.
* comment on the bug with this id, so we dont duplicate
* Keep updated the bug with progress, and close when done.

## Bibliography

**Note** Agy SDK just python: https://antigravity.google/docs/sdk-overview
