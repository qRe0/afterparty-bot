package app

import (
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/handlers"
	"github.com/qRe0/afterparty-bot/internal/migrations"
	"github.com/qRe0/afterparty-bot/internal/repository"
	"github.com/qRe0/afterparty-bot/internal/service"
)

func Run() error {
	cfg, err := configs.LoadEnvs()
	if err != nil {
		return fmt.Errorf("app.LoadEnv(): failed to load env vars: %v", err)
	}
	log.Println("Envs loaded successfully")

	db, err := ticket_repository.NewDatabaseConnection(cfg.DB)
	if err != nil {
		return fmt.Errorf("app.NewDatabaseConnection(): failed to init database: %v", err)
	}
	log.Println("Database connection configured")

	m, err := migrator.New(db)
	if err != nil {
		return fmt.Errorf("app.migrator.New(): failed to init database migrtor: %v", err)
	}
	log.Println("Migrator inited")
	err = m.Latest()
	if err != nil {
		return fmt.Errorf("app.m.Latest(): failed to migrate database to latest version: %v", err)
	}
	log.Println("Database migrated successfully")

	botInstance, err := tgbotapi.NewBotAPI(cfg.TG.Token)
	if err != nil {
		return fmt.Errorf("app.NewBotAPI(): failed to init telegram bot instance: %v", err)
	}
	log.Println("Bot API instance inited")

	repository := ticket_repository.New(db, cfg.DB)
	log.Println("Repository layer inited")
	service := ticket_service.New(repository, cfg.LacesColor)
	log.Println("Service layer inited")
	handler := handlers.New(service)
	log.Println("Handler layer inited")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 300
	updates := botInstance.GetUpdatesChan(u)
	log.Println("App inited successfully")

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
