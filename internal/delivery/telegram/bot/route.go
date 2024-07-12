package bot

import (
	"context"
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
		case "startTrading":
			h.StartTrading(ctx, b, update)
		case "stopTrading":
			h.StopTrading(ctx, b, update)
		case "delete":
			h.DeleteCoinCmd(ctx, b, update)
		case "addCoin":
			h.AddCoinCmd(ctx, b, update)
		default:
			h.UnknownCommand(ctx, b, update)
		}
	}
}
