package ticket_service

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/repository"
	"github.com/qRe0/afterparty-bot/internal/shared"
)

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

	countOfSurnames, err := ts.repo.CheckCountOfSurnames(ctx, formattedSurname)
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Покупатель с заданной фамилией не найден")
		bot.Send(msg)
		return
	}

	if countOfSurnames <= 1 {
		resp, err := ts.repo.SearchByFullSurname(ctx, formattedSurname)
		if err != nil || resp == nil {
			msg := tgbotapi.NewMessage(*chatID, "Покупатель с заданной фамилией не найден")
			bot.Send(msg)
			return
		}
		mappedResp := shared.ResponseMapper(resp, ts.cfg)
		msg := tgbotapi.NewMessage(*chatID, mappedResp)
		bot.Send(msg)

		if resp.PassedControlZone == false {
			yesButton := tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("confirm_yes_%s", resp.Id))
			noButton := tgbotapi.NewInlineKeyboardButtonData("Нет", fmt.Sprintf("confirm_no_%s", resp.Id))
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(yesButton, noButton),
			)
			confirmMsg := tgbotapi.NewMessage(*chatID, "Отметить вход?")
			confirmMsg.ReplyMarkup = keyboard
			bot.Send(confirmMsg)
			return
		}

		return
	} else {
		ts.SearchBySurnamePart(ctx, &formattedSurname, chatID, bot)
	}
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
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Ошибка при поиске покупателя")
		bot.Send(msg)
		return
	}

	if len(respList) == 0 {
		msg := tgbotapi.NewMessage(*chatID, "Нет покупателей с указанной фамилией, которые еще не прошли контроль")
		bot.Send(msg)
		return
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(shared.ResponseMapper(&resp, ts.cfg) + "\n\n")
	}

	msg := tgbotapi.NewMessage(*chatID, result.String())
	bot.Send(msg)

	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	for _, resp := range respList {
		if resp.PassedControlZone == false {
			btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (ID: %s)", resp.Name, resp.Id), resp.Id)
			inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
		}
	}
	msg = tgbotapi.NewMessage(*chatID, "Выберите нужного покупателя, чтобы отметить вход:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
	bot.Send(msg)
}

func (ts *TicketsService) MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if userId == nil || *userId == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.MarkAsEntered: Предоставлен пустой ID")
		bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.MarkAsEntered: Предоставлен пустой ID чата")
		bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.MarkAsEntered: Пустой инстанс бота")
	}

	resp, err := ts.repo.MarkAsEntered(ctx, *userId)
	if err != nil || resp == nil {
		msg := tgbotapi.NewMessage(*chatID, "service.MarkAsEntered: Покупатель с данным ID не найден")
		bot.Send(msg)
		return
	}

	mappedResp := fmt.Sprintf("%s прошел контроль (ID: %s)", resp.Name, resp.Id)
	msg := tgbotapi.NewMessage(*chatID, mappedResp)
	bot.Send(msg)
}
