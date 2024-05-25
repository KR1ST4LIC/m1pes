package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"m1pes/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoinList(ctx context.Context, userId int64) ([]string, error)
		ExistCoin(ctx context.Context, coinTag string) (bool, error)
		StartTrading(ctx context.Context, userId int64) error
		StopTradingCoin(ctx context.Context, userId int64, coin string) error
		AddCoin(userId int64, coinTag string) error
		CheckStatus(userId int64) (string, error)
		UpdateStatus(userID int64, status string) error
		UpdatePercent(userID int64, percent float64) error
	}

	UserService interface {
		NewUser(ctx context.Context, user models.User) error
		ReplenishBalance(ctx context.Context, userId, amount int64) error
	}
)

type Handler struct {
	ss StockService
	us UserService
}

func New(ss StockService, us UserService) *Handler {
	return &Handler{ss: ss, us: us}
}

func (h *Handler) Start(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)
	slog.InfoContext(ctx, "new call of start handler!")

	user := models.NewUser(update.Message.From.ID)
	err := h.us.NewUser(ctx, user)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in NewUser", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "hi there!")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) StartTrading(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	err := h.ss.StartTrading(ctx, update.Message.From.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StartTrading", err)
	}
}

func (h *Handler) StopTrading(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	err := h.ss.StopTradingCoin(ctx, update.Message.From.ID, ctx.Value("coin").(string))
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StopTrading", err)
	}
}

func (h *Handler) GetCoinList(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	list, err := h.ss.GetCoinList(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StockService.GetCoinList", err)
	}
	var text string
	text = "Ваши монеты: "
	for i := range list {
		text += list[i] + " "
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err = b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}
}
func (h *Handler) ReplenishBalance(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	fmt.Println("herererere")
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	err := h.us.ReplenishBalance(ctx, update.Message.Chat.ID, ctx.Value("replenishAmount").(int64))
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in ReplenishBalance", err)
	}
}
func (h *Handler) GetNewPercent(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)
	err := h.ss.UpdateStatus(update.Message.From.ID, "updatePercent")
	if err != nil {
		log.Println(err)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На скольки процентах вы хотите торговать?")
	_, err = b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}
}

func (h *Handler) GetNewCoin(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	list, err := h.ss.GetCoinList(ctx, update.Message.Chat.ID)
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
	} else {
		//h.ss.ExistCoin(coinTag) check est li moneta
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас уже 5 монет, если хотите добавить новую - удалить старую /delete")
		_, err = b.Send(msg)
		if err != nil {
			log.Println(err)
		}

	}
}

func (h *Handler) UnknownCommand(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

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
		can, err := h.ss.ExistCoin(ctx, update.Message.Text)
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
