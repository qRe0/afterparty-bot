package service

import (
	"context"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/repository"
	"github.com/qRe0/afterparty-bot/internal/shared"
)

const (
	OrgLace     = "Красный"
	VipLace     = "Синий"
	DefaultLace = "Желтый"
)

type TicketsService struct {
	repo *repository.TicketsRepo
}

func NewTicketsService(repo *repository.TicketsRepo) *TicketsService {
	return &TicketsService{
		repo: repo,
	}
}

func (ts *TicketsService) SearchByFullSurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if surname == nil || *surname == "" {
		msg := tgbotapi.NewMessage(-1, "service.SearchByFullSurname: Предоставлена пустая фамилия")
		bot.Send(msg)
		return
	}
	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchByFullSurname: Предоставлен пустой ID чата")
		bot.Send(msg)
		return
	}
	if bot == nil {
		log.Fatalln("service.SearchByFullSurname: Пустой инстанс бота")
	}

	formattedSurname := strings.ToLower(*surname)
	resp, err := ts.repo.SearchByFullSurname(ctx, formattedSurname)
	if err != nil || resp == nil {
		msg := tgbotapi.NewMessage(*chatID, "Покупатель с заданной фамилией не найден")
		bot.Send(msg)
		return
	}

	mappedResp := shared.ResponseMapper(resp)
	msg := tgbotapi.NewMessage(*chatID, mappedResp)
	bot.Send(msg)
}

func (ts *TicketsService) SearchBySurnamePart(ctx context.Context, surnamePart *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if surnamePart == nil || *surnamePart == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SearchBySurnamePart: Предоставлена пустая фамилия")
		bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchBySurnamePart: Предоставлен пустой ID чата")
		bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchBySurnamePart: Пустой инстанс бота")
	}

	formattedSurname := strings.ToLower(*surnamePart)
	respList, err := ts.repo.SearchBySurnamePart(ctx, formattedSurname)
	if err != nil || len(respList) == 0 {
		msg := tgbotapi.NewMessage(*chatID, "Покупателей с такими фамилиями не найдено.")
		bot.Send(msg)
		return
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(shared.ResponseMapper(&resp) + "\n\n")
	}

	msg := tgbotapi.NewMessage(*chatID, result.String())
	bot.Send(msg)
}

func (ts *TicketsService) SearchByID(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if userId == nil || *userId == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SearchByID: Предоставлен пустой ID")
		bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchByID: Предоставлен пустой ID чата")
		bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchByID: Пустой инстанс бота")
	}

	resp, err := ts.repo.SearchByID(ctx, *userId)
	if err != nil || resp == nil {
		msg := tgbotapi.NewMessage(*chatID, "Покупатель с данным ID не найден")
		bot.Send(msg)
		return
	}

	mappedResp := shared.ResponseMapper(resp)
	msg := tgbotapi.NewMessage(*chatID, mappedResp)
	bot.Send(msg)
}

func (ts *TicketsService) MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if userId == nil || *userId == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SearchByID: Предоставлен пустой ID")
		bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchByID: Предоставлен пустой ID чата")
		bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchByID: Пустой инстанс бота")
	}

	resp, err := ts.repo.MarkAsEntered(ctx, *userId)
	if err != nil || resp == nil {
		msg := tgbotapi.NewMessage(*chatID, "Покупатель с данным ID не найден")
		bot.Send(msg)
		return
	}

	mappedResp := shared.ResponseMapper(resp)
	msg := tgbotapi.NewMessage(*chatID, mappedResp)
	bot.Send(msg)
}
