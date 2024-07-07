package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"

	"m1pes/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type fgh func() error

type (
	StockService interface {
		GetCoin(ctx context.Context, coinReq models.GetCoinRequest, apiKey, secretKey string) (models.GetCoinResponse, error)
		GetCoinList(ctx context.Context, userId int64) ([]models.Coin, error)
		ExistCoin(ctx context.Context, coinTag string) (bool, error)
		AddCoin(coin models.Coin) error
		InsertIncome(userID int64, coinTag string, income, count float64) error
		GetCoiniks(ctx context.Context, coinTag string) (models.Coiniks, error)
		CreateOrder(apiKey, apiSecret string, order models.OrderCreate) (string, error)
		GetUserWalletBalance(ctx context.Context, apiKey, apiSecret string) (float64, error)
	}

	UserService interface {
		UpdateUser(ctx context.Context, user models.User) error
		GetAllUsers(ctx context.Context) ([]models.User, error)
		NewUser(ctx context.Context, user models.User) error
		GetUser(ctx context.Context, userId int64) (models.User, error)
		GetIncomeLastDay(ctx context.Context, userID int64) (float64, error)
	}

	AlgorithmService interface {
		StartTrading(ctx context.Context, userId int64, actionChanMap map[int64]chan models.Message) error
		StopTrading(ctx context.Context, userID int64) error
		DeleteCoin(ctx context.Context, userId int64, coin string) error
	}
)

type Handler struct {
	as            AlgorithmService
	ss            StockService
	us            UserService
	actionChanMap map[int64]chan models.Message
}

const (
	SellAction = "sell"
	BuyAction  = "buy"

	ReportErrorChatId = -4216803774 // TG id of chat where bot sends errors.
)

func New(ss StockService, us UserService, as AlgorithmService, b *tgbotapi.BotAPI) *Handler {
	ctx := context.Background()

	h := &Handler{ss: ss, us: us, as: as, actionChanMap: make(map[int64]chan models.Message)}

	users, err := h.us.GetAllUsers(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	for _, user := range users {
		if user.TradingActivated {
			ctx = logging.WithUserId(ctx, user.Id)

			if _, ok := h.actionChanMap[user.Id]; !ok {
				h.actionChanMap[user.Id] = make(chan models.Message)
			}

			// This goroutine waits for action from algorithm.
			go func() {
				funcUser := user
				for {
					select {
					case msg := <-h.actionChanMap[funcUser.Id]:
						var text string
						var chatId int64

						switch msg.Action {
						case SellAction:
							def := fmt.Sprintf("Монета: %s\nПо цене: %.4f\nКол-во: %.4f\nВы заработали: %.4f 💲", msg.Coin.Name, msg.Coin.CurrentPrice, msg.Coin.Count, msg.Coin.Income)

							text = "ПРОДАЖА\n" + def
							chatId = msg.User.Id
						case BuyAction:
							def := fmt.Sprintf("Монета: %s\nПо цене: %.4f 💲\nКол-во: %.4f", msg.Coin.Name, msg.Coin.Buy[len(msg.Coin.Buy)-1], msg.Coin.Count/float64(len(msg.Coin.Buy)))

							text = "ПОКУПКА\n" + def
							chatId = msg.User.Id
						default:
							text = fmt.Sprintf("Ошибка: %s \nfile: %s line: %d", msg.Action, msg.File, msg.Line)
							chatId = ReportErrorChatId
						}
						botMsg := tgbotapi.NewMessage(chatId, text)
						_, err = b.Send(botMsg)
						if err != nil {
							slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
						}
					}
				}
			}()

			err = h.as.StartTrading(ctx, user.Id, h.actionChanMap)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StartTrading", err)
			}
			msg := tgbotapi.NewMessage(user.Id, "Ты начал торговлю, бот пришлет сообщение если купит или продаст монеты!")
			_, err = b.Send(msg)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
			}
		}
	}

	return h
}

// Handlers with Cmd at the end need for setting new status for user, after them should be normal handler.

