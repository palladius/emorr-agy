package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()

	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	subCommand := os.Args[2]

	if command != "telegram" || subCommand != "send" {
		printUsage()
		os.Exit(1)
	}

	// Join all remaining arguments as the message text
	message := strings.Join(os.Args[3:], " ")
	if message == "" {
		log.Fatal("Error: Message text cannot be empty")
	}

	err := sendTelegramMessage(message)
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	fmt.Println("🎉 Message sent successfully to Telegram!")
}

func printUsage() {
	fmt.Println("Usage: emorragy telegram send <message>")
	fmt.Println("Example: emorragy telegram send Ciao come va")
}

func sendTelegramMessage(text string) error {
	botToken := getEnvWithFallback("TELEGRAM_BOT_TOKEN", "TELEGRAM_APITOKEN")
	chatID := getEnvWithFallback("TELEGRAM_CHAT_ID", "TELEGRAM_CHANNEL_ID")

	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN or TELEGRAM_APITOKEN is not set in .env")
	}
	if chatID == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID or TELEGRAM_CHANNEL_ID is not set in .env")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {chatID},
		"text":       {text},
		"parse_mode": {"Markdown"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK status: %s", resp.Status)
	}

	return nil
}

func getEnvWithFallback(key, fallbackKey string) string {
	val := os.Getenv(key)
	if val == "" {
		val = os.Getenv(fallbackKey)
	}
	return val
}
