package ticket_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qRe0/afterparty-bot/internal/shared/logger"
	"github.com/qRe0/afterparty-bot/internal/shared/utils"
	"go.uber.org/zap"

	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lib/pq"
	"github.com/qRe0/afterparty-bot/internal/configs"
	errs "github.com/qRe0/afterparty-bot/internal/errors"
	"github.com/qRe0/afterparty-bot/internal/models"
	"github.com/qRe0/afterparty-bot/internal/repository"
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
	Cfg  configs.Config
	mu   sync.Mutex
}

func New(repo *ticket_repository.TicketsRepo, cfg configs.Config) *TicketsService {
	return &TicketsService{
		repo: repo,
		Cfg:  cfg,
	}
}

func (ts *TicketsService) SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) ([]models.TicketResponse, string, error) {
	lgr := logger.New(ctx)

	lgr.Info("TicketService:: Started SearchBySurname method call")

	if surname == nil || *surname == "" {
		msg := "Предоставлена пустая фамилия пользователя"
		lgr.Error("TicketService:: SearchBySurname:: Empty surname passed")
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "surname")
	}
	lgr.Debug("TicketsService:: SearchBySurname:: surname checked")

	if chatID == nil {
		msg := "Предоставлен пустая ID чата"
		lgr.Error("TicketService:: SearchBySurname:: Empty chatId passed")
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatId")
	}
	lgr.Debug("TicketsService:: SearchBySurname:: chatId checked")

	if bot == nil {
		lgr.Panic("TicketsService:: SearchBySurname:: Bot instance is empty (nil)")
	}

	formattedSurname := strings.ToLower(*surname)
	formattedSurname = strings.Replace(formattedSurname, "ё", "е", -1)
	partSurnameToSearch := formattedSurname + "%"
	respList, err := ts.repo.SearchBySurname(ctx, partSurnameToSearch)
	if err != nil {
		lgr.Error("TicketService:: SearchBySurname_1:: Repository method returned error", zap.Error(err))
		msg := "Ошибка при поиске по фамилии"
		return nil, msg, err
	}
	if len(respList) == 0 {
		lgr.Debug("TicketService:: SearchBySurname:: No clients found by part of surname")
		lgr.Debug("TicketService:: SearchBySurname:: Trying to find client by full surname")
		fullSurnameToSearch := formattedSurname
		newRespList, err := ts.repo.SearchBySurname(ctx, fullSurnameToSearch)
		if err != nil {
			lgr.Error("TicketService:: SearchBySurname_2:: Repository method returned error", zap.Error(err))
			msg := "Ошибка при поиске по фамилии"
			return nil, msg, err
		}
		if len(newRespList) == 0 {
			lgr.Info("TicketService:: SearchBySurname:: No clients found with specified surname")
			msg := "Не удалось найти клиента с указанной фамилией"
			return nil, msg, err
		}
	}
	lgr.Info("TicketsService:: SearchBySurname:: Repository method returned result successfully")

	var result strings.Builder
	result.WriteString("Найдены следующие покупатели:\n\n")
	for _, resp := range respList {
		result.WriteString(utils.ResponseMapper(&resp, ts.Cfg.LacesColor) + "\n\n")
	}

	lgr.Info("TicketsService:: Finished SearchBySurname method call")

	return respList, result.String(), nil
}

func (ts *TicketsService) SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (*models.TicketResponse, string, error) {
	lgr := logger.New(ctx)

	lgr.Info("TicketService:: Started SearchById method call")

	if userId == nil || *userId == "" {
		msg := "Предоставлен пустой ID пользователя"
		lgr.Error("TicketService:: SearchById:: Empty userId passed")
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "userId")
	}
	lgr.Debug("TicketsService:: SearchById:: userId checked")

	if chatID == nil {
		msg := "Предоставлен пустой ID чата"
		lgr.Error("TicketService:: SearchById:: Empty chatId passed")
		return nil, msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatId")
	}
	lgr.Debug("TicketsService:: SearchById:: chatId checked")

	if bot == nil {
		lgr.Panic("TicketsService:: SearchById:: Bot instance is empty (nil)")
	}

	resp, err := ts.repo.SearchById(ctx, *userId)
	if err != nil {
		lgr.Error("TicketService:: SearchById:: Repository method returned error", zap.Error(err))
		msg := "Не найдено пользователся с указанным номером билета или ФИО"
		return nil, msg, err
	}
	lgr.Info("TicketsService:: SearchById:: Repository method returned result successfully")

	var resultMsg strings.Builder
	resultMsg.WriteString("Найдены следующие покупатели:\n\n")
	resultMsg.WriteString(utils.ResponseMapper(resp, ts.Cfg.LacesColor) + "\n\n")

	lgr.Info("TicketsService:: Finished SearchById method call")

	return resp, resultMsg.String(), nil
}