func (h *Handler) Start(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := update.Message.From.ID

	ctx = logging.WithUserId(ctx, userId)

	user := models.NewUser(userId)
	err := h.us.NewUser(ctx, user)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in NewUser", err)
	}

	msg := tgbotapi.NewMessage(userId, "Привет, это бот для торговли на биржах!")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) StartTrading(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := update.Message.From.ID

	ctx = logging.WithUserId(ctx, userId)

	// Check if trading already has been started.
	user, err := h.us.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetUser", err)
	}

	if user.TradingActivated {
		botMsg := tgbotapi.NewMessage(userId, "Вы уже начали торговлю!)")
		_, err := b.Send(botMsg)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
		}

		return
	}
	// Check over.

	if _, ok := h.actionChanMap[userId]; !ok {
		h.actionChanMap[userId] = make(chan models.Message)
	}

	// This goroutine waits for action from algorithm.
	go func() {
		for {
			select {
			case msg := <-h.actionChanMap[update.Message.From.ID]:
				var text string
				var chatId int64

				switch msg.Action {
				case SellAction:
					def := fmt.Sprintf("Монета: %s\nПо цене: %.4f\nКол-во: %.4f\nВы заработали: %.4f 💲", msg.Coin.Name, msg.Coin.CurrentPrice, msg.Coin.Count, msg.Coin.Income)

					text = "ПРОДАЖА\n" + def
					chatId = msg.User.Id
				case BuyAction:
					def := fmt.Sprintf("Монета: %s\nПо цене: %.4f 💲\nКол-во: %.4f", msg.Coin.Name, msg.Coin.Buy[len(msg.Coin.Buy)-1], msg.Coin.Count/float64(len(msg.Coin.Buy)))

					text = "ПОКУПКА\n" + def
					chatId = msg.User.Id
				default:
					text = fmt.Sprintf("Ошибка: %s \nfile: %s line: %d", msg.Action, msg.File, msg.Line)
					chatId = ReportErrorChatId
				}
				botMsg := tgbotapi.NewMessage(chatId, text)
				_, err := b.Send(botMsg)
				if err != nil {
					slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
				}
			}
		}

		//for {
		//	msg := <-h.actionChanMap[update.Message.From.ID]
		//
		//	var text string
		//	var chatId int64
		//
		//	switch msg.Action {
		//	case SellAction:
		//		err := h.ss.InsertIncome(msg.User.Id, msg.Coin.Name, msg.Coin.Income, msg.Coin.Count)
		//		if err != nil {
		//			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in InsertIncome", err)
		//		}
		//
		//		err = h.us.ReplenishBalance(ctx, msg.User.Id, msg.Coin.Income)
		//		if err != nil {
		//			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in ReplenishBalance", err)
		//		}
		//
		//		def := fmt.Sprintf("Монета: %s\nПо цене: %.4f\nКол-во: %.4f\nВы заработали: %.4f 💲", msg.Coin.Name, msg.Coin.CurrentPrice, msg.Coin.Count, msg.Coin.Income)
		//
		//		text = "ПРОДАЖА\n" + def
		//		chatId = msg.User.Id
		//	case BuyAction:
		//		def := fmt.Sprintf("Монета: %s\nПо цене: %.4f 💲\nКол-во: %.4f", msg.Coin.Name, msg.Coin.Buy[len(msg.Coin.Buy)-1], msg.Coin.Count/float64(len(msg.Coin.Buy)))
		//
		//		text = "ПОКУПКА\n" + def
		//		chatId = msg.User.Id
		//	default:
		//		text = msg.Action
		//		chatId = ReportErrorChatId
		//	}
		//	botMsg := tgbotapi.NewMessage(chatId, text)
		//	_, err := b.Send(botMsg)
		//	if err != nil {
		//		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
		//	}
		//}
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

func (h *Handler) StopTrading(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	err := h.as.StopTrading(ctx, update.Message.From.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StopTrading", err)
	}
}

func (h *Handler) DeleteCoinCmd(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
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

func (h *Handler) DeleteCoin(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	updateData, err := h.us.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetUser", err)
	}

	user := models.NewUser(update.Message.From.ID)
	user.Status = "none"
	user.TradingActivated = updateData.TradingActivated

	err = h.us.UpdateUser(ctx, user)
	if err != nil {
		log.Println(err)
	}

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
}

func (h *Handler) GetCoinList(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)
	user, err := h.us.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetUser", err)
	}
	if user.ApiKey == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "у тебя нет apiKey обратитесь к @n1fawin")
		_, err = b.Send(msg)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
		}
		return
	}
	if user.SecretKey == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "у тебя нет secretKey обратитесь к @n1fawin")
		_, err = b.Send(msg)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
		}
		return
	}

	list, err := h.ss.GetCoinList(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StockService.GetCoinList", err)
	}

	text := "Ваши монеты:\n"

	bal, err := h.ss.GetUserWalletBalance(ctx, user.ApiKey, user.SecretKey)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in Get Balance Frim Bybit", err)
	}

	user.USDTBalance = bal

	err = h.us.UpdateUser(ctx, user)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in update user", err)
	}

	income, err := h.us.GetIncomeLastDay(ctx, user.Id)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in Get income", err)
	}

	var userSum float64
	for i := 0; i < len(list); i++ {
		var coinSum float64
		for d := 0; d < len(list[i].Buy); d++ {
			coinSum += list[i].Buy[d]
		}

		if len(list[i].Buy) != 0 {
			avg := coinSum / float64(len(list[i].Buy))

			userSum += list[i].Count * avg

			text += fmt.Sprintf("%s  купленно на: %.3f💲\n", list[i].Name, list[i].Count*avg)
		} else {
			text += fmt.Sprintf("%s  купленно на: 0💲\n", list[i].Name)
		}
	}

	text += fmt.Sprintf("\nСумарный закуп: %.3f", userSum)

	text += fmt.Sprintf("\nОбщий баланс: %.4f", user.USDTBalance)

	text += fmt.Sprintf("\nЗаработал в процентах за последний день: %.3f", income/user.USDTBalance*100) + "%"

	text += fmt.Sprintf("\nИспользуется баланса: %.3f", userSum/user.USDTBalance*100) + "%"

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	_, err = b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}
}

