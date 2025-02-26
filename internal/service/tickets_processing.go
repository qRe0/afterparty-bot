package ticket_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	Logger *zap.Logger
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
		Logger: lgr,
	}
}

func (ts *TicketsService) SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) ([]models.TicketResponse, string, error) {
	ts.Logger.Info("TicketService:: Started SearchBySurname method call")

	if surname == nil || *surname == "" {
		msg := "TicketService:: SearchBySurname:: Предоставлена пустая фамилия пользователя"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "surname")
	}
	ts.Logger.Debug("TicketsService:: SearchBySurname:: surname checked")

	if chatID == nil {
		msg := "TicketService:: SearchBySurname:: Предоставлен пустая ID чата"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatId")
	}
	ts.Logger.Debug("TicketsService:: SearchBySurname:: chatId checked")

	if bot == nil {
		ts.Logger.Panic("TicketsService:: SearchBySurname:: Bot instance is empty (nil)")
	}

	formattedSurname := strings.ToLower(*surname)
	partSurnameToSearch := formattedSurname + "%"
	respList, err := ts.repo.SearchBySurname(ctx, partSurnameToSearch)
	if err != nil {
		ts.Logger.Warn("TicketService:: SearchBySurname_1:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: SearchBySurname:: Ошибка вызова метода репозитория SearchBySurname"
		return nil, msg, err
	}
	if len(respList) == 0 {
		ts.Logger.Debug("TicketService:: SearchBySurname:: No clients found by part of surname")
		ts.Logger.Debug("TicketService:: SearchBySurname:: Trying to find client by full surname")
		fullSurnameToSearch := formattedSurname
		newRespList, err := ts.repo.SearchBySurname(ctx, fullSurnameToSearch)
		if err != nil {
			ts.Logger.Warn("TicketService:: SearchBySurname_2:: Repository method returned error", zap.Error(err))
			msg := "TicketService:: SearchBySurname:: Ошибка вызова метода репозитория SearchBySurname"
			return nil, msg, err
		}
		if len(newRespList) == 0 {
			ts.Logger.Info("TicketService:: SearchBySurname:: No clients found with specified surname")
			msg := "TicketService:: SearchBySurname:: Не удалось найти клиента с указанной фамилией"
			return nil, msg, err
		}
	}
	ts.Logger.Info("TicketsService:: SearchBySurname:: Repository method returned result successfully")

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(utils.ResponseMapper(&resp, ts.Cfg.LacesColor) + "\n\n")
	}

	ts.Logger.Info("TicketsService:: Finished SearchBySurname method call")

	return respList, result.String(), nil
}

func (ts *TicketsService) SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (*models.TicketResponse, string, error) {
	ts.Logger.Info("TicketService:: Started SearchById method call")

	if userId == nil || *userId == "" {
		msg := "TicketService:: SearchById:: Предоставлен пустой ID пользователя"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "userId")
	}
	ts.Logger.Debug("TicketsService:: SearchById:: userId checked")

	if chatID == nil {
		msg := "TicketService:: SearchById:: Предоставлен пустой ID чата"
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatId")
	}
	ts.Logger.Debug("TicketsService:: SearchById:: chatId checked")

	if bot == nil {
		ts.Logger.Panic("TicketsService:: SearchById:: Bot instance is empty (nil)")
	}

	resp, err := ts.repo.SearchById(ctx, *userId)
	if err != nil {
		ts.Logger.Warn("TicketService:: SearchById:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: SearchById:: Ошибка вызова метода репозитория SearchById"
		return nil, msg, err
	}
	ts.Logger.Info("TicketsService:: SearchById:: Repository method returned result successfully")

	var resultMsg strings.Builder
	resultMsg.WriteString("Найдены следующие покупатели:\n\n")
	resultMsg.WriteString(utils.ResponseMapper(resp, ts.Cfg.LacesColor) + "\n\n")

	ts.Logger.Info("TicketsService:: Finished SearchById method call")

	return resp, resultMsg.String(), nil
}

