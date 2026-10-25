package configs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	customErrors "github.com/qRe0/afterparty-bot/internal/errors"
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
	TableID       string `env:"TABLE_ID"`
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

type tempAllowList struct {
	AllowedSellers  []string `env:"ALLOWED_SELLERS"  envSeparator:","`
	AllowedCheckers []string `env:"ALLOWED_CHECKERS" envSeparator:","`
	VIPSellers      []string `env:"VIP_SELLERS"  envSeparator:","`
	SSSellers       []string `env:"SS_SELLERS"  envSeparator:","`
}

type AllowList struct {
	AllowedSellers  map[string]bool
	AllowedCheckers map[string]bool
	VIPSellers      map[string]bool
	SSSellers       map[string]bool
}

type TicketImageConfig struct {
	NumberPosX     int    `env:"TICKET_NUMBER_POS_X,required"`
	NumberPosY     int    `env:"TICKET_NUMBER_POS_Y,required"`
	NumberFontSize int    `env:"TICKET_NUMBER_FONT_SIZE,required"`
	NumberColor    string `env:"TICKET_NUMBER_COLOR"`
}

const defaultTicketNumberColor = "#ffffff"

var ticketNumberColorPattern = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

type Config struct {
	DB          DBConfig
	TG          TelegramAPIConfig
	LacesColor  LacesColors
	SalesOption SalesOptions
	Sheet       GoogleSheets
	AllowList   AllowList
	TicketImage TicketImageConfig
}

func LoadEnvs() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, customErrors.ErrLoadEnvVars
	}

	var (
		dbCfg        DBConfig
		tgConfig     TelegramAPIConfig
		lacesColor   LacesColors
		salesOptions SalesOptions
		sheet        GoogleSheets
		tmpAllowList tempAllowList
		ticketImage  TicketImageConfig
	)

	err = env.Parse(&dbCfg)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "DB")
	}

	err = env.Parse(&tgConfig)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "Telegram API")
	}

	err = env.Parse(&lacesColor)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "Lace colors")
	}

	err = env.Parse(&salesOptions)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "Sales options")
	}

	err = env.Parse(&sheet)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "Google Sheets")
	}

	err = env.Parse(&tmpAllowList)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "List of allowed users")
	}

	err = env.Parse(&ticketImage)
	if err != nil {
		return nil, errors.Wrap(customErrors.ErrLoadEnvVars, "Ticket image")
	}

	ticketImage, err = normalizeTicketImageConfig(ticketImage)
	if err != nil {
		return nil, fmt.Errorf("%w: ticket image: %v", customErrors.ErrLoadEnvVars, err)
	}

	cfg := &Config{
		DB:          dbCfg,
		TG:          tgConfig,
		LacesColor:  lacesColor,
		SalesOption: salesOptions,
		Sheet:       sheet,
		TicketImage: ticketImage,
		AllowList: AllowList{
			AllowedSellers:  SliceToMap(tmpAllowList.AllowedSellers),
			AllowedCheckers: SliceToMap(tmpAllowList.AllowedCheckers),
			VIPSellers:      SliceToMap(tmpAllowList.VIPSellers),
			SSSellers:       SliceToMap(tmpAllowList.SSSellers),
		},
	}

	return cfg, nil
}

func SliceToMap(userSlice []string) map[string]bool {
	mp := make(map[string]bool, len(userSlice))
	for _, user := range userSlice {
		mp[user] = true
	}

	return mp
}

func normalizeTicketImageConfig(cfg TicketImageConfig) (TicketImageConfig, error) {
	if cfg.NumberPosX < 0 {
		return cfg, fmt.Errorf("TICKET_NUMBER_POS_X must be >= 0, got %d", cfg.NumberPosX)
	}
	if cfg.NumberPosY < 0 {
		return cfg, fmt.Errorf("TICKET_NUMBER_POS_Y must be >= 0, got %d", cfg.NumberPosY)
	}
	if cfg.NumberFontSize <= 0 {
		return cfg, fmt.Errorf(
			"TICKET_NUMBER_FONT_SIZE must be > 0, got %d",
			cfg.NumberFontSize,
		)
	}

	color := strings.TrimSpace(cfg.NumberColor)
	if color == "" {
		cfg.NumberColor = defaultTicketNumberColor
		return cfg, nil
	}

	if !ticketNumberColorPattern.MatchString(color) {
		return cfg, fmt.Errorf(
			"TICKET_NUMBER_COLOR must be a six-digit hex color, got %q",
			cfg.NumberColor,
		)
	}
	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}
	cfg.NumberColor = strings.ToLower(color)

	return cfg, nil
}