func (h *Handler) AddCoinCmd(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	list, err := h.ss.GetCoinList(ctx, update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
	}

	if len(list) < 5 {
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

func (h *Handler) AddCoin(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)
	user, err := h.us.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetUser", err)
	}
	if user.ApiKey == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "у тебя нет apiKey обратитесь к @n1fawin")
		_, err = b.Send(msg)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
		}
		return
	}
	if user.SecretKey == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "у тебя нет secretKey обратитесь к @n1fawin")
		_, err = b.Send(msg)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
		}
		return
	}
	can, err := h.ss.ExistCoin(ctx, update.Message.Text)
	if err != nil {
		log.Println(err)
	}
	if can {
		coiniks, err := h.ss.GetCoiniks(ctx, update.Message.Text)
		if err != nil {
			slog.ErrorContext(ctx, "Error getting coiniks", err)
			return
		}

		balance, err := h.ss.GetUserWalletBalance(ctx, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in Get Balance Frim Bybit", err)
		}

		getCoinReqParams := make(models.GetCoinRequest)
		getCoinReqParams["category"] = "spot"
		getCoinReqParams["symbol"] = update.Message.Text

		getCoinResp, err := h.ss.GetCoin(ctx, getCoinReqParams, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error getting coin from algorithm", err)
			return
		}

		currentPrice, err := strconv.ParseFloat(getCoinResp.Result.List[0].Price, 64)
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing current price to float", err)
			return
		}

		if balance*0.015/currentPrice < coiniks.MinSumBuy*1.1 {
			user := models.NewUser(update.Message.From.ID)
			user.Status = "none"

			err = h.us.UpdateUser(ctx, user)
			if err != nil {
				log.Println(err)
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("На этой монете есть ограничение для минимального баланса на аккаунте - %.4f$ , попробуйте еще раз после пополнения баланса /addCoin", coiniks.MinSumBuy*1.1*currentPrice*67.0))
			_, err = b.Send(msg)
			if err != nil {
				log.Println(err)
			}
			return
		}
		if err != nil {
			log.Println(err)
		}

		coin := models.Coin{Name: update.Message.Text, UserId: update.Message.From.ID}
		err = h.ss.AddCoin(coin)
		if err != nil {
			log.Println(err)
		}

		user.Status = "none"

		err = h.us.UpdateUser(ctx, user)
		if err != nil {
			log.Println(err)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Монета успешно добавленна")
		_, err = b.Send(msg)
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
}

func (h *Handler) UnknownCommand(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	user, err := h.us.GetUser(ctx, update.Message.From.ID)
	if err != nil {
		log.Println(err)
	}

	switch user.Status {
	case "addCoin":
		h.AddCoin(ctx, b, update)
	case "deleteCoin":
		h.DeleteCoin(ctx, b, update)
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой команды нет")
		_, err := b.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
