package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	GPTToken      string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		GPTToken:      os.Getenv("GPT_TOKEN"),
	}
	if cfg.TelegramToken == "" || cfg.GPTToken == "" {
		log.Fatal("Missing required environment variables")
	}
	return cfg
}
