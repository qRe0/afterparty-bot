package app

import (
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/handlers"
	"github.com/qRe0/afterparty-bot/internal/migrations"
	repo "github.com/qRe0/afterparty-bot/internal/repository"
	serv "github.com/qRe0/afterparty-bot/internal/service"
)

func Run() error {
	// Loading environmental variables
	cfg, err := configs.LoadEnv()
	if err != nil {
		return fmt.Errorf("app.LoadEnv(): failed to load env vars: %v", err)
	}

	// Database initialization
	db, err := repo.Init(cfg.DB)
	if err != nil {
		return fmt.Errorf("app.Init(): failed to init database: %v", err)
	}

	// DB migrator initialization
	migrator, err := migrations.NewMigrator(db)
	if err != nil {
		return fmt.Errorf("app.NewMigrator(): failed to init database migrtor: %v", err)
	}
	// Migrating database to latest version
	err = migrator.Latest()
	if err != nil {
		return fmt.Errorf("app.migrator.Latest(): failed to migrate database to latest version: %v", err)
	}

	// Bot instance initialization
	botInstance, err := tgbotapi.NewBotAPI(cfg.TG.Token)
	if err != nil {
		return fmt.Errorf("app.NewBotAPI(): failed to init telegram bot instance: %v", err)
	}
	botInstance.Debug = true

	// Layers initialization
	repository := repo.NewTicketsRepository(db, cfg.DB)
	service := serv.NewTicketsService(repository)
	handler := handlers.NewMessagesHandler(service)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 300
	updates := botInstance.GetUpdatesChan(u)

	// Channel to limit count of requests up to 5 at the same time
	ch := make(chan struct{}, 5)

	var wg sync.WaitGroup
	for update := range updates {
		if update.Message != nil {
			wg.Add(1)

			go func(update tgbotapi.Update) {
				defer wg.Done()

				// Waiting for data
				ch <- struct{}{}

				handler.HandleMessages(update, botInstance)

				// Clearing channel
				<-ch
			}(update)
		}
	}

	wg.Wait()

	return nil
}
