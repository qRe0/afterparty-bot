package configs

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var (
	ErrLoadEnvVars = errors.New("failed to load env vars")
)

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	Port     string `env:"DB_PORT"`
}

type TelegramAPIConfig struct {
	Token string `env:"TELEGRAM_TOKEN"`
}

type Config struct {
	DB DBConfig
	TG TelegramAPIConfig
}

func LoadEnv() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, ErrLoadEnvVars
	}

	var dbCfg DBConfig
	err = env.Parse(&dbCfg)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "DB")
	}

	var tgConfig TelegramAPIConfig
	err = env.Parse(&tgConfig)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "Telegram API")
	}

	cfg := &Config{
		DB: dbCfg,
		TG: tgConfig,
	}

	return cfg, nil
}
