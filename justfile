
list:
	just -l

build:
	mkdir -p bin
	GOROOT=/usr/lib/go go build -o bin/emorr-agy main.go

telegram-test:
	./bin/emorr-agy telegram send "🟢 Test message from emorr-agy CLI from justfile! [blood emoji]"

clean:
	rm -rf bin/ *.out

test:
	GOROOT=/usr/lib/go go test -v ./...

run-server-under-tmux: build
	tmux new-session -d -s emorr-agy-server "./bin/emorr-agy server" || echo "Session 'emorr-agy-server' already exists. Use 'just attach-server' to view."

attach-server:
	tmux attach -t emorr-agy-server

version:
	@cat VERSION

show-logs limit="10":
	#!/usr/bin/env bash
	set -euo pipefail
	PROJ="${PROJECT_ID:-${GCP_PROJECT:-${GOOGLE_CLOUD_PROJECT:-}}}"
	if [ -z "$PROJ" ] && [ -f .env ]; then
		PROJ=$(grep -E '^(PROJECT_ID|CLOUD_PROJECT_ID|GCP_PROJECT|GOOGLE_CLOUD_PROJECT)=' .env | head -n1 | cut -d= -f2 | tr -d "'\"" | xargs)
	fi
	if [ -z "$PROJ" ]; then
		PROJ=$(gcloud config get-value project 2>/dev/null)
	fi
	if [ -z "$PROJ" ]; then
		echo "Error: Could not resolve GCP project ID. Set PROJECT_ID or configure .env" >&2
		exit 1
	fi
	echo "Reading latest {{limit}} logs from Cloud Logging for project '${PROJ}'..."
	gcloud logging read "logName=projects/${PROJ}/logs/emorr-agy-server" --limit={{limit}} --project="${PROJ}" --format="table(timestamp, severity, jsonPayload.message)"
