package shared

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/models"
)

const (
	OrgLace     = "Красный"
	VipLace     = "Синий"
	DefaultLace = "Желтый"
)

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

func ResponseMapper(resp *models.TicketResponse) string {
	switch resp.TicketType {
	case "ОРГ":
		return fmt.Sprintf("Номер билета: %s,\nФИО: %s,\nТип браслета: %s,\nЦвет браслета: %s", resp.Id, resp.Name, resp.TicketType, OrgLace)
	case "ВИП":
		return fmt.Sprintf("Номер билета: %s,\nФИО: %s,\nТип браслета: %s,\nЦвет браслета: %s", resp.Id, resp.Name, resp.TicketType, VipLace)
	case "БАЗОВЫЙ":
		return fmt.Sprintf("Номер билета: %s,\nФИО: %s,\nТип браслета: %s,\nЦвет браслета: %s", resp.Id, resp.Name, resp.TicketType, DefaultLace)
	}

	return "Неизвестный тип билета"
}
