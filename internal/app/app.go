package app

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"m1pes/internal/config"
	handler "m1pes/internal/delivery/telegram/bot"
	"m1pes/internal/logging"
	"m1pes/internal/repository/api/stocks/bybit"
	stockPostgres "m1pes/internal/repository/storage/stocks/postgres"
	userPostgres "m1pes/internal/repository/storage/user/postgres"
	"m1pes/internal/service/algorithm"
	"m1pes/internal/service/stocks"
	"m1pes/internal/service/user"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	bot *tgbotapi.BotAPI
	cfg *config.Config
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
	// Stock dependencies.
	storageStock := stockPostgres.New(a.cfg.DBConn)
	apiStock := bybit.New()
	stockService := stocks.New(apiStock, storageStock)

	// User dependencies.
	storageUser := userPostgres.New(a.cfg.DBConn)
	userService := user.New(storageUser)

	// Algorithm dependencies.
	algoService := algorithm.New(apiStock, storageStock, storageUser)

	// Init handler.
	h := handler.New(stockService, userService, algoService, a.bot)

	go func() {
		if err := a.RunTelegramBot(ctx, h); err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in ParsingPrice", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	slog.Info("Shutting down app...")

	storageUser.Conn.Close()
	storageStock.Conn.Close()

	return nil
}
