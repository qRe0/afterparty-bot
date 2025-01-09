package ticket_service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
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
	SellTicket(ctx context.Context, client models.ClientData, seller string, clientSurname string, actualPrice int) (int64, error)
	UpdateSellersTable(ctx context.Context, ticketId, sellerId int64, seller string) error
}

type TicketsService struct {
	repo *ticket_repository.TicketsRepo
	cfg  configs.Config
}

func New(repo *ticket_repository.TicketsRepo, cfg configs.Config) *TicketsService {
	return &TicketsService{
		repo: repo,
		cfg:  cfg,
	}
}

func (ts *TicketsService) SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) {
	if surname == nil || *surname == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SearchBySurname: Предоставлена пустая фамилия")
		_, _ = bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchBySurname: Предоставлен пустой ID чата")
		_, _ = bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchBySurname: Пустой инстанс бота")
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
		result.WriteString(shared.ResponseMapper(&resp, ts.cfg.LacesColor) + "\n\n")
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
		msg := tgbotapi.NewMessage(*chatID, "service.SearchById: Предоставлен пустой номер билета (ID покупателя)")
		_, _ = bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SearchById: Предоставлен пустой ID чата")
		_, _ = bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SearchById: Пустой инстанс бота")
	}

	resp, err := ts.repo.SearchById(ctx, *userId)
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Ошибка при поиске покупателя")
		_, _ = bot.Send(msg)
		return
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	result.WriteString(shared.ResponseMapper(resp, ts.cfg.LacesColor) + "\n\n")

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
		msg := tgbotapi.NewMessage(*chatID, "Покупатель с данным ID не найден")
		_, _ = bot.Send(msg)
		return
	}

	mappedResp := fmt.Sprintf("%s прошел контроль (ID: %s)", resp.Name, resp.Id)
	msg := tgbotapi.NewMessage(*chatID, mappedResp)
	_, _ = bot.Send(msg)
}

func (ts *TicketsService) SellTicket(ctx context.Context, update *tgbotapi.Update, chatID *int64, bot *tgbotapi.BotAPI) {
	clientDetails := &update.Message.Text

	if clientDetails == nil || *clientDetails == "" {
		msg := tgbotapi.NewMessage(*chatID, "service.SellTicket: Предоставлены пустые данные клиента")
		_, _ = bot.Send(msg)
		return
	}

	if chatID == nil {
		msg := tgbotapi.NewMessage(-1, "service.SellTicket: Предоставлен пустой ID чата")
		_, _ = bot.Send(msg)
		return
	}

	if bot == nil {
		log.Fatalln("service.SellTicket: Пустой инстанс бота")
	}

	formattedInput := strings.Split(*clientDetails, "; ")
	if len(formattedInput) != 4 {
		msg := tgbotapi.NewMessage(*chatID, "Проверьте формат введеных даных")
		_, _ = bot.Send(msg)
		return
	}

	var client models.ClientData
	formattedFio, err := shared.FormatFIO(formattedInput[0])
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Проверьте введенное ФИО")
		_, _ = bot.Send(msg)
		return
	}
	client.FIO = formattedFio

	ticketType, ok := shared.ValidateTicketType(formattedInput[1], ts.cfg.SalesOption)
	if !ok {
		msg := tgbotapi.NewMessage(*chatID, "Проверьте тип билета или порядок введенных данных. Формат: ФИО; Тип билета; Цена")
		_, _ = bot.Send(msg)
		return
	}
	client.TicketType = ticketType

	price, err := shared.ParseTicketPrice(formattedInput[2])
	if err != nil {
		msg := tgbotapi.NewMessage(*chatID, "Проверьте введенную цену")
		_, _ = bot.Send(msg)
		return
	}
	client.Price = price

	exists := shared.CheckRepost(formattedInput[3])
	if !exists {
		client.RepostExists = false
	} else {
		client.RepostExists = true
	}

	actualTicketPrice := shared.CalculateActualTicketPrice(time.Now(), ts.cfg.SalesOption, client)
	sellerTag := update.Message.From.UserName
	sellerId := update.Message.From.ID
	clientSurname := shared.GetSurnameLowercase(client.FIO)

	insertedId, err := ts.repo.SellTicket(ctx, client, sellerTag, clientSurname, actualTicketPrice)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			msg := tgbotapi.NewMessage(*chatID, "Данный покупатель уже купил билет")
			_, _ = bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(*chatID, "Не удалось выполнить продажу билета. Попробуйте еще раз")
		_, _ = bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(*chatID, "Билет успешно продан")
	_, _ = bot.Send(msg)

	err = ts.repo.UpdateSellersTable(ctx, insertedId, sellerId, sellerTag)
	if err != nil {
		log.Println("Failed to update sellers table with data of latest ticket transaction")
		return
	}
}
