package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/service"
	"github.com/qRe0/afterparty-bot/internal/shared"
)

type MessagesHandler struct {
	service    *service.TicketsService
	userStates map[int64]string
}

func NewMessagesHandler(service *service.TicketsService) MessagesHandler {
	return MessagesHandler{
		service:    service,
		userStates: make(map[int64]string),
	}
}

func (mh *MessagesHandler) HandleMessages(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	ctx := context.Background()
	chatID := update.Message.Chat.ID

	switch update.Message.Text {
	case "/start":
		shared.ShowOptions(chatID, bot)

	case "Фамилия":
		mh.userStates[chatID] = "full"
		msg := tgbotapi.NewMessage(chatID, "Введите фамилию для поиска:")
		bot.Send(msg)

	case "Часть фамилии":
		mh.userStates[chatID] = "partial"
		msg := tgbotapi.NewMessage(chatID, "Введите часть фамилии для поиска:")
		bot.Send(msg)

	case "ID":
		mh.userStates[chatID] = "id"
		msg := tgbotapi.NewMessage(chatID, "Введите ID для поиска:")
		bot.Send(msg)

	case "Проход":
		mh.userStates[chatID] = "entrance"
		msg := tgbotapi.NewMessage(chatID, "Введите ID для отметки о входе:")
		bot.Send(msg)

	default:
		if update.Message.Text != "" {
			messageType := mh.userStates[chatID]
			if messageType == "full" {
				mh.service.SearchByFullSurname(ctx, &update.Message.Text, &chatID, bot)
			} else if messageType == "partial" {
				mh.service.SearchBySurnamePart(ctx, &update.Message.Text, &chatID, bot)
			} else if messageType == "id" {
				mh.service.SearchByID(ctx, &update.Message.Text, &chatID, bot)
			} else if messageType == "entrance" {
				mh.service.MarkAsEntered(ctx, &update.Message.Text, &chatID, bot)
			}
		}
		shared.ShowOptions(chatID, bot)
	}
}
