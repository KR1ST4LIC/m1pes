package app

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	handler "m1pes/internal/delivery/telegram/bot"
	"m1pes/internal/repository/api/stocks/bybit"
	stockPostgres "m1pes/internal/repository/storage/stocks/postgres"
	userPostgres "m1pes/internal/repository/storage/user/postgres"
	"m1pes/internal/service/stocks"
	"m1pes/internal/service/user"
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
	// stock dependencies
	storageStock := stockPostgres.New()
	apiStock := bybit.New()
	stockService := stocks.New(apiStock, storageStock)

	// user dependencies
	storageUser := userPostgres.New()
	userService := user.New(storageUser)

	h := handler.New(stockService, userService)

	if err := a.RunTelegramBot(h); err != nil {
		log.Fatal(err)
	}
}
