package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	Bot *tgbotapi.BotAPI
}

func NewClient(token string) *Client {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{Bot: bot}
}

func (c *Client) UpdatesChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := c.Bot.GetUpdatesChan(u)
	return updates
}

func (c *Client) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := c.Bot.Send(msg)
	if err != nil {
		log.Println("SendMessage error:", err)
	}
}
