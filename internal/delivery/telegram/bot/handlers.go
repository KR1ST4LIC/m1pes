package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoinList(userId int64) ([]string, error)
	}

	UserService interface {
		NewUser(user models.User) error
	}
)

type Handler struct {
	ss StockService
	us UserService
}

func New(ss StockService, us UserService) *Handler {
	return &Handler{ss: ss, us: us}
}

func (h *Handler) Start(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	user := models.NewUser(update.Message.From.ID)
	err := h.us.NewUser(user)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "hi there!")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) GetCoinList(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	list, err := h.ss.GetCoinList(update.Message.Chat.ID)
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
