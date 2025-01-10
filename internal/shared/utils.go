package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qRe0/afterparty-bot/internal/configs"
	"github.com/qRe0/afterparty-bot/internal/models"
)

const (
	vipTicketTypeTemplate = "ВИП%d"
	formattedFIOTemplate  = "%s %s %s"
)

func ShowOptions(chatID int64, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, "Выберите опцию поиска покупателя:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Фамилия"),
			tgbotapi.NewKeyboardButton("Номер билета (ID)"),
			tgbotapi.NewKeyboardButton("Продать билет"),
		),
	)

	msg.ReplyMarkup = keyboard
	_, _ = bot.Send(msg)
}

func ResponseMapper(resp *models.TicketResponse, cfg configs.LacesColors) string {
	successEmoji := "✅"
	failEmoji := "❌"

	var laceColor string
	switch {
	case resp.TicketType == "ОРГ":
		laceColor = cfg.Org
	case strings.HasPrefix(resp.TicketType, "ВИП"):
		laceColor = cfg.VIP
	case resp.TicketType == "БАЗОВЫЙ":
		laceColor = cfg.Base
	default:
		return "Неизвестный тип билета"
	}

	controlStatus := failEmoji
	if resp.PassedControlZone {
		controlStatus = successEmoji
	}

	return fmt.Sprintf("Номер билета: %s,\nФИО: %s,\nТип браслета: %s,\nЦвет браслета: %s,\nПрошел контроль: %s",
		resp.Id, resp.Name, resp.TicketType, laceColor, controlStatus)
}

func ValidateTicketType(ticketType string, cfg configs.SalesOptions) (string, bool) {
	allowedTicketTypes := make([]string, 0)
	allowedTicketTypes = append(allowedTicketTypes, "БАЗОВЫЙ")
	for i := 0; i < cfg.VIPTablesCount; i++ {
		allowedTicketTypes = append(allowedTicketTypes, fmt.Sprintf(vipTicketTypeTemplate, i+1))
	}

	allowed := false
	for _, allowedTicketType := range allowedTicketTypes {
		if ticketType == allowedTicketType {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", false
	}

	return ticketType, true
}

func ParseTicketPrice(input string) (int, error) {
	if input == "" {
		return 0, fmt.Errorf("ticket price is not specified")
	}

	re := regexp.MustCompile(`\d+`)
	match := re.FindString(input)
	if match == "" {
		return 0, fmt.Errorf("failed to find digit in string: %s", input)
	}

	value, err := strconv.Atoi(match)
	if err != nil {
		return 0, fmt.Errorf("failes to parse string to int %q: %v", match, err)
	}

	return value, nil
}

func FormatFIO(fio string) (string, error) {
	var err error
	parts := strings.Fields(fio)

	if len(parts) != 3 {
		return "", err
	}

	return fmt.Sprintf(formattedFIOTemplate, parts[0], parts[1], parts[2]), nil
}

func CheckRepost(msg string) bool {
	formattedInput := strings.ToLower(msg)

	parts := strings.Fields(formattedInput)
	if len(parts) != 1 {
		return false
	}
	keyWord := parts[0]

	if keyWord != "репост" {
		return false
	}

	return true
}

func convertStringsToDates(dates []string) ([]time.Time, error) {
	if len(dates) == 0 {
		return nil, fmt.Errorf("dates slice can't be empty")
	}

	var result []time.Time
	for _, dateStr := range dates {
		t, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			return nil, fmt.Errorf("failes to parse date %q: %w", dateStr, err)
		}
		result = append(result, t)
	}

	return result, nil
}

// prices = {20,15,25,20,30} ->
// prices[0] - цена без репоста до повышения
// prices[1] - цена с репостом до повышения
// prices[2] - цена без репоста после повышения
// prices[3] - цена с репостом после повышения
// prices[4] - цена одного ВИП-билета

// dates = {xxx} ->
// dates[0] - дата повышения

func CalculateActualTicketPrice(timeNow time.Time, cfg configs.SalesOptions, client models.ClientData) int {
	dates, err := convertStringsToDates(cfg.Dates)
	if err != nil {
		return -1
	}

	switch {
	case client.TicketType == "БАЗОВЫЙ":
		if timeNow.Before(dates[0]) {
			if client.RepostExists {
				return cfg.Prices[1]
			} else {
				return cfg.Prices[0]
			}
		} else if timeNow.After(dates[0]) {
			if client.RepostExists {
				return cfg.Prices[3]
			} else {
				return cfg.Prices[2]
			}
		}
	case strings.HasPrefix(client.TicketType, "ВИП"):
		return cfg.Prices[4]
	}

	return -1
}

func GetSurnameLowercase(surname string) string {
	parts := strings.Split(surname, " ")

	if len(parts) != 3 {
		return ""
	}

	formattedSurname := strings.ToLower(parts[0])
	return formattedSurname
}
