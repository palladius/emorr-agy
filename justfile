
list:
	just -l

build:
	mkdir -p bin
	GOROOT=/usr/lib/go go build -o bin/emorr-agy main.go

telegram-test:
	./bin/emorr-agy telegram send "🟢 Test message from emorr-agy CLI from justfile! [blood emoji]"

clean:
	rm -rf bin/ *.out

