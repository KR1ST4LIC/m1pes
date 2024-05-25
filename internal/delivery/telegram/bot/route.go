package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

func (h *Handler) Route(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message != nil {
		parts := strings.Split(update.Message.Command(), "_")

		switch parts[0] {
		case "start":
			h.Start(ctx, b, update)
		case "coin":
			h.GetCoinList(ctx, b, update)
		case "percent":
			h.GetNewPercent(ctx, b, update)
		case "replenish":
			amount, _ := strconv.ParseInt(parts[1], 10, 64)
			ctx = context.WithValue(ctx, "replenishAmount", amount)
			h.ReplenishBalance(ctx, b, update)
			h.GetCoinList(b, update)
		case "addCoin":
			h.GetNewCoin(b, update)
		default:
			h.UnknownCommand(b, update)
		}
	}
}
