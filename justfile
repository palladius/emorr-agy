
list:
	just -l

build:
	mkdir -p bin
	GOROOT=/usr/lib/go go build -o bin/emorragy main.go

telegram-test:
	./bin/emorragy telegram send "🟢 Test message from emorragy CLI from justfile! [blood emoji]"
