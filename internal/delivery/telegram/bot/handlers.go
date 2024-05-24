package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoinList(userId int64) ([]string, error)
		ExistCoin(coinTag string) (bool, error)
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
	text = "Ваши монеты: "
	for i := range list {
		text += list[i] + " "
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) GetNewPercent(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На скольки процентах вы хотите торговать?")
	_, err := b.Send(msg)
	if err != nil {
		log.Println(err)
	}
	// записать в дб статус /create percent
}

func (h *Handler) GetNewCoin(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	list, err := h.ss.GetCoinList(update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
	}
	if len(list) < 5 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Скиньте тег койна, например: BTCUSDT или ETHUSDT")
		_, err := b.Send(msg)
		if err != nil {
			log.Println(err)
		}
		// status v db /createCoin
	} else {
		//h.ss.ExistCoin(coinTag) check est li moneta
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас уже 5 монет, если хотите добавить новую - удалить старую /delete")
		_, err := b.Send(msg)
		if err != nil {
			log.Println(err)
		}

	}
}
