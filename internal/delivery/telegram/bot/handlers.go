package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"

	"m1pes/internal/logging"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"m1pes/internal/models"
)

type (
	StockService interface {
		GetCoinFromStockExchange(ctx context.Context, coinReq models.GetCoinRequest, apiKey, secretKey string) (models.GetCoinResponse, error)
		DeleteCoin(ctx context.Context, coinTag string, userId int64) error
		GetCoinList(ctx context.Context, userId int64) ([]models.Coin, error)
		ExistCoin(ctx context.Context, coinTag string) (bool, error)
		AddCoin(coin models.Coin) error
		InsertIncome(userID int64, coinTag string, income, count float64) error
		GetCoiniks(ctx context.Context, coinTag string) (models.Coiniks, error)
		EditBuy(ctx context.Context, userId int64, buy bool) error
		CreateOrder(apiKey, apiSecret string, order models.OrderCreate) (string, error)
		GetUserWalletBalance(ctx context.Context, apiKey, apiSecret string) (float64, error)
		GetApiKeyPermissions(ctx context.Context, apiKey, apiSecret string) (models.GetApiKeyPermissionsResponse, error)
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

	// Starting trading for all users which have tradingActivated as true.
	for _, user := range users {
		if !user.TradingActivated {
			continue
		}

		update := &tgbotapi.Update{
			Message: &tgbotapi.Message{
				From: &tgbotapi.User{
					ID: user.Id,
				},
			},
		}

		h.StartTrading(ctx, b, update)
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

func (h *Handler) StopBuy(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := update.Message.From.ID

	ctx = logging.WithUserId(ctx, userId)

	err := h.ss.EditBuy(ctx, userId, false)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewMessage(userId, "Торговля на монетах остановятся сразу как продадутся")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) StartBuy(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := update.Message.From.ID

	ctx = logging.WithUserId(ctx, userId)

	err := h.ss.EditBuy(ctx, userId, true)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewMessage(userId, "Торговля на монетах возобновилась")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) StartTrading(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	userId := update.Message.From.ID

	ctx = logging.WithUserId(ctx, userId)
	err := h.ss.EditBuy(ctx, userId, true)
	if err != nil {
		log.Println(err)
	}

	user, err := h.us.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetUser", err)
	}

	// If this handler was triggered by message from user.
	if update.Message.Text != "" {
		// Check if trading already has been started.
		if user.TradingActivated {
			botMsg := tgbotapi.NewMessage(userId, "Вы уже начали торговлю!)")
			_, err = b.Send(botMsg)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
			}
			return
		}
		// Check over.

		// Getting user's api and secret keys.
		if user.ApiKey == "" || user.SecretKey == "" {
			botMsg := tgbotapi.NewMessage(userId, "У вас отсутсвуют api ключи, чтобы добавить их - /changeKeys  Введите ваш api и secret ключи через пробел.\nВАЖНО: у api ключа обязательно должны быть разрешения на: запись и чтение, торговлю на спотовом рынку и вывод средств для сбора комиссии.")
			_, err = b.Send(botMsg)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
			}

			return
		}
	}

	if _, ok := h.actionChanMap[userId]; !ok {
		h.actionChanMap[userId] = make(chan models.Message)
	}

	// This goroutine waits for action from algorithm.
	go func() {
		// This function needs for catching panics.
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "Recovered in goroutine in Handler.StartTrading", slog.String("stacktrace", string(debug.Stack())), r)

				botMsg := tgbotapi.NewMessage(ReportErrorChatId, fmt.Sprintf("stacktrace: %s\n\npanic: %v", string(debug.Stack()), r))
				_, err := b.Send(botMsg)
				if err != nil {
					slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
				}
			}
		}()

		for {
			funcUser := user

			msg := <-h.actionChanMap[funcUser.Id]

			var text string
			var chatId int64
			coiniks, err := h.ss.GetCoiniks(ctx, msg.Coin.Name)
			if err != nil {
				slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in GetCoiniks", err)
				msg.Action = err.Error()
			}

			switch msg.Action {
			case SellAction:
				a := trimTrailingZeros(fmt.Sprintf("%f", msg.Coin.CurrentPrice))
				d := trimTrailingZeros(fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", msg.Coin.Count))
				c := trimTrailingZeros(fmt.Sprintf("%.5f", msg.Coin.Income))

				def := fmt.Sprintf("Монета: %s\nПо цене: %s\nКол-во: %s\nВы заработали: %s 💲", msg.Coin.Name, a, d, c)

				text = "ПРОДАЖА\n" + def
				chatId = msg.User.Id
			case BuyAction:
				a := trimTrailingZeros(fmt.Sprintf("%f", msg.Coin.Buy[len(msg.Coin.Buy)-1]))
				d := trimTrailingZeros(fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", msg.Coin.Count/float64(len(msg.Coin.Buy))))
				def := fmt.Sprintf("Монета: %s\nПо цене: %s 💲\nКол-во: %s", msg.Coin.Name, a, d)

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
	}()

	err = h.as.StartTrading(ctx, update.Message.From.ID, h.actionChanMap)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in StartTrading", err)
	}
	msg := tgbotapi.NewMessage(update.Message.From.ID, "Ты начал торговлю, бот пришлет сообщение если купит или продаст монеты!")
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
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно остановили торговлю на аккаунте!")
	_, err = b.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) ChangeApiAndSecretKeyCmd(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	botMsg := tgbotapi.NewMessage(update.Message.From.ID, "Введите ваш api и secret ключи через пробел.\nВАЖНО: у api ключа обязательно должны быть разрешения на: запись и чтение, торговлю на спотовом рынке и вывод средств для сбора комиссии.")
	_, err := b.Send(botMsg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in SendMessage", err)
	}

	updateUser := models.NewUser(update.Message.From.ID)
	updateUser.Status = "changeKeys"

	err = h.us.UpdateUser(ctx, updateUser)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in UpdateUser", err)
	}
}

