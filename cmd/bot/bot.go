package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tysion/spotter/internal/logger"
	"github.com/tysion/spotter/internal/model"
)

var AMENITY_CAFE = "cafe"

func main() {
	logger.Setup()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal().Msg("TELEGRAM_BOT_TOKEN not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to initialize tgbotapi")
	}

	log.Info().
		Str("username", bot.Self.UserName).
		Msg("Authorized on account")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			switch {
			case update.Message.Text == "/start":
				sendAmenityButtons(bot, update.Message.Chat.ID)
			case update.Message.Text == "📍 Отправить локацию":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, отправьте вашу геолокацию.")
				bot.Send(msg)
			case update.Message.Location != nil:
				sendNearbyPOIs(bot, update.Message.Chat.ID, update.Message.Location.Latitude, update.Message.Location.Longitude)
			}
		} else if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		}
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	switch cq.Data {
	case "amenity_cafe":
		msg := tgbotapi.NewMessage(cq.Message.Chat.ID, "Вы выбрали ☕ Кафе.\nТеперь отправьте свою геолокацию.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("📍 Отправить локацию"),
			),
		)

		edit := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tgbotapi.InlineKeyboardMarkup{})
		bot.Send(edit)

		bot.Send(msg)
	}
}

func sendAmenityButtons(bot *tgbotapi.BotAPI, chatID int64) {
	cafeButton := tgbotapi.NewInlineKeyboardButtonData("☕ Кафе", "amenity_cafe")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(cafeButton),
	)

	msg := tgbotapi.NewMessage(chatID, "Выберите категорию POI:")
	msg.ReplyMarkup = keyboard

	bot.Send(msg)
}

func sendNearbyPOIs(bot *tgbotapi.BotAPI, chatId int64, lat float64, lon float64) {
	lonStr := strconv.FormatFloat(lon, 'f', 6, 64)
	latStr := strconv.FormatFloat(lat, 'f', 6, 64)

	errMsg := tgbotapi.NewMessage(chatId, "⚠️ Произошла ошибка. Попробуйте, пожалуйста, позже")

	bot.Send(tgbotapi.NewMessage(chatId, "🔍 Ищу рядом с вами..."))

	url := "http://localhost:8080/poi?lat=" + latStr + "&lon=" + lonStr

	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch POIs")
		bot.Send(errMsg)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status", resp.StatusCode).Msg("Unexpected status from POI API")
		bot.Send(errMsg)
		return
	}

	var pois []model.POI
	if err := json.NewDecoder(resp.Body).Decode(&pois); err != nil {
		log.Error().Err(err).Msg("Failed to parse POIs response")
		bot.Send(errMsg)
		return
	}

	if len(pois) == 0 {
		bot.Send(tgbotapi.NewMessage(chatId, "Поблизости ничего не найдено 😕"))
	}

	for _, poi := range pois {
		text := fmt.Sprintf("📍 [%s](https://maps.google.com/?q=%.5f,%.5f)\n🔖 `%s`",
			poi.Name, poi.Lat, poi.Lon, poi.Amenity)

		msg := tgbotapi.NewMessage(chatId, text)
		msg.ParseMode = "Markdown"

		bot.Send(msg)
	}

	sendAmenityButtons(bot, chatId)
}
