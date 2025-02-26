package app

import (
	"fmt"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/handlers"
	"github.com/qRe0/afterparty-bot/internal/migrations"
	"github.com/qRe0/afterparty-bot/internal/repository"
	"github.com/qRe0/afterparty-bot/internal/service"
	"go.uber.org/zap"
)

func Run() error {
	var logger *zap.Logger
	if os.Getenv("APP_ENV") == "dev" {
		logger = zap.Must(zap.NewDevelopment())
	} else if os.Getenv("APP_ENV") == "prod" {
		logger = zap.Must(zap.NewProduction())
	}
	defer logger.Sync()

	cfg, err := configs.LoadEnvs()
	if err != nil {
		return fmt.Errorf("app.LoadEnv(): failed to load env vars: %v", err)
	}
	logger.Debug("Envs loaded successfully")

	db, err := ticket_repository.NewDatabaseConnection(cfg.DB)
	if err != nil {
		return fmt.Errorf("app.NewDatabaseConnection(): failed to init database: %v", err)
	}
	logger.Debug("Database connection configured")

	m, err := migrator.New(db)
	if err != nil {
		return fmt.Errorf("app.migrator.New(): failed to init database migrtor: %v", err)
	}
	logger.Debug("Migrator inited")
	err = m.Latest()
	if err != nil {
		return fmt.Errorf("app.m.Latest(): failed to migrate database to latest version: %v", err)
	}
	logger.Debug("Database migrated successfully")

	botInstance, err := tgbotapi.NewBotAPI(cfg.TG.Token)
	if err != nil {
		return fmt.Errorf("app.NewBotAPI(): failed to init telegram bot instance: %v", err)
	}
	logger.Debug("Bot API instance inited")

	repository := ticket_repository.New(db, cfg.DB)
	logger.Debug("Repository layer inited")
	service := ticket_service.New(repository, *cfg)
	logger.Debug("Service layer inited")
	handler := handlers.New(service, cfg.AllowList)
	logger.Debug("Handler layer inited")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 300
	updates := botInstance.GetUpdatesChan(u)
	logger.Info("App inited successfully")

	ch := make(chan struct{}, cfg.TG.UsersCount)

	var wg sync.WaitGroup
	for update := range updates {
		wg.Add(1)

		go func(update tgbotapi.Update) {
			defer wg.Done()
			ch <- struct{}{}
			handler.HandleMessages(update, botInstance)
			<-ch
		}(update)
	}

	wg.Wait()

	return nil
}
