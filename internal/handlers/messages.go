package handlers

import (
	"context"
	"log"
	"strings"

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
	var chatID int64

	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		data := update.CallbackQuery.Data

		if strings.HasPrefix(data, "confirm_yes_") {
			userId := strings.TrimPrefix(data, "confirm_yes_")
			mh.service.MarkAsEntered(ctx, &userId, &chatID, bot)
		} else if strings.HasPrefix(data, "confirm_no_") {
			msg := tgbotapi.NewMessage(chatID, "Операция отменена.")
			bot.Send(msg)
		} else {
			userId := data
			mh.service.MarkAsEntered(ctx, &userId, &chatID, bot)
		}

		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.Request(callback); err != nil {
			log.Printf("Ошибка при отправке Callback: %v", err)
		}
		return
	}

	if update.Message != nil {
		chatID = update.Message.Chat.ID

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

		default:
			if update.Message.Text != "" {
				messageType := mh.userStates[chatID]
				if messageType == "full" {
					mh.service.SearchByFullSurname(ctx, &update.Message.Text, &chatID, bot)
				} else if messageType == "partial" {
					mh.service.SearchBySurnamePart(ctx, &update.Message.Text, &chatID, bot)
				}
			}
		}
	}
}
