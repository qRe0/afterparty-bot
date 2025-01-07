package ticket_service

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/models"
	"github.com/qRe0/afterparty-bot/internal/repository"
	"github.com/qRe0/afterparty-bot/internal/shared"
)

type TicketsRepo interface {
	SearchBySurname(ctx context.Context, surname string) ([]models.TicketResponse, error)
	MarkAsEntered(ctx context.Context, id string) (*models.TicketResponse, error)
	CheckCountOfSurnames(ctx context.Context, surname string) (int64, error)
	SearchById(ctx context.Context, id string) (*models.TicketResponse, error)
}

type TicketsService struct {
	repo *ticket_repository.TicketsRepo
	cfg  configs.LacesColors
}

func New(repo *ticket_repository.TicketsRepo, cfg configs.LacesColors) *TicketsService {
	return &TicketsService{
		repo: repo,
		cfg:  cfg,
	}
}

func (ts *TicketsService) SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if surname == nil || *surname == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SearchBySurnamePart: Предоставлена пустая фамилия")
		_, _ = bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchBySurnamePart: Предоставлен пустой ID чата")
		_, _ = bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchBySurnamePart: Пустой инстанс бота")
	}

	formattedSurname := strings.ToLower(*surname)
	partSurnameToSearch := formattedSurname + "%"
	respList, err := ts.repo.SearchBySurname(ctx, partSurnameToSearch)
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Ошибка при поиске покупателя")
		_, _ = bot.Send(msg)
		return
	}
	if len(respList) == 0 {
		fullSurnameToSearch := formattedSurname
		newRespList, err := ts.repo.SearchBySurname(ctx, fullSurnameToSearch)
		if err != nil {
			msg := tgbotapi.NewMessage(*chatID, "Ошибка при поиске покупателя")
			_, _ = bot.Send(msg)
			return
		}
		if len(newRespList) == 0 {
			msg := tgbotapi.NewMessage(*chatID, "Нет покупателей с указанной фамилией")
			_, _ = bot.Send(msg)
			return
		}
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(shared.ResponseMapper(&resp, ts.cfg) + "\n\n")
	}

	msg := tgbotapi.NewMessage(*chatID, result.String())
	_, _ = bot.Send(msg)

	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	for _, resp := range respList {
		if resp.PassedControlZone == false {
			btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (ID: %s)", resp.Name, resp.Id), resp.Id)
			inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
		}
	}
	msg = tgbotapi.NewMessage(*chatID, "Выберите нужного покупателя, чтобы отметить вход:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
	_, _ = bot.Send(msg)
}

func (ts *TicketsService) SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if userId == nil || *userId == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SearchBySurnamePart: Предоставлен пустой номер билета (ID покупателя)")
		_, _ = bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchBySurnamePart: Предоставлен пустой ID чата")
		_, _ = bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchBySurnamePart: Пустой инстанс бота")
	}

	resp, err := ts.repo.SearchById(ctx, *userId)
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Ошибка при поиске покупателя")
		_, _ = bot.Send(msg)
		return
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	result.WriteString(shared.ResponseMapper(resp, ts.cfg) + "\n\n")

	msg := tgbotapi.NewMessage(*chatID, result.String())
	_, _ = bot.Send(msg)

	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	if resp.PassedControlZone == false {
		btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (ID: %s)", resp.Name, resp.Id), resp.Id)
		inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
	}
	msg = tgbotapi.NewMessage(*chatID, "Выберите нужного покупателя, чтобы отметить вход:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
	_, _ = bot.Send(msg)
}

func (ts *TicketsService) MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if userId == nil || *userId == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.MarkAsEntered: Предоставлен пустой ID")
		_, _ = bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.MarkAsEntered: Предоставлен пустой ID чата")
		_, _ = bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.MarkAsEntered: Пустой инстанс бота")
	}

	resp, err := ts.repo.MarkAsEntered(ctx, *userId)
	if err != nil || resp == nil {
		msg := tgbotapi.NewMessage(*chatID, "service.MarkAsEntered: Покупатель с данным ID не найден")
		_, _ = bot.Send(msg)
		return
	}

	mappedResp := fmt.Sprintf("%s прошел контроль (ID: %s)", resp.Name, resp.Id)
	msg := tgbotapi.NewMessage(*chatID, mappedResp)
	_, _ = bot.Send(msg)
}
