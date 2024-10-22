package app

import (
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/handlers"
	repo "github.com/qRe0/afterparty-bot/internal/repository"
	serv "github.com/qRe0/afterparty-bot/internal/service"
)

func Run() error {
	// Загрузка переменных окружения
	cfg, err := configs.LoadEnv()
	if err != nil {
		return fmt.Errorf("app.LoadEnv(): не удалось загрузить переменные окружения: %v", err)
	}

	// Инициализация базы данных
	db, err := repo.Init(cfg.DB)
	if err != nil {
		return fmt.Errorf("app.Init(): не удалось инициализировать базу данных: %v", err)
	}

	// Инициализация бота
	botInstance, err := tgbotapi.NewBotAPI(cfg.TG.Token)
	if err != nil {
		return fmt.Errorf("app.NewBotAPI(): не удалось инициализировать экземпляр Telegram-бота: %v", err)
	}
	botInstance.Debug = true

	// Инициализация слоев проекта
	repository := repo.NewTicketsRepository(db, cfg.DB)
	service := serv.NewTicketsService(repository)
	handler := handlers.NewMessagesHandler(service)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 300
	updates := botInstance.GetUpdatesChan(u)

	// Создаем канал для ограничения одновременной обработки до 5 запросов
	ch := make(chan struct{}, 5)

	var wg sync.WaitGroup
	for update := range updates {
		if update.Message != nil {
			wg.Add(1)

			// Запускаем горутину для обработки каждого сообщения
			go func(update tgbotapi.Update) {
				defer wg.Done()

				// Ожидание разрешения через канал
				ch <- struct{}{}

				// Обработка сообщения
				handler.HandleMessages(update, botInstance)

				// Освобождаем канал
				<-ch
			}(update)
		}
	}

	wg.Wait()

	return nil
}
