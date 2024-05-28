package bot

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Route(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message != nil && update.Message.Chat.IsPrivate() {
		parts := strings.Split(update.Message.Command(), "_")

		switch parts[0] {
		case "start":
			h.Start(ctx, b, update)
		case "coin":
			h.GetCoinList(ctx, b, update)
		case "percent":
			h.GetNewPercent(ctx, b, update)
		case "replenish":
			if len(parts) != 2 {
				return
			}
			amount, _ := strconv.ParseInt(parts[1], 10, 64)
			ctx = context.WithValue(ctx, "replenishAmount", amount)
			h.ReplenishBalance(ctx, b, update)
		case "starttrading":
			h.StartTrading(ctx, b, update)
		case "delete":
			h.DeleteCoin(ctx, b, update)
		case "addCoin":
			h.GetNewCoin(ctx, b, update)
		default:
			h.UnknownCommand(ctx, b, update)
		}
	}
}
