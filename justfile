build:
	mkdir -p bin
	go build -o bin/emorragy main.go

telegram-test:
	./bin/emorragy telegram send "🟢 Test message from emorragy CLI from justfile! [blood emoji]"
