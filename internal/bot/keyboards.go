package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func MainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Пройти интервью"),
			tgbotapi.NewKeyboardButton("Мой профиль"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Доступные задачи"),
			tgbotapi.NewKeyboardButton("Создать задачу"),
		),
	)
}

func BackKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назад"),
		),
	)
}

func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
}
