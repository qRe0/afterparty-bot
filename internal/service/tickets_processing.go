package ticket_service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fogleman/gg"
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
	GetActualTicketNumber(ctx context.Context, ticketNo int64) (*int, error)
}

type TicketsService struct {
	repo *ticket_repository.TicketsRepo
	Cfg  configs.Config
}

func New(repo *ticket_repository.TicketsRepo, cfg configs.Config) *TicketsService {
	return &TicketsService{
		repo: repo,
		Cfg:  cfg,
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
		result.WriteString(utils.ResponseMapper(&resp, ts.Cfg.LacesColor) + "\n\n")
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
		msg := tgbotapi.NewMessage(*chatID, "Нет покупателей с указанным номером билета")
		_, _ = bot.Send(msg)
		return
	}

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	result.WriteString(utils.ResponseMapper(resp, ts.Cfg.LacesColor) + "\n\n")

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

func (ts *TicketsService) SellTicket(
	ctx context.Context,
	chatID int64,
	update tgbotapi.Update,
	bot *tgbotapi.BotAPI,
	client *models.ClientData,
) error {
	if client == nil {
		return fmt.Errorf("SellTicket: client is nil")
	}

	if bot == nil {
		return fmt.Errorf("SellTicket: bot is nil")
	}

	clientSurname := utils.GetSurnameLowercase(client.FIO)
	actualTicketPrice := utils.CalculateActualTicketPrice(time.Now(), ts.Cfg.SalesOption, *client)
	sellerTag := update.Message.From.UserName
	sellerId := update.Message.From.ID

	ticketNo, err := ts.repo.SellTicket(ctx, *client, "@"+sellerTag, clientSurname, actualTicketPrice)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			msg := tgbotapi.NewMessage(chatID, "Данный покупатель уже купил билет")
			_, _ = bot.Send(msg)
			return err
		}
		msg := tgbotapi.NewMessage(chatID, "Не удалось выполнить продажу билета. Попробуйте еще раз")
		_, _ = bot.Send(msg)
		return err
	}

	msgTmpl := fmt.Sprintf("Билет успешно продан!\nФИО покупателя: %s, номер билета: %d", client.FIO, ticketNo)
	msg := tgbotapi.NewMessage(chatID, msgTmpl)
	_, _ = bot.Send(msg)

	err = ts.repo.UpdateSellersTable(ctx, ticketNo, sellerId, "@"+sellerTag)
	if err != nil {
		log.Println("Не удалось обновить базу данных продавцов информацией о последней транзакции:", err)
	}

	if err := ts.addRowToGoogleSheet(ctx, *client, sellerTag, ticketNo); err != nil {
		log.Printf("Не удалось добавить данные в Google Таблицу: %v", err)
	}

	imageBuffer, err := ts.generateTicketImage(ticketNo)
	if err != nil {
		log.Println("Ошибка при генерации изображения")
		return nil
	}

	photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
		Name:  "ticket.png",
		Bytes: imageBuffer.Bytes(),
	})
	photoMsg.Caption = fmt.Sprintf("Покупатель: %s\nНомер билета: %d", client.FIO, ticketNo)
	if _, err = bot.Send(photoMsg); err != nil {
		log.Println("Ошибка при отправке фото")
	}

	return nil
}

func (ts *TicketsService) addRowToGoogleSheet(_ context.Context, client models.ClientData, sellerTag string, ticketNo int64) error {
	data := map[string]interface{}{
		"secret":     ts.Cfg.Sheet.Secret,
		"TableId":    ts.Cfg.Sheet.TableID,
		"TicketNo":   ticketNo,
		"FIO":        client.FIO,
		"TicketType": client.TicketType,
		"Price":      client.Price,
		"SellerTag":  sellerTag,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(ts.Cfg.Sheet.DeploymentURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	return nil
}

func (ts *TicketsService) generateTicketImage(ticketNo int64) (*bytes.Buffer, error) {
	const (
		backgroundPath = "ticket.png"
		fontPath       = "font.ttf"
		fontSize       = 82
		posX, posY     = 1450, 895
	)

	bg, err := gg.LoadImage(backgroundPath)
	if err != nil {
		log.Printf("не удалось загрузить фоновое изображение: %w", err)
		return nil, fmt.Errorf("не удалось загрузить фоновое изображение: %w", err)
	}

	dc := gg.NewContextForImage(bg)

	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		log.Printf("не удалось загрузить шрифт: %w", err)
		return nil, fmt.Errorf("не удалось загрузить шрифт: %w", err)
	}

	dc.SetHexColor("#f88707")
	ticketText := fmt.Sprintf("%d", ticketNo)
	dc.DrawStringAnchored(ticketText, posX, posY, 0.5, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		log.Printf("не удалось закодировать PNG: %w", err)
		return nil, fmt.Errorf("не удалось закодировать PNG: %w", err)
	}

	return &buf, nil
}
