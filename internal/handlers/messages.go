package handlers

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/models"
	"github.com/qRe0/afterparty-bot/internal/service"
	"github.com/qRe0/afterparty-bot/internal/shared/logger"
	"github.com/qRe0/afterparty-bot/internal/shared/utils"
	"go.uber.org/zap"
)

type TicketsService interface {
	SearchBySurname(ctx context.Context, surname *string, chatID *int64, bot *tgbotapi.BotAPI) ([]models.TicketResponse, string, error)
	SearchById(ctx context.Context, userId *string, chatID *int64, bot *tgbotapi.BotAPI) (*models.TicketResponse, string, error)
	SellTicket(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI, client *models.ClientData) (string, *bytes.Buffer, bool, error)
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

func (mh *MessagesHandler) HandleMessages(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var chatID int64
	lgr := logger.New(ctx)

	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		data := update.CallbackQuery.Data

		if strings.HasPrefix(data, "confirm_yes_") {
			userId := strings.TrimPrefix(data, "confirm_yes_")
			msg, err := mh.service.MarkAsEntered(ctx, &userId, &chatID, bot)
			if err != nil {
				lgr.Warn("HandleMessages:: MarkAsEntered:: Error during MarkAsEntered service method (1st call) with error: ", zap.Error(err))
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
				lgr.Warn("HandleMessages:: MarkAsEntered:: Error during MarkAsEntered service method (2nd call) with error: ", zap.Error(err))
				botMsg := tgbotapi.NewMessage(chatID, msg)
				_, _ = bot.Send(botMsg)
			}
			botMsg := tgbotapi.NewMessage(chatID, msg)
			_, _ = bot.Send(botMsg)
		}

		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		_, err := bot.Request(callback)
		if err != nil {
			lgr.Warn("HandleMessages:: Failed to send callback with error: ", zap.Error(err))
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
				lgr.Info("Unauthorized user trying to use bot")
				msg := tgbotapi.NewMessage(chatID, "У Вас нет прав на использование бота.")
				_, _ = bot.Send(msg)
				return
			}
			mh.userStates[chatID] = ""
			utils.ShowOptions(chatID, bot, userName, mh.cfg)
			return

		case "Отметить вход":
			if !utils.UserInList(userName, mh.cfg.AllowedCheckers) {
				lgr.Info("Unauthorized user trying to use bot")
				msg := tgbotapi.NewMessage(chatID, "У Вас нет прав для отметки входа.")
				_, _ = bot.Send(msg)
				return
			}
			mh.userStates[chatID] = "awaiting_id_surname"
			msg := tgbotapi.NewMessage(chatID, "Введите фамилию или номер билета для поиска:")
			_, _ = bot.Send(msg)
			return

		case "Продать билет":
			if !utils.UserInList(userName, mh.cfg.AllowedSellers) {
				lgr.Info("Unauthorized user trying to use bot")
				msg := tgbotapi.NewMessage(chatID, "У Вас нет прав для продажи билетов.")
				_, _ = bot.Send(msg)
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
				resp, respMsg, err := mh.service.SearchById(ctx, &update.Message.Text, &chatID, bot)
				if err != nil {
					lgr.Warn("HandleMessages:: SearchById:: Error during MarkAsEntered service method", zap.Error(err))
					botMsg := tgbotapi.NewMessage(chatID, respMsg)
					_, _ = bot.Send(botMsg)
				}
				if respMsg != "Не найдено пользователся с указанным номером билета или ФИО" {
					msg := tgbotapi.NewMessage(chatID, respMsg)
					_, _ = bot.Send(msg)
				}

				var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
				if resp != nil {
					if !resp.PassedControlZone {
						btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (ID: %s)", resp.Name, resp.Id), resp.Id)
						inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
					}
					msg := tgbotapi.NewMessage(chatID, "Выберите нужного покупателя, чтобы отметить вход:")
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
					_, _ = bot.Send(msg)
				} else {
					lgr.Info("HandleMessages:: SearchById:: Response is nil. No clients found")
				}
			} else {
				respList, respMsg, err := mh.service.SearchBySurname(ctx, &update.Message.Text, &chatID, bot)
				if err != nil {
					lgr.Warn("HandleMessages:: SearchById:: Error during MarkAsEntered service method", zap.Error(err))
					botMsg := tgbotapi.NewMessage(chatID, respMsg)
					_, _ = bot.Send(botMsg)
				}
				msg := tgbotapi.NewMessage(chatID, respMsg)
				_, _ = bot.Send(msg)

				var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
				for _, resp := range respList {
					if !resp.PassedControlZone {
						btn := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s (ID: %s)", resp.Name, resp.Id), resp.Id)
						inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(btn))
					}
				}
				msg = tgbotapi.NewMessage(chatID, "Выберите нужного покупателя, чтобы отметить вход:")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
				_, _ = bot.Send(msg)
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

			baseButton := tgbotapi.NewKeyboardButton("Базовый")
			vipButton := tgbotapi.NewKeyboardButton("ВИП")
			var replyKeyboard tgbotapi.ReplyKeyboardMarkup

			if userName == mh.cfg.VIPSeller {
				replyKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(baseButton, vipButton),
				)
			} else {
				replyKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(baseButton),
				)
			}
			replyKeyboard.OneTimeKeyboard = true
			replyKeyboard.ResizeKeyboard = true

			msg := tgbotapi.NewMessage(chatID, "Выберите тип билета:")
			msg.ReplyMarkup = replyKeyboard
			_, _ = bot.Send(msg)

			mh.userStates[chatID] = "awaiting_client_ticket_type_choice"

		case "awaiting_client_ticket_type_choice":
			removeKeyboard := tgbotapi.NewRemoveKeyboard(true)

			if strings.ToLower(text) == "базовый" {
				mh.clientData[chatID].TicketType = "Базовый"

				removeMsg := tgbotapi.NewMessage(chatID, "Выбран тип: Базовый.")
				removeMsg.ReplyMarkup = removeKeyboard
				_, _ = bot.Send(removeMsg)

				msg := tgbotapi.NewMessage(chatID, "Введите стоимость билета:")
				_, _ = bot.Send(msg)
				mh.userStates[chatID] = "awaiting_client_price"
			} else if strings.ToLower(text) == "вип" {
				removeMsg := tgbotapi.NewMessage(chatID, "Выбран тип: ВИП. Введите номер столика:")
				removeMsg.ReplyMarkup = removeKeyboard
				_, _ = bot.Send(removeMsg)

				mh.userStates[chatID] = "awaiting_vip_table_number"
			} else {
				msg := tgbotapi.NewMessage(chatID, "Неверный выбор. Нажмите «Базовый» или «ВИП».")
				_, _ = bot.Send(msg)
			}
		case "awaiting_vip_table_number":
			tableNumber := text
			if tableNumber == "" {
				msg := tgbotapi.NewMessage(chatID, "Номер столика не может быть пустым. Введите ещё раз:")
				_, _ = bot.Send(msg)
				return
			}

			vipType := "ВИП" + tableNumber

			ticketType, ok := utils.ValidateTicketType(vipType, mh.service.Cfg.SalesOption)
			if !ok {
				msg := tgbotapi.NewMessage(chatID, "Неверный тип (возможно неправильный формат стола). Попробуйте ещё раз:")
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

			price, err := utils.ParseTicketPrice(text, userName, mh.cfg)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Проверьте введенную цену. Попробуйте ещё раз:")
				_, _ = bot.Send(msg)
				return
			}
			mh.clientData[chatID].Price = price

			yesButton := tgbotapi.NewKeyboardButton("Да")
			noButton := tgbotapi.NewKeyboardButton("Нет")
			replyKeyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(yesButton, noButton),
			)
			replyKeyboard.OneTimeKeyboard = true
			replyKeyboard.ResizeKeyboard = true

			msg := tgbotapi.NewMessage(chatID, "Укажите наличие репоста:")
			msg.ReplyMarkup = replyKeyboard
			_, _ = bot.Send(msg)
			mh.userStates[chatID] = "awaiting_client_repost"

		case "awaiting_client_repost":
			if text == "" {
				msg := tgbotapi.NewMessage(chatID, "Ответ не может быть пустым. Укажите наличие репоста (Да/Нет):")
				_, _ = bot.Send(msg)
				return
			}

			removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
			removeMsg := tgbotapi.NewMessage(chatID, "Ответ получен.")
			removeMsg.ReplyMarkup = removeKeyboard
			_, _ = bot.Send(removeMsg)

			mh.clientData[chatID].RepostExists = utils.CheckRepost(text)

			msg := tgbotapi.NewMessage(chatID, "Операция обрабатывается...")
			_, _ = bot.Send(msg)

			respMsg, imgBuffer, ticketGenerated, err := mh.service.SellTicket(ctx, update, bot, mh.clientData[chatID])
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, respMsg)
				_, _ = bot.Send(msg)
				return
			}

			if imgBuffer != nil && ticketGenerated {
				photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileBytes{
					Name:  "ticket.png",
					Bytes: imgBuffer.Bytes(),
				})
				photoMsg.Caption = respMsg
				_, _ = bot.Send(photoMsg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Не удалось отправить изображение")
				_, _ = bot.Send(msg)
			}

			mh.userStates[chatID] = ""
			delete(mh.clientData, chatID)

			utils.ShowOptions(chatID, bot, userName, mh.cfg)
		}
	}
}
