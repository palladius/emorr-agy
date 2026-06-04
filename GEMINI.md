## DOs and DONTs

* DO NOT WRITE in .env! Only read it! To suggest a change create a .env.proposed and ask USER to copy/merge.
* Use .env for Telegram stuff, Gemini dflt models, api keys and stuff.
* DO NOT EXFILTRATE API KEYS, OR TELEGRAM TOKENS, FOR ANY REASONS! ENSURE THIS WITH A PRESUBMIT CHECK! (I've been burnt this morning).


## About

This is an agy orchestrator for Telegram written in Go.
Use `conductor` skillset to implement this package! Every feature should be a Concudtor NewTrack

## Code

* Code should be modular (more small files with folders which make sense!)
* Should be unit tested.
* CLI should be battle tested (--help should work well, and --options should try to work as possible with most verbs)
* add a `VERSION` file (or some go idiomatic way to track verison) and a `CHANGELOG.md` to keep track of changes by version.

## GitHub Issues

Periodically (and at startup) check open GH Issues.

* If no conductor track name/id is mentioned in it, use `/conductor:newTrack` to onboard it.
* Add a comment on the Issue (via `gh`) with the conductor id, so we don't duplicate work!
* Label it "Conductor": so it appears here: https://github.com/palladius/emorr-agy/issues?q=state%3Aopen%20label%3A%22Conductor%22
* Keep updated the Issue with progress, and close when done.
* Sign your GHI messages with `-- Written by <HARNESS_NAME> on behalf of <PERSON_NAME>`

## Bibliography

**Note** Agy SDK just python: https://antigravity.google/docs/sdk-overview
