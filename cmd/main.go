package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/qRe0/afterparty-bot/configs"
)

func main() {
	// Loading envs
	cfg, err := configs.LoadEnv()
	if err != nil {
		log.Fatalf("failed to load env vars: %v", err)
	}

	// Bot initialization
	bot, err := tgbotapi.NewBotAPI(cfg.TG.Token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	connectingStringTemplate := "postgres://%s:%s@%s:%s/%s?sslmode=disable"

	connStr := fmt.Sprintf(connectingStringTemplate, cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("Connected to DB successfully!")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(update, bot, db)
		}
	}
}

func handleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI, db *sqlx.DB) {
	switch update.Message.Text {
	case "/start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажмите кнопку, чтобы найти по фамилии.")
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Найти по фамилии"),
			),
		)
		msg.ReplyMarkup = keyboard
		bot.Send(msg)

	case "Найти по фамилии":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите фамилию для поиска:")
		bot.Send(msg)

	default:
		searchBySurname(update.Message.Text, update.Message.Chat.ID, bot, db)
	}
}

func searchBySurname(surname string, chatID int64, bot *tgbotapi.BotAPI, db *sqlx.DB) {
	var name, otherInfo string
	err := db.QueryRow("SELECT full_name, ticket_type FROM tickets WHERE surname=$1", surname).Scan(&name, &otherInfo)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Пользователь не найден.")
		bot.Send(msg)
		return
	}

	response := fmt.Sprintf("Имя: %s\nДоп. информация: %s", name, otherInfo)
	msg := tgbotapi.NewMessage(chatID, response)
	bot.Send(msg)
}