func (h *Handler) ChangeApiAndSecretKey(ctx context.Context, b *tgbotapi.BotAPI, update *tgbotapi.Update) {
	ctx = logging.WithUserId(ctx, update.Message.Chat.ID)

	keys := strings.Split(update.Message.Text, " ")

	perm, err := h.ss.GetApiKeyPermissions(ctx, keys[0], keys[1])
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in NewApiAndSecretKey", err)
	}

	// Checking permissions.

	var isAllowed int
	for _, permission := range perm.Result.Permissions.Wallet {
		if permission == "Withdraw" {
			isAllowed++
			break
		}
	}

	for _, permission := range perm.Result.Permissions.Spot {
		if permission == "SpotTrade" {
			isAllowed++
			break
		}
	}

	if isAllowed != 2 || perm.Result.ReadOnly != 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "В указанном api ключе отсутствуют некоторые разрешения.")
		_, err = b.Send(msg)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in sending message", err)
		}

		return
	}

	// If everything ok:

	updateUser := models.NewUser(update.Message.From.ID)
	updateUser.ApiKey = keys[0]
	updateUser.SecretKey = keys[1]

	err = h.us.UpdateUser(ctx, updateUser)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in UpdateUser", err)
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно изменили свои ключи ;)")
	_, err = b.Send(msg)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in sending message", err)
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

	// Deleting coin from trading.
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

	// Deleting coin from db.
	err = h.ss.DeleteCoin(ctx, update.Message.Text, user.Id)
	if err != nil {
		slog.ErrorContext(ctx, "Error delete coin", err)
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

		getCoinResp, err := h.ss.GetCoinFromStockExchange(ctx, getCoinReqParams, user.ApiKey, user.SecretKey)
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
	case "changeKeys":
		h.ChangeApiAndSecretKey(ctx, b, update)

		updateUser := models.NewUser(update.Message.From.ID)
		updateUser.Status = "none"

		err = h.us.UpdateUser(ctx, updateUser)
		if err != nil {
			slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in UpdateUser", err)
		}
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такой команды нет")
		_, err := b.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func trimTrailingZeros(numStr string) string {
	f, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return numStr
	}
	// Преобразуем число обратно в строку
	trimmedStr := strconv.FormatFloat(f, 'f', -1, 64)

	// Если строка содержит точку, удалим все нули, стоящие в конце
	if strings.Contains(trimmedStr, ".") {
		trimmedStr = strings.TrimRight(trimmedStr, "0")
		// Если точка осталась в конце строки, удалим её тоже
		trimmedStr = strings.TrimRight(trimmedStr, ".")
	}

	return trimmedStr
}
