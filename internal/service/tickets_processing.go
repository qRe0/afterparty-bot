package ticket_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"github.com/qRe0/afterparty-bot/internal/configs"
	errs "github.com/qRe0/afterparty-bot/internal/errors"
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
	repo   *ticket_repository.TicketsRepo
	Cfg    configs.Config
	mu     sync.Mutex
	logger *zap.Logger
}

func New(repo *ticket_repository.TicketsRepo, cfg configs.Config) *TicketsService {
	var lgr *zap.Logger
	if os.Getenv("APP_ENV") == "dev" {
		lgr = zap.Must(zap.NewDevelopment())
	} else if os.Getenv("APP_ENV") == "prod" {
		lgr = zap.Must(zap.NewProduction())
	}
	defer lgr.Sync()

	return &TicketsService{
		repo:   repo,
		Cfg:    cfg,
		logger: lgr,
	}
}

func (ts *TicketsService) SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) ([]models.TicketResponse, string, error) {
	if surname == nil || *surname == "" {
		msg := "TicketService:: SearchBySurname:: Предоставлена пустая фамилия пользователя"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "surname")
	}
	ts.logger.Debug("TicketsService:: SearchBySurname:: surname checked")

	if chatID == nil {
		msg := "TicketService:: SearchBySurname:: Предоставлен пустая ID чата"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatId")
	}
	ts.logger.Debug("TicketsService:: SearchBySurname:: chatId checked")

	if bot == nil {
		ts.logger.Panic("TicketsService:: SearchBySurname:: Bot instance is empty (nil)")
	}

	formattedSurname := strings.ToLower(*surname)
	partSurnameToSearch := formattedSurname + "%"
	respList, err := ts.repo.SearchBySurname(ctx, partSurnameToSearch)
	if err != nil {
		ts.logger.Warn("TicketService:: SearchBySurname_1:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: SearchBySurname:: Ошибка вызова метода репозитория SearchBySurname"
		return nil, msg, err
	}
	if len(respList) == 0 {
		ts.logger.Debug("TicketService:: SearchBySurname:: No clients found by part of surname")
		ts.logger.Debug("TicketService:: SearchBySurname:: Trying to find client by full surname")
		fullSurnameToSearch := formattedSurname
		newRespList, err := ts.repo.SearchBySurname(ctx, fullSurnameToSearch)
		if err != nil {
			ts.logger.Warn("TicketService:: SearchBySurname_2:: Repository method returned error", zap.Error(err))
			msg := "TicketService:: SearchBySurname:: Ошибка вызова метода репозитория SearchBySurname"
			return nil, msg, err
		}
		if len(newRespList) == 0 {
			ts.logger.Info("TicketService:: SearchBySurname:: No clients found with specified surname")
			msg := "TicketService:: SearchBySurname:: Не удалось найти клиента с указанной фамилией"
			return nil, msg, err
		}
	}
	ts.logger.Info("TicketsService:: SearchBySurname:: Repository method returned result successfully")

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(utils.ResponseMapper(&resp, ts.Cfg.LacesColor) + "\n\n")
	}

	return respList, result.String(), nil
}

func (ts *TicketsService) SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (*models.TicketResponse, string, error) {
	if userId == nil || *userId == "" {
		msg := "TicketService:: SearchById:: Предоставлен пустой ID пользователя"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "userId")
	}
	ts.logger.Debug("TicketsService:: SearchById:: userId checked")

	if chatID == nil {
		msg := "TicketService:: SearchById:: Предоставлен пустой ID чата"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatId")
	}
	ts.logger.Debug("TicketsService:: SearchById:: chatId checked")

	if bot == nil {
		ts.logger.Panic("TicketsService:: SearchById:: Bot instance is empty (nil)")
	}

	resp, err := ts.repo.SearchById(ctx, *userId)
	if err != nil {
		ts.logger.Warn("TicketService:: SearchById:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: SearchById:: Ошибка вызова метода репозитория SearchById"
		return nil, msg, err
	}
	ts.logger.Info("TicketsService:: SearchById:: Repository method returned result successfully")

	var resultMsg strings.Builder
	resultMsg.WriteString("Найдены следующие покупатели:\n\n")
	resultMsg.WriteString(utils.ResponseMapper(resp, ts.Cfg.LacesColor) + "\n\n")

	return resp, resultMsg.String(), nil
}

func (ts *TicketsService) MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (string, error) {
	if userId == nil || *userId == "" {
		msg := "TicketService:: MarkAsEntered:: Предоставлен пустой ID пользователя"
		return msg, errors.Wrap(errs.ErrCheckingBaseParameters, "userId")
	}
	ts.logger.Debug("TicketsService:: MarkAsEntered:: userId checked")

	if chatID == nil {
		msg := "TicketService:: MarkAsEntered:: Предоставлен пустой ID чата"
		return msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatID")
	}
	ts.logger.Debug("TicketsService:: MarkAsEntered:: chatId checked")

	if bot == nil {
		ts.logger.Panic("TicketsService:: MarkAsEntered:: Bot instance is empty (nil)")
	}

	resp, err := ts.repo.MarkAsEntered(ctx, *userId)
	if err != nil || resp == nil {
		ts.logger.Warn("TicketService:: MarkAsEntered:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: MarkAsEntered:: Ошибка вызова метода репозитория MarkAsEntered"
		return msg, err
	}
	ts.logger.Info("TicketsService:: MarkAsEntered:: Repository method returned result successfully")

	mappedResp := fmt.Sprintf("%s прошел контроль (ID: %s)", resp.Name, resp.Id)
	return mappedResp, nil
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

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if err := ts.addRowToGoogleSheet(*client, sellerTag, ticketNo); err != nil {
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

func (ts *TicketsService) addRowToGoogleSheet(client models.ClientData, sellerTag string, ticketNo int64) error {
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
		log.Printf("Не удалось загрузить фоновое изображение: %v", err)
		return nil, fmt.Errorf("failed to load background image: %v", err)
	}

	dc := gg.NewContextForImage(bg)

	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		log.Printf("Не удалось загрузить шрифт: %v", err)
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	dc.SetHexColor("#f88707")
	ticketText := fmt.Sprintf("%d", ticketNo)
	dc.DrawStringAnchored(ticketText, posX, posY, 0.5, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		log.Printf("Не удалось закодировать PNG: %v", err)
		return nil, fmt.Errorf("failed to encode .png file: %v", err)
	}

	return &buf, nil
}
