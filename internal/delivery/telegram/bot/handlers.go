package bot

import (
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoinList(userId int64) ([]string, error)
		AddCoin(userId int64, coinTag string) error
		ExistCoin(coinTag string) (bool, error)
		CheckStatus(userId int64) (string, error)
		UpdateStatus(userID int64, status string) error
		UpdatePercent(userID int64, percent float64) error
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
	err := h.ss.UpdateStatus(update.Message.From.ID, "updatePercent")
	if err != nil {
		log.Println(err)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На скольки процентах вы хотите торговать?")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) GetNewCoin(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	list, err := h.ss.GetCoinList(update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
	}
	if len(list) < 5 {
		err = h.ss.UpdateStatus(update.Message.From.ID, "addCoin")
		if err != nil {
			log.Println(err)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Скиньте тег койна, например: BTCUSDT или ETHUSDT")
		_, err = b.Send(msg)
		if err != nil {
			log.Println(err)
		}
		// status v db /createCoin
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас уже 5 монет, если хотите добавить новую - удалить старую /deleteCoin") // сделать койн дилит
		_, err := b.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func (h *Handler) UnknownCommand(b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	var status string
	status, err := h.ss.CheckStatus(update.Message.From.ID)
	if err != nil {
		log.Println(err)
	}
	switch status {
	case "updatePercent":
		text := strings.Replace(update.Message.Text, ",", ".", -1)
		percent, err := strconv.ParseFloat(text, 64)
		if err != nil {
			log.Println(err)
		}
		if percent >= 0.25 && percent <= 20 {
			err = h.ss.UpdatePercent(update.Message.From.ID, percent)
			if err != nil {
				log.Println(err)
			}
			err = h.ss.UpdateStatus(update.Message.From.ID, "none")
			if err != nil {
				log.Println(err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Процент торговли успешно изменен")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else {
			err = h.ss.UpdateStatus(update.Message.From.ID, "none")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неправильно введены проценты. Максимальное значение процентов - 20, а минимальное - 0.25 попробуйте ещё раз - /percent")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	case "addCoin":
		can, err := h.ss.ExistCoin(update.Message.Text)
		if err != nil {
			log.Println(err)
		}
		if can {
			err = h.ss.AddCoin(update.Message.From.ID, update.Message.Text)
			if err != nil {
				log.Println(err)
			}
			err = h.ss.UpdateStatus(update.Message.From.ID, "none")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Монета успешно добавленна")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else {
			err = h.ss.UpdateStatus(update.Message.From.ID, "none")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой монеты не существует попробуйте ещё раз - /addCoin")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой команды нет")
		_, err := b.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
