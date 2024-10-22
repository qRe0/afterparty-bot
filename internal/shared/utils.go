package shared

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func ShowOptions(chatID int64, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Выберите опцию поиска покупателя:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Найти по фамилии"),
			tgbotapi.NewKeyboardButton("Найти по части фамилии"),
		),
	)

	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
