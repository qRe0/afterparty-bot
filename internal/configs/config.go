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
	Token      string `env:"TELEGRAM_TOKEN"`
	UsersCount int    `env:"USERS_COUNT"`
}

type GoogleSheets struct {
	Secret        string `env:"SECRET_KEY"`
	DeploymentURL string `env:"DEPLOYMENT_URL"`
}

type LacesColors struct {
	Base string `env:"BASE_LACE"`
	VIP  string `env:"VIP_LACE"`
	Org  string `env:"ORG_LACE"`
}

type SalesOptions struct {
	VIPTablesCount int      `env:"VIP_TABLES_COUNT"`
	Prices         []int    `env:"PRICES" envSeparator:","`
	Dates          []string `env:"DATES" envSeparator:","`
}

type Config struct {
	DB          DBConfig
	TG          TelegramAPIConfig
	LacesColor  LacesColors
	SalesOption SalesOptions
	Sheet       GoogleSheets
}

func LoadEnvs() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, ErrLoadEnvVars
	}

	var (
		dbCfg        DBConfig
		tgConfig     TelegramAPIConfig
		lacesColor   LacesColors
		salesOptions SalesOptions
		sheet        GoogleSheets
	)

	err = env.Parse(&dbCfg)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "DB")
	}

	err = env.Parse(&tgConfig)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "Telegram API")
	}

	err = env.Parse(&lacesColor)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "Lace colors")
	}

	err = env.Parse(&salesOptions)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "Sales options")
	}

	err = env.Parse(&sheet)
	if err != nil {
		return nil, errors.Wrap(ErrLoadEnvVars, "Google Sheets")
	}

	cfg := &Config{
		DB:          dbCfg,
		TG:          tgConfig,
		LacesColor:  lacesColor,
		SalesOption: salesOptions,
		Sheet:       sheet,
	}

	return cfg, nil
}
