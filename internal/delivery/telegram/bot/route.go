package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Route(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message != nil {
		switch update.Message.Command() {
		case "start":
			h.Start(ctx, b, update)
		case "coin":
			h.GetCoinList(ctx, b, update)
		case "/percent":
			h.GetNewPercent(ctx, b, update)
		}
	}
}
