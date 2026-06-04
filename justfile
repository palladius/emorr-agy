
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