func (ts *TicketsService) MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (string, error) {
	ts.Logger.Info("TicketService:: Started MarkAsEntered method call")

	if userId == nil || *userId == "" {
		msg := "TicketService:: MarkAsEntered:: Предоставлен пустой ID пользователя"
		return msg, errors.Wrap(errs.ErrCheckingBaseParameters, "userId")
	}
	ts.Logger.Debug("TicketsService:: MarkAsEntered:: userId checked")

	if chatID == nil {
		msg := "TicketService:: MarkAsEntered:: Предоставлен пустой ID чата"
		return msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatID")
	}
	ts.Logger.Debug("TicketsService:: MarkAsEntered:: chatId checked")

	if bot == nil {
		ts.Logger.Panic("TicketsService:: MarkAsEntered:: Bot instance is empty (nil)")
	}

	resp, err := ts.repo.MarkAsEntered(ctx, *userId)
	if err != nil || resp == nil {
		ts.Logger.Warn("TicketService:: MarkAsEntered:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: MarkAsEntered:: Ошибка вызова метода репозитория MarkAsEntered"
		return msg, err
	}
	ts.Logger.Info("TicketsService:: MarkAsEntered:: Repository method returned result successfully")

	ts.Logger.Info("TicketsService:: Finished MarkAsEntered method call")

	mappedResp := fmt.Sprintf("%s прошел контроль (ID: %s)", resp.Name, resp.Id)
	return mappedResp, nil
}

func (ts *TicketsService) SellTicket(
	ctx context.Context,
	update tgbotapi.Update,
	bot *tgbotapi.BotAPI,
	client *models.ClientData,
) (string, *bytes.Buffer, bool, error) {
	ts.Logger.Info("TicketService:: Started SellTicket method call")

	if client == nil {
		msg := "TicketService:: SellTicket:: Данные клиента не были предоставлены"
		return msg, nil, false, errors.Wrap(errs.ErrCheckingBaseParameters, "client")
	}
	ts.Logger.Debug("TicketsService:: SellTicket:: client checked")

	if bot == nil {
		ts.Logger.Panic("TicketsService:: SellTicket:: Bot instance is empty (nil)")
	}

	ts.Logger.Debug("TicketsService:: SellTicket:: Starting data preparation to call repository layer")
	clientSurname := utils.GetSurnameLowercase(client.FIO)
	actualTicketPrice := utils.CalculateActualTicketPrice(time.Now(), ts.Cfg.SalesOption, *client)
	sellerTag := update.Message.From.UserName
	sellerId := update.Message.From.ID
	ts.Logger.Debug("TicketsService:: SellTicket:: All the data prepared to call repository layer")

	ts.Logger.Debug("TicketsService:: SellTicket:: Calling repository method")
	ticketNo, err := ts.repo.SellTicket(ctx, *client, "@"+sellerTag, clientSurname, actualTicketPrice)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			ts.Logger.Info("TicketService:: SellTicket:: This client already bought a ticket")
			msg := "TicketService:: SellTicket:: Данный клиент уже купил билет"
			return msg, nil, false, err
		}
		ts.Logger.Warn("TicketService:: SellTicket:: Repository method returned error", zap.Error(err))
		msg := "TicketService:: SellTicket:: Ошибка вызова метода репозитория SellTicket"
		return msg, nil, false, err
	}
	ts.Logger.Info("TicketsService:: SellTicket:: Repository method returned result successfully")

	ts.Logger.Debug("TicketsService:: SellTicket:: Trying to update seller's table")
	err = ts.repo.UpdateSellersTable(ctx, ticketNo, sellerId, "@"+sellerTag)
	if err != nil {
		ts.Logger.Warn("TicketService:: SellTicket:: Can't update sellers table with error: ", zap.Error(err))
	}
	ts.Logger.Info("TicketsService:: SellTicket:: Sellers table updated successfully")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.Logger.Debug("TicketsService:: SellTicket:: Trying to add row to Google Sheet")
	err = ts.addRowToGoogleSheet(*client, sellerTag, ticketNo)
	if err != nil {
		ts.Logger.Warn("TicketService:: SellTicket:: Can't update Google Sheet with error: ", zap.Error(err))
		msgTmpl := "TicketService:: SellTicket:: Не удалось записать данные в гугл таблицу. Напишите @yahor_malinouski номер билета (%v), который не был записан в гугл таблицу и сгенерируйте билет вручную через Canva"
		msg := fmt.Sprintf(msgTmpl, ticketNo)
		return msg, nil, false, err
	}
	ts.Logger.Info("TicketsService:: SellTicket:: Google Sheet updated successfully")

	ts.Logger.Debug("TicketsService:: SellTicket:: Trying to generate ticket image")
	ticketGenerated := true
	imageBuffer, err := ts.generateTicketImage(ticketNo)
	if err != nil {
		ticketGenerated = false
		ts.Logger.Warn("TicketService:: SellTicket:: Can't generate ticket image with error: ", zap.Error(err))
		msgTmpl := "TicketService:: SellTicket:: Не удалось сгенерировать изображение билета. Cгенерируйте билет вручную через Canva (номер билета: %v)"
		msg := fmt.Sprintf(msgTmpl, ticketNo)
		return msg, nil, ticketGenerated, err
	}
	ts.Logger.Info("TicketsService:: SellTicket:: Ticket image generated successfully")

	ts.Logger.Info("TicketsService:: Finished SellTicket method call")

	msg := fmt.Sprintf("Билет успешно продан!\nФИО покупателя: %s\nНомер билета: %d", client.FIO, ticketNo)
	return msg, imageBuffer, ticketGenerated, nil
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
		return nil, fmt.Errorf("failed to load background image: %v", err)
	}

	dc := gg.NewContextForImage(bg)

	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	dc.SetHexColor("#f88707")
	ticketText := fmt.Sprintf("%d", ticketNo)
	dc.DrawStringAnchored(ticketText, posX, posY, 0.5, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode .png file: %v", err)
	}

	return &buf, nil
}
