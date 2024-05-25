package bot

import (
	"context"
	"log"
	"log/slog"
	"m1pes/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoinList(ctx context.Context, userId int64) ([]string, error)
		ExistCoin(ctx context.Context, coinTag string) (bool, error)
	}

	UserService interface {
		NewUser(ctx context.Context, user models.User) error
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

func (h *Handler) GetNewPercent(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "На скольки процентах вы хотите торговать?")
	_, err := b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}
	// записать в дб статус /create percent
}

func (h *Handler) GetNewCoin(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	list, err := h.ss.GetCoinList(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
	}
	if len(list) < 5 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Скиньте тег койна, например: BTCUSDT или ETHUSDT")
		_, err = b.Send(msg)
		if err != nil {
			log.Println(err)
		}
		// status v db /createCoin
	} else {
		//h.ss.ExistCoin(coinTag) check est li moneta
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас уже 5 монет, если хотите добавить новую - удалить старую /delete")
		_, err = b.Send(msg)
		if err != nil {
			log.Println(err)
		}

	}
}
