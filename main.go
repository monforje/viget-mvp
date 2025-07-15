package main

import (
	"log"
	"viget-mvp/config"
	"viget-mvp/internal/bot"
	"viget-mvp/internal/matcher"
	"viget-mvp/internal/profile"
	"viget-mvp/internal/vibot"
	"viget-mvp/pkg/gpt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg := config.LoadConfig()

	// Telegram Bot API
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal(err)
	}

	// GPT Client
	gptClient := gpt.NewClient(cfg.GPTToken)

	// In-memory storage
	storage := profile.NewInMemoryStorage()

	// Interviewer
	interviewer := vibot.NewInterviewer(gptClient)

	// Matcher
	matcherService := matcher.NewMatcher()

	// Handler (Telegram bot logic)
	handler := bot.NewHandler(botAPI, storage, interviewer, matcherService)

	log.Println("Bot started.")
	handler.Start()
}
