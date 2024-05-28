package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"m1pes/internal/algorithm"

	"m1pes/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoin(ctx context.Context, userId int64, coin string)
		GetCoinList(ctx context.Context, userId int64) (models.List, error)
		ExistCoin(ctx context.Context, coinTag string) (bool, error)
		AddCoin(coin models.Coin) error
		InsertIncome(userID int64, coinTag string, income, count float64) error
	}

	UserService interface {
		UpdateUser(ctx context.Context, user models.User) error
		NewUser(ctx context.Context, user models.User) error
		GetUser(ctx context.Context, userId int64) (models.User, error)
		ReplenishBalance(ctx context.Context, userId int64, amount float64) error
	}

	AlgorithmService interface {
		StartTrading(ctx context.Context, userId int64, actionChanMap map[int64]chan models.Message) error
		DeleteCoin(ctx context.Context, userId int64, coin string) error
	}
)

type Handler struct {
	as            AlgorithmService
	ss            StockService
	us            UserService
	actionChanMap map[int64]chan models.Message
}

func New(ss StockService, us UserService, as AlgorithmService) *Handler {
	return &Handler{ss: ss, us: us, as: as, actionChanMap: make(map[int64]chan models.Message)}
}

func (h *Handler) Start(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	user := models.NewUser(update.Message.From.ID)
	err := h.us.NewUser(ctx, user)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in NewUser", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, это бот для торговли на биржах!")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) StartTrading(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.From.ID)

	user, err := h.us.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetUser", err)
	}
	user.UpdateUserId(update.Message.From.ID)

	if _, ok := h.actionChanMap[update.Message.Chat.ID]; !ok {
		h.actionChanMap[update.Message.Chat.ID] = make(chan models.Message)
	}

	// this goroutine waits for action from algorithm
	go func() {
		for {
			select {
			case msg := <-h.actionChanMap[update.Message.From.ID]:

				var text string
				switch msg.Action {
				case algorithm.SellAction:
					err = h.ss.InsertIncome(msg.User.Id, msg.Coin.Name, msg.Coin.Income, msg.Coin.Count)
					if err != nil {
						slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in InsertIncome", err)
					}
					err = h.us.ReplenishBalance(ctx, msg.User.Id, msg.Coin.Income)
					def := fmt.Sprintf("Монета: %s\nПо цене: %.4f\nКол-во: %.4f\nВы заработали: %.4f 💲", msg.Coin.Name, msg.Coin.CurrentPrice, msg.Coin.Count, msg.Coin.Income)
					text = "ПРОДАЖА\n" + def
				case algorithm.BuyAction:
					def := fmt.Sprintf("Монета: %s\nПо цене: %.4f 💲\nКол-во: %.4f", msg.Coin.Name, msg.Coin.Buy[len(msg.Coin.Buy)-1], msg.Coin.Count/float64(len(msg.Coin.Buy)))
					text = "ПОКУПКА\n" + def
				}
				botMsg := tgbotapi.NewMessage(msg.User.Id, text)
				_, err = b.Send(botMsg)
				if err != nil {
					slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
				}
			}
		}
	}()

	err = h.as.StartTrading(ctx, update.Message.From.ID, h.actionChanMap)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StartTrading", err)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ты начал торговлю, бот пришлет сообщение если купит или продаст монеты!")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) DeleteCoin(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	user := models.NewUser(update.Message.From.ID)
	user.Status = "deleteCoin"

	err := h.us.UpdateUser(ctx, user)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in UpdateStatus", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите название монеты, которой вы хотите удалить")
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
	text = "Ваши монеты:\n"
	for i := 0; i < len(list.Buy); i++ {
		var sum float64 = 0
		for d := 0; d < len(list.Buy[list.Name[i]]); d++ {
			sum += list.Buy[list.Name[i]][d]
		}
		if len(list.Buy[list.Name[i]]) != 0 {
			avg := sum / float64(len(list.Buy[list.Name[i]]))
			text += fmt.Sprintf("%s  купленно на: %.3f💲\n", list.Name[i], list.Count[i]*avg)
		} else {
			text += fmt.Sprintf("%s  купленно на: 0💲\n", list.Name[i])
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err = b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}
}

func (h *Handler) ReplenishBalance(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	err := h.us.ReplenishBalance(ctx, update.Message.Chat.ID, ctx.Value("replenishAmount").(float64))
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in ReplenishBalance", err)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "баланс успешно добавлен")
	_, err = b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}
}

func (h *Handler) GetNewPercent(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	user := models.NewUser(update.Message.From.ID)
	user.Status = "updatePercent"

	err := h.us.UpdateUser(ctx, user)
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
	if len(list.Buy) < 5 {
		user := models.NewUser(update.Message.From.ID)
		user.Status = "addCoin"

		err = h.us.UpdateUser(ctx, user)
		if err != nil {
			log.Println(err)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Скиньте тег койна, например: BTCUSDT или ETHUSDT")
		_, err = b.Send(msg)
		if err != nil {
			log.Println(err)
		}
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас уже 5 монет, если хотите добавить новую - удалить старую /delete")
		_, err = b.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func (h *Handler) UnknownCommand(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	user, err := h.us.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(user)
	switch user.Status {
	case "updatePercent":
		text := strings.Replace(update.Message.Text, ",", ".", -1)
		percent, err := strconv.ParseFloat(text, 64)
		if err != nil {
			log.Println(err)
		}
		if percent >= 0.25 && percent <= 20 {
			user = models.NewUser(update.Message.From.ID)
			user.Percent = percent * 0.01

			err = h.us.UpdateUser(ctx, user)
			if err != nil {
				log.Println(err)
			}

			user = models.NewUser(update.Message.From.ID)
			user.Status = "none"

			err = h.us.UpdateUser(ctx, user)
			if err != nil {
				log.Println(err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Процент торговли успешно изменен")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else {
			user := models.NewUser(update.Message.From.ID)
			user.Status = "none"

			err = h.us.UpdateUser(ctx, user)
			if err != nil {
				log.Println(err)
			}

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
			coin := models.Coin{Name: update.Message.Text, UserId: update.Message.From.ID}

			err = h.ss.AddCoin(coin)
			if err != nil {
				log.Println(err)
			}

			user := models.NewUser(update.Message.From.ID)
			user.Status = "none"

			err = h.us.UpdateUser(ctx, user)
			if err != nil {
				log.Println(err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Монета успешно добавленна")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else {
			user := models.NewUser(update.Message.From.ID)
			user.Status = "none"

			err = h.us.UpdateUser(ctx, user)
			if err != nil {
				log.Println(err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой монеты не существует попробуйте ещё раз - /addCoin")
			_, err := b.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	case "deleteCoin":
		err = h.as.DeleteCoin(ctx, update.Message.From.ID, update.Message.Text)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StopTrading", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ты не торгуешь на этой монете, чтобы посмотреть список монет - /coin")
			_, err = b.Send(msg)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in sending message", err)
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ты удалил эту монету!")
			_, err = b.Send(msg)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in sending message", err)
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
