package app

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	"m1pes/internal/config"
	handler "m1pes/internal/delivery/telegram/bot"
	"m1pes/internal/logging"
	"m1pes/internal/repository/api/stocks/bybit"
	stockPostgres "m1pes/internal/repository/storage/stocks/postgres"
	userPostgres "m1pes/internal/repository/storage/user/postgres"
	"m1pes/internal/service/stocks"
	"m1pes/internal/service/user"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	bot *tgbotapi.BotAPI
	cfg *config.Config
	//logger *slog.Logger
}

func New() (*App, error) {
	app := new(App)
	err := app.InitDeps()
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	// stock dependencies
	storageStock := stockPostgres.New(a.cfg.DBConn)
	apiStock := bybit.New()
	stockService := stocks.New(apiStock, storageStock)

	// user dependencies
	storageUser := userPostgres.New(a.cfg.DBConn)
	userService := user.New(storageUser)

	// init handler
	h := handler.New(stockService, userService)

	go func() {
		if err := a.RunTelegramBot(ctx, h); err != nil {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("Shutting down app...")

	err := storageUser.Conn.Close()
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "Error: ", err.Error())
	}

	err = storageStock.Conn.Close()
	if err != nil {
		log.Println(err)
	}

	return nil
}
