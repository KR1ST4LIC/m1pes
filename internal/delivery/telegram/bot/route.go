package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (h *Handler) Route(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.Message != nil {
		switch update.Message.Command() {
		case "start":
			h.Start(b, update)
		case "coin":
			h.GetCoinList(b, update)
		case "percent":
			h.GetNewPercent(b, update)
		case "addCoin":
			h.GetNewCoin(b, update)
		default:
			h.UnknownCommand(b, update)
		}
	}
}
