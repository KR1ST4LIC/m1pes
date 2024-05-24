package app

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	userHandler "m1pes/internal/delivery/telegram/bot"
	"m1pes/internal/repository/api/stocks/bybit"
	"m1pes/internal/repository/storage/stocks/postgres"
	"m1pes/internal/service/stocks"
)

type App struct {
	bot *tgbotapi.BotAPI
	cfg *Config
}

func New() *App {
	app := new(App)
	app.InitDeps()
	return app
}

func (a *App) Start() {
	storageStock := postgres.New()
	apiStock := bybit.New()

	stockService := stocks.New(apiStock, storageStock)

	handler := userHandler.New(stockService)

	if err := a.RunTelegramBot(handler); err != nil {
		log.Fatal(err)
	}
}
