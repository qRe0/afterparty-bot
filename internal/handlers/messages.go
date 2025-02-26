package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/models"
	"github.com/qRe0/afterparty-bot/internal/service"
	utils "github.com/qRe0/afterparty-bot/internal/shared"
)

type TicketsService interface {
	SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI)
	SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (*models.TicketResponse, string, error)
	SellTicket(ctx context.Context, chatID int64, update tgbotapi.Update, bot *tgbotapi.BotAPI, client *models.ClientData) error
	MarkAsEntered(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (string, error)
}

type MessagesHandler struct {
	service    *ticket_service.TicketsService
	userStates map[int64]string
	clientData map[int64]*models.ClientData
	cfg        configs.AllowList
}

func New(service *ticket_service.TicketsService, cfg configs.AllowList) MessagesHandler {
	return MessagesHandler{
		service:    service,
		userStates: make(map[int64]string),
		clientData: make(map[int64]*models.ClientData),
		cfg:        cfg,
	}
}

func (mh *MessagesHandler) HandleMessages(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	ctx := context.Background()
	var chatID int64

	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		data := update.CallbackQuery.Data

		if strings.HasPrefix(data, "confirm_yes_") {
			userId := strings.TrimPrefix(data, "confirm_yes_")
			msg, err := mh.service.MarkAsEntered(ctx, &userId, &chatID, bot)
			if err != nil {
				log.Println("HandleMessages:: MarkAsEntered:: Error during MarkAsEntered service method (1st call)")
				botMsg := tgbotapi.NewMessage(chatID, msg)
				_, _ = bot.Send(botMsg)
			}
			botMsg := tgbotapi.NewMessage(chatID, msg)
			_, _ = bot.Send(botMsg)
		} else if strings.HasPrefix(data, "confirm_no_") {
			msg := tgbotapi.NewMessage(chatID, "Операция отменена.")
			_, _ = bot.Send(msg)
		} else {
			userId := data
			msg, err := mh.service.MarkAsEntered(ctx, &userId, &chatID, bot)
			if err != nil {
				log.Println("HandleMessages:: MarkAsEntered:: Error during MarkAsEntered service method (2nd call)")
				botMsg := tgbotapi.NewMessage(chatID, msg)
				_, _ = bot.Send(botMsg)
			}
			botMsg := tgbotapi.NewMessage(chatID, msg)
			_, _ = bot.Send(botMsg)
		}

		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		_, err := bot.Request(callback)
		if err != nil {
			log.Printf("HandleMessages:: Failed to send callback with error: %v", err)
		}
		return
	}

	if update.Message != nil {
		chatID = update.Message.Chat.ID
		text := update.Message.Text
		userName := update.Message.From.UserName

		switch text {
		case "/start":
			if !utils.UserInList(userName, mh.cfg.AllowedCheckers) && !utils.UserInList(userName, mh.cfg.AllowedSellers) {
				msg := tgbotapi.NewMessage(chatID, "У Вас нет прав на использование бота.")
				bot.Send(msg)
				return
			}
			mh.userStates[chatID] = ""
			utils.ShowOptions(chatID, bot)
			return

		case "Отметить вход":
			if !utils.UserInList(userName, mh.cfg.AllowedCheckers) {
				msg := tgbotapi.NewMessage(chatID, "У Вас нет прав для отметки входа.")
				bot.Send(msg)
				return
			}
			mh.userStates[chatID] = "awaiting_id_surname"
			msg := tgbotapi.NewMessage(chatID, "Введите фамилию или номер билета для поиска:")
			_, _ = bot.Send(msg)
			return

		case "Продать билет":
			if !utils.UserInList(userName, mh.cfg.AllowedSellers) {
				msg := tgbotapi.NewMessage(chatID, "У Вас нет прав для продажи билетов.")
				bot.Send(msg)
				return
			}
			mh.clientData[chatID] = &models.ClientData{}
			mh.userStates[chatID] = "awaiting_client_fio"
			msg := tgbotapi.NewMessage(chatID, "Введите ФИО покупателя:")
			_, _ = bot.Send(msg)
			return
		}

		switch mh.userStates[chatID] {
		case "awaiting_id_surname":
			if _, err := strconv.Atoi(text); err == nil {
				log.Println("messages.HandleMessages: Ищем пользователя по номеру билета")
				resp, respMsg, err := mh.service.SearchById(ctx, &update.Message.Text, &chatID, bot)
				if err != nil {
					log.Println("HandleMessages:: SearchById:: Error during MarkAsEntered service method")
					botMsg := tgbotapi.NewMessage(chatID, respMsg)
					_, _ = bot.Send(botMsg)
				}
				msg := tgbotapi.NewMessage(chatID, respMsg)
				_, _ = bot.Send(msg)

				var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
				if resp != nil {
					if resp.PassedControlZone == false {
						btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (ID: %s)", resp.Name, resp.Id), resp.Id)
						inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
					}
					msg = tgbotapi.NewMessage(chatID, "Выберите нужного покупателя, чтобы отметить вход:")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
					_, _ = bot.Send(msg)
				} else {
					log.Panic("HandleMessages:: SearchById:: Response is empty (nil)")
				}
			} else {
				log.Println("messages.HandleMessages: Ищем пользователя по фамилии")
				mh.service.SearchBySurname(ctx, &update.Message.Text, &chatID, bot)
			}
		case "awaiting_client_fio":
			if text == "" {
				msg := tgbotapi.NewMessage(chatID, "ФИО не может быть пустым. Введите ещё раз:")
				_, _ = bot.Send(msg)
				return
			}
			formattedFio, err := utils.FormatFIO(text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Проверьте введенное ФИО")
				_, _ = bot.Send(msg)
				return
			}
			mh.clientData[chatID].FIO = formattedFio
			msg := tgbotapi.NewMessage(chatID, "Введите тип билета (ВИПх или БАЗОВЫЙ):")
			_, _ = bot.Send(msg)
			mh.userStates[chatID] = "awaiting_client_ticketType"
		case "awaiting_client_ticketType":
			if text == "" {
				msg := tgbotapi.NewMessage(chatID, "Введите тип билета:")
				_, _ = bot.Send(msg)
				return
			}

			ticketType, ok := utils.ValidateTicketType(text, mh.service.Cfg.SalesOption)
			if !ok {
				msg := tgbotapi.NewMessage(chatID, "Неверный тип билета. Попробуйте ещё раз:")
				_, _ = bot.Send(msg)
				return
			}

			mh.clientData[chatID].TicketType = ticketType
			msg := tgbotapi.NewMessage(chatID, "Введите стоимость билета:")
			_, _ = bot.Send(msg)
			mh.userStates[chatID] = "awaiting_client_price"
		case "awaiting_client_price":
			if text == "" {
				msg := tgbotapi.NewMessage(chatID, "Цена не может быть пустой. Повторите ввод:")
				_, _ = bot.Send(msg)
				return
			}

			price, err := utils.ParseTicketPrice(text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Проверьте введенную цену. Попробуйте ещё раз:")
				_, _ = bot.Send(msg)
				return
			}
			mh.clientData[chatID].Price = price

			msg := tgbotapi.NewMessage(chatID, "Укажите наличие репоста (да/нет):")
			_, _ = bot.Send(msg)
			mh.userStates[chatID] = "awaiting_client_repost"

		case "awaiting_client_repost":
			if text == "" {
				msg := tgbotapi.NewMessage(chatID, "Ответ не может быть пустым. Укажите наличие репоста (да/нет):")
				_, _ = bot.Send(msg)
				return
			}

			mh.clientData[chatID].RepostExists = utils.CheckRepost(text)

			err := mh.service.SellTicket(ctx, chatID, update, bot, mh.clientData[chatID])
			if err != nil {
				log.Printf("Ошибка при продаже билета: %v", err)
			}

			mh.userStates[chatID] = ""
			delete(mh.clientData, chatID)
		}
	}
}
