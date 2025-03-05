package app

import (
	"context"
	"fmt"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/handlers"
	"github.com/qRe0/afterparty-bot/internal/migrations"
	"github.com/qRe0/afterparty-bot/internal/repository"
	"github.com/qRe0/afterparty-bot/internal/service"
	"github.com/qRe0/afterparty-bot/internal/shared/logger"
	"go.uber.org/zap"
)

func Run() error {
	ctx := context.Background()

	lgr := zap.Must(zap.NewDevelopment())
	if os.Getenv("APP_ENV") == "prod" {
		lgr = zap.Must(zap.NewProduction())
	}
	defer lgr.Sync()

	ctx = logger.InjectInContext(ctx, lgr)

	cfg, err := configs.LoadEnvs()
	if err != nil {
		return fmt.Errorf("app.LoadEnv(): failed to load env vars: %v", err)
	}
	lgr.Debug("Envs loaded successfully")

	db, err := ticket_repository.NewDatabaseConnection(cfg.DB)
	if err != nil {
		return fmt.Errorf("app.NewDatabaseConnection(): failed to init database: %v", err)
	}
	lgr.Debug("Database connection configured")

	m, err := migrator.New(db)
	if err != nil {
		return fmt.Errorf("app.migrator.New(): failed to init database migrtor: %v", err)
	}
	lgr.Debug("Migrator inited")
	err = m.Latest()
	if err != nil {
		return fmt.Errorf("app.m.Latest(): failed to migrate database to latest version: %v", err)
	}
	lgr.Debug("Database migrated successfully")

	botInstance, err := tgbotapi.NewBotAPI(cfg.TG.Token)
	if err != nil {
		return fmt.Errorf("app.NewBotAPI(): failed to init telegram bot instance: %v", err)
	}
	lgr.Debug("Bot API instance inited")

	repository := ticket_repository.New(db, cfg.DB)
	lgr.Debug("Repository layer inited")
	service := ticket_service.New(repository, *cfg)
	lgr.Debug("Service layer inited")
	handler := handlers.New(service, cfg.AllowList)
	lgr.Debug("Handler layer inited")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 300
	updates := botInstance.GetUpdatesChan(u)
	lgr.Info("App inited successfully")

	ch := make(chan struct{}, cfg.TG.UsersCount)

	var wg sync.WaitGroup
	for update := range updates {
		wg.Add(1)

		go func(update tgbotapi.Update) {
			defer wg.Done()
			ch <- struct{}{}
			handler.HandleMessages(ctx, update, botInstance)
			<-ch
		}(update)
	}

	wg.Wait()

	return nil
}
