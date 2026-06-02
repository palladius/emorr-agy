build:
	go build -o emorragy main.go

telegram-test:
	./emorragy telegram send "Test message from emorragy CLI! 🟢"
