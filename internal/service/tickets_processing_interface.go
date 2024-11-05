package ticket_service

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TicketsServiceInterface interface {
	SearchByFullSurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI)
	SearchBySurnamePart(ctx context.Context, surnamePart *string, chatID *int64, bot *tgbotapi.BotAPI)
	MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI)
}
