package bot

import (
	"context"
	"fmt"
	"log/slog"
	"m1pes/internal/logging"
	"runtime/debug"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Route(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// This function needs for catching panics.
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Recovered in Handler.Route", slog.String("stacktrace", string(debug.Stack())), r)

			botMsg := tgbotapi.NewMessage(ReportErrorChatId, fmt.Sprintf("stacktrace: %s\n\npanic: %v", string(debug.Stack()), r))
			_, err := b.Send(botMsg)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
			}
		}
	}()

	if update.Message != nil && update.Message.Chat.IsPrivate() {
		parts := strings.Split(update.Message.Command(), "_")

		switch parts[0] {
		case "start":
			h.Start(ctx, b, update)
		case "coin":
			h.GetCoinList(ctx, b, update)
		case "stopBuy":
			h.StopBuy(ctx, b, update)
		case "startBuy":
			h.StartBuy(ctx, b, update)
		case "startTrading":
			h.StartTrading(ctx, b, update)
		case "stopTrading":
			h.StopTrading(ctx, b, update)
		case "delete":
			h.DeleteCoinCmd(ctx, b, update)
		case "addCoin":
			h.AddCoinCmd(ctx, b, update)
		case "changeKeys":
			h.ChangeApiAndSecretKeyCmd(ctx, b, update)
		default:
			h.UnknownCommand(ctx, b, update)
		}
	}
}
