package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/models"
	"github.com/qRe0/afterparty-bot/internal/repository"
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

	resp, err := ts.repo.SearchByFullSurname(ctx, *surname)
	if err != nil || resp == nil {
		msg := tgbotapi.NewMessage(*chatID, "Покупатель с заданной фамилией не найден")
		bot.Send(msg)
		return
	}

	mappedResp := responseMapper(resp)
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

	respList, err := ts.repo.SearchBySurnamePart(ctx, *surnamePart)
	if err != nil || len(respList) == 0 {
		msg := tgbotapi.NewMessage(*chatID, "Покупателей с такими фамилиями не найдено.")
		bot.Send(msg)
		return
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(responseMapper(&resp) + "\n\n")
	}

	msg := tgbotapi.NewMessage(*chatID, result.String())
	bot.Send(msg)
}

func responseMapper(resp *models.TicketResponse) string {
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
