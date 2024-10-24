package shared

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/models"
)

const (
	OrgLace     = "КРАСНЫЙ"
	VipLace     = "СИНИЙ"
	DefaultLace = "ЖЕЛТЫЙ"
)

func ShowOptions(chatID int64, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Выберите опцию поиска покупателя:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Фамилия"),
			tgbotapi.NewKeyboardButton("Часть фамилии"),
		),
	)

	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ResponseMapper(resp *models.TicketResponse) string {
	successEmoji := "✅"
	failEmoji := "❌"

	var laceColor string
	switch resp.TicketType {
	case "ОРГ":
		laceColor = OrgLace
	case "ВИП":
		laceColor = VipLace
	case "БАЗОВЫЙ":
		laceColor = DefaultLace
	default:
		return "Неизвестный тип билета"
	}

	controlStatus := failEmoji
	if resp.PassedControlZone {
		controlStatus = successEmoji
	}

	return fmt.Sprintf("Номер билета: %s,\nФИО: %s,\nТип браслета: %s,\nЦвет браслета: %s\nПрошел контроль: %s",
		resp.Id, resp.Name, resp.TicketType, laceColor, controlStatus)
}
