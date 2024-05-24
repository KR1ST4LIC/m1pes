package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type (
	StockService interface {
		GetCoinList() ([]string, error)
	}
)

type Handler struct {
	ss StockService
}

func New(ss StockService) *Handler {
	return &Handler{ss: ss}
}

func (h *Handler) Start(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "hi there!")
	_, err := b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) GetCoinList(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	list, err := h.ss.GetCoinList()
	if err != nil {
		log.Println(err)
	}
	var text string
	for i := range list {
		text += list[i] + " "
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
