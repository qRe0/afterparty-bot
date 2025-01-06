package handlers

import (
	"context"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/service"
	"github.com/qRe0/afterparty-bot/internal/shared"
)

type TicketsServiceInterface interface {
	SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI)
	SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI)
	MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI)
}

type MessagesHandler struct {
	service    *ticket_service.TicketsService
	userStates map[int64]string
}

func New(service *ticket_service.TicketsService) MessagesHandler {
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
			_, _ = bot.Send(msg)
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
			msg := tgbotapi.NewMessage(chatID, "Введите фамилию или часть фамилии для поиска в списках:")
			_, _ = bot.Send(msg)
			mh.userStates[chatID] = "awaiting_surname"

		case "Номер билета (ID)":
			msg := tgbotapi.NewMessage(chatID, "Введите номер билета (ID):")
			_, _ = bot.Send(msg)
			mh.userStates[chatID] = "awaiting_ticket_id"

		default:
			if update.Message.Text != "" {
				if mh.userStates[chatID] == "awaiting_ticket_id" {
					mh.service.SearchById(ctx, &update.Message.Text, &chatID, bot)
				} else if mh.userStates[chatID] == "awaiting_surname" {
					mh.service.SearchBySurname(ctx, &update.Message.Text, &chatID, bot)
				}
			}
		}
	}
}