func (ts *TicketsService) MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (string, error) {
	lgr := logger.New(ctx)

	lgr.Info("TicketService:: Started MarkAsEntered method call")

	if userId == nil || *userId == "" {
		msg := "Предоставлен пустой ID пользователя"
		lgr.Error("TicketService:: MarkAsEntered:: Empty userId passed")
		return msg, errors.Wrap(errs.ErrCheckingBaseParameters, "userId")
	}
	lgr.Debug("TicketsService:: MarkAsEntered:: userId checked")

	if chatID == nil {
		msg := "Предоставлен пустой ID чата"
		lgr.Error("TicketService:: MarkAsEntered:: Empty chatId passed")
		return msg, errors.Wrap(errs.ErrCheckingBaseParameters, "chatID")
	}
	lgr.Debug("TicketsService:: MarkAsEntered:: chatId checked")

	if bot == nil {
		lgr.Panic("TicketsService:: MarkAsEntered:: Bot instance is empty (nil)")
	}

	resp, err := ts.repo.MarkAsEntered(ctx, *userId)
	if err != nil || resp == nil {
		lgr.Error("TicketService:: MarkAsEntered:: Repository method returned error", zap.Error(err))
		msg := "Ошибка вызова метода репозитория MarkAsEntered"
		return msg, err
	}
	lgr.Info("TicketsService:: MarkAsEntered:: Repository method returned result successfully")

	lgr.Info("TicketsService:: Finished MarkAsEntered method call")

	mappedResp := fmt.Sprintf("%s прошел контроль (ID: %s)", resp.Name, resp.Id)
	return mappedResp, nil
}

func (ts *TicketsService) SellTicket(
	ctx context.Context,
	update tgbotapi.Update,
	bot *tgbotapi.BotAPI,
	client *models.ClientData,
) (string, *bytes.Buffer, bool, error) {
	lgr := logger.New(ctx)

	lgr.Info("TicketService:: Started SellTicket method call")

	if client == nil {
		msg := "Данные клиента не были предоставлены"
		lgr.Error("TicketService:: SellTicket:: Empty userId passed")
		return msg, nil, false, errors.Wrap(errs.ErrCheckingBaseParameters, "client")
	}
	lgr.Debug("TicketsService:: SellTicket:: client checked")

	if bot == nil {
		lgr.Panic("TicketsService:: SellTicket:: Bot instance is empty (nil)")
	}

	lgr.Debug("TicketsService:: SellTicket:: Starting data preparation to call repository layer")
	client.FIO = strings.Title(client.FIO)
	clientSurname := utils.GetSurnameLowercase(client.FIO)
	actualTicketPrice := utils.CalculateActualTicketPrice(time.Now(), ts.Cfg.SalesOption, *client)
	sellerTag := "@" + update.Message.From.UserName
	sellerId := update.Message.From.ID
	client.TicketType = strings.ToUpper(client.TicketType)
	lgr.Debug("TicketsService:: SellTicket:: All the data prepared to call repository layer")

	lgr.Debug("TicketsService:: SellTicket:: Calling repository method")
	ticketNo, err := ts.repo.SellTicket(ctx, *client, sellerTag, clientSurname, actualTicketPrice)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			lgr.Info("TicketService:: SellTicket:: This client already bought a ticket")
			msg := "Данный клиент уже купил билет"
			return msg, nil, false, err
		}
		lgr.Error("TicketService:: SellTicket:: Repository method returned error", zap.Error(err))
		msg := "Ошибка вызова метода репозитория SellTicket"
		return msg, nil, false, err
	}
	lgr.Info("TicketsService:: SellTicket:: Repository method returned result successfully")

	lgr.Debug("TicketsService:: SellTicket:: Trying to update seller's table")
	err = ts.repo.UpdateSellersTable(ctx, ticketNo, sellerId, sellerTag)
	if err != nil {
		lgr.Error("TicketService:: SellTicket:: Can't update sellers table with error: ", zap.Error(err))
	}
	lgr.Info("TicketsService:: SellTicket:: Sellers table updated successfully")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	lgr.Debug("TicketsService:: SellTicket:: Trying to add row to Google Sheet")
	err = ts.addRowToGoogleSheet(*client, sellerTag, ticketNo)
	if err != nil {
		lgr.Error("TicketService:: SellTicket:: Can't update Google Sheet with error: ", zap.Error(err))
		msgTmpl := "Не удалось записать данные в гугл таблицу. Напишите @yahor_malinouski номер билета (%v), который не был записан в гугл таблицу и сгенерируйте билет вручную через Canva"
		msg := fmt.Sprintf(msgTmpl, ticketNo)
		return msg, nil, false, err
	}
	lgr.Info("TicketsService:: SellTicket:: Google Sheet updated successfully")

	lgr.Debug("TicketsService:: SellTicket:: Trying to generate ticket image")
	ticketGenerated := true
	imageBuffer, err := ts.generateTicketImage(ticketNo)
	if err != nil {
		ticketGenerated = false
		lgr.Error("TicketService:: SellTicket:: Can't generate ticket image with error: ", zap.Error(err))
		msgTmpl := "Не удалось сгенерировать изображение билета. Cгенерируйте билет вручную через Canva (номер билета: %v)"
		msg := fmt.Sprintf(msgTmpl, ticketNo)
		return msg, nil, ticketGenerated, err
	}
	lgr.Info("TicketsService:: SellTicket:: Ticket image generated successfully")

	lgr.Info("TicketsService:: Finished SellTicket method call")

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
		backgroundPath = "assets/ticket.png"
		fontPath       = "assets/font.ttf"
		fontSize       = 105
		posX, posY     = 870, 415
	)

	bg, err := gg.LoadImage(backgroundPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load background image: %v", err)
	}

	dc := gg.NewContextForImage(bg)

	if err := dc.LoadFontFace(fontPath, fontSize); err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}

	dc.SetHexColor("#000000")
	ticketText := fmt.Sprintf("%d", ticketNo)
	dc.DrawStringAnchored(ticketText, posX, posY, 0.5, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, fmt.Errorf("failed to encode .png file: %v", err)
	}

	return &buf, nil
}
