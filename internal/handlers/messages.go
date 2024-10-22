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

	case "Найти по фамилии":
		mh.userStates[chatID] = "full"
		msg := tgbotapi.NewMessage(chatID, "Введите фамилию для поиска:")
		bot.Send(msg)

	case "Найти по части фамилии":
		mh.userStates[chatID] = "partial"
		msg := tgbotapi.NewMessage(chatID, "Введите часть фамилии для поиска:")
		bot.Send(msg)

	default:
		if update.Message.Text != "" {
			searchType := mh.userStates[chatID]
			if searchType == "full" {
				mh.service.SearchByFullSurname(ctx, &update.Message.Text, &chatID, bot)
			} else if searchType == "partial" {
				mh.service.SearchBySurnamePart(ctx, &update.Message.Text, &chatID, bot)
			}
		}

		shared.ShowOptions(chatID, bot)
	}
}
