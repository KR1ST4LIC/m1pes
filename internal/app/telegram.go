package app

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"m1pes/internal/delivery/telegram/bot"
)

func (a *App) InitTelegramBot() error {
	b, err := tgbotapi.NewBotAPI(a.cfg.Bot.Token)
	if err != nil {
		return err
	}

	// assigning new bot to app's bot
	a.bot = b

	slog.Info("Authorized on account " + b.Self.UserName)
	return nil
}

func (a *App) RunTelegramBot(ctx context.Context, h *bot.Handler) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := a.bot.GetUpdatesChan(u)

	for update := range updates {
		h.Route(ctx, a.bot, &update)
	}
	return nil
}
