package algorithm

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"m1pes/internal/logging"

	"m1pes/internal/delivery/telegram/bot"
	"m1pes/internal/models"

	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
	storageUser "m1pes/internal/repository/storage/user"
)

const (
	SuccessfulOrderStatus = "Filled"
)

type Service struct {
	apiRepo      apiStock.Repository
	sStorageRepo storageStock.Repository
	uStorageRepo storageUser.Repository
	stopCoinMap  map[int64]map[string]chan struct{}
}

func New(apiRepo apiStock.Repository, sStoRepo storageStock.Repository, uStoRepo storageUser.Repository) *Service {
	return &Service{apiRepo, sStoRepo, uStoRepo, make(map[int64]map[string]chan struct{})}
}

func (s *Service) StartTrading(ctx context.Context, userId int64, actionChanMap map[int64]chan models.Message) error {
	coinList, err := s.sStorageRepo.GetCoinList(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin list from storage", err)
		return err
	}

	// In that for loop we are starting handling every user's coin.
	for _, coin := range coinList {
		ctx = logging.WithCoinTag(ctx, coin.Name)

		// init map that stores coin name as key and map2 as value
		// map2 stores userId as key and struct{} as value
		if _, ok := s.stopCoinMap[userId][coin.Name]; ok {
			continue
		}

		if _, ok := s.stopCoinMap[userId]; !ok {
			s.stopCoinMap[userId] = make(map[string]chan struct{})
		}
		s.stopCoinMap[userId][coin.Name] = make(chan struct{})

		go func(ctx context.Context, coin models.Coin) {
			// This function needs for catching panics.
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "Recovered in goroutine in Service.StartTrading", slog.String("stacktrace", string(debug.Stack())), r)

					actionChanMap[userId] <- models.Message{
						Coin:   coin,
						Action: r.(string),
						File:   string(debug.Stack()),
					}
				}
			}()

			ctx = logging.WithCoinTag(ctx, coin.Name)

			for {
				select {
				case <-s.stopCoinMap[userId][coin.Name]:
					delete(s.stopCoinMap[userId], coin.Name)
					return
				default:
					err, eris := s.HandleCoinUpdate(ctx, coin, userId, actionChanMap)
					if err != nil {
						actionChanMap[userId] <- models.Message{
							Coin:   coin,
							Action: err.Error(),
							File:   eris.File,
							Line:   eris.Line,
						}
					}
				}
			}
		}(ctx, coin)
	}

	// Indicates that the user has started trading.
	user := models.NewUser(userId)
	user.TradingActivated = true

	err = s.uStorageRepo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) StopTrading(ctx context.Context, userID int64) error {
	user := models.NewUser(userID)
	user.TradingActivated = false

	err := s.uStorageRepo.UpdateUser(ctx, user)
	if err != nil {
		slog.ErrorContext(ctx, "Error updating user", err)
		return err
	}

	// Stopping and deleting all user.
	for coinName := range s.stopCoinMap[userID] {
		err = s.DeleteCoin(ctx, userID, coinName)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting coin", err)
			return err
		}
	}
	return nil
}

func (s *Service) DeleteCoin(ctx context.Context, userId int64, coinTag string) error {
	// Deleting coin from map of coins.
	if _, ok := s.stopCoinMap[userId][coinTag]; ok {
		s.stopCoinMap[userId][coinTag] <- struct{}{}
	}

	// Getting user from storage.
	user, err := s.uStorageRepo.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user", err)
		return err
	}

	// Getting coin from storage.
	coin, err := s.sStorageRepo.GetCoin(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin", err)
		return err
	}

	// Canceling sell order.
	if coin.SellOrderId != "" {
		cancelReq := models.CancelOrderRequest{
			Category: "spot",
			OrderId:  coin.SellOrderId,
			Symbol:   coin.Name,
		}

		_, err = s.apiRepo.CancelOrder(ctx, cancelReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error canceling sell order", err)
			return err
		}
	}

	// Canceling buy order.
	if coin.BuyOrderId != "" {
		cancelReq := models.CancelOrderRequest{
			Category: "spot",
			OrderId:  coin.BuyOrderId,
			Symbol:   coin.Name,
		}

		_, err = s.apiRepo.CancelOrder(ctx, cancelReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error canceling buy order", err)
			return err
		}
	}

	// If user already bought something.
	if len(coin.Buy) > 0 {
		// Getting coin's data from storage.
		coiniks, err := s.sStorageRepo.GetCoiniks(ctx, coin.Name)
		if err != nil {
			slog.ErrorContext(ctx, "Error getting coiniks", err)
			return err
		}

		// Creating sell order.
		createOrderReq := models.CreateOrderRequest{
			Category:    "spot",
			Side:        "Sell",
			Symbol:      coin.Name,
			OrderType:   "Market",
			Qty:         fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", coin.Count),
			TimeInForce: "GTC",
		}

		_, err = s.apiRepo.CreateOrder(ctx, createOrderReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating order", err)
			return err
		}

		// Getting current price of coin from api.
		getCoinReqParams := make(models.GetCoinRequest)
		getCoinReqParams["category"] = "spot"
		getCoinReqParams["symbol"] = coin.Name

		getCoinResp, err := s.apiRepo.GetCoin(ctx, getCoinReqParams, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error getting coin from algorithm", err)
			return err
		}

		currentPrice, err := strconv.ParseFloat(getCoinResp.Result.List[0].Price, 64)
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing current price to float", err)
			return err
		}

		// Counting income.
		var money float64
		for i := 0; i < len(coin.Buy); i++ {
			money += coin.Buy[i]
		}

		avg := money / float64(len(coin.Buy))

		spentMoney := avg * coin.Count
		earnMoney := currentPrice * coin.Count

		income := earnMoney - spentMoney

		err = s.sStorageRepo.InsertIncome(userId, coinTag, income, coin.Count)
		if err != nil {
			slog.ErrorContext(ctx, "Error insert income", err)
			return err
		}
	}

	// Setting all coin columns to null.
	err = s.sStorageRepo.SetCoinToDefault(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error setting coin to default", err)
		return err
	}

	return nil
}

// These functions do not need for implementing AlgorithmService.

func (s *Service) HandleCoinUpdate(ctx context.Context, coin models.Coin, userId int64, actionChanMap map[int64]chan models.Message) (error, models.Error) {
	var candik bool
	candik = true
	var eris models.Error
	user, err := s.uStorageRepo.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user from algorithm", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	// Getting coin from storage.
	coin, err = s.sStorageRepo.GetCoin(ctx, user.Id, coin.Name)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin from storage", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	// Getting user's wallet balance from api.
	getUserWalletParams := make(models.GetUserWalletRequest)
	getUserWalletParams["accountType"] = "UNIFIED"

	getUserWalletResp, err := s.apiRepo.GetUserWalletBalance(ctx, getUserWalletParams, user.ApiKey, user.SecretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user wallet balance", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	userUSDTBalance, err := strconv.ParseFloat(getUserWalletResp.Result.List[0].TotalEquity, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error converting user USDT wallet balance to float", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	user.USDTBalance = userUSDTBalance

	list, err := s.sStorageRepo.GetCoinList(ctx, user.Id)
	if err != nil {
		slog.ErrorContext(logging.ErrorCtx(ctx, err), "error in Algorithm.GetCoinList", err)
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
		}
	}

	if userSum > userUSDTBalance*0.95 {
		candik = false
	}
	// Getting current price of coin from api.
	getCoinReqParams := make(models.GetCoinRequest)
	getCoinReqParams["category"] = "spot"
	getCoinReqParams["symbol"] = coin.Name

	getCoinResp, err := s.apiRepo.GetCoin(ctx, getCoinReqParams, user.ApiKey, user.SecretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin from algorithm", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	currentPrice, err := strconv.ParseFloat(getCoinResp.Result.List[0].Price, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing current price to float", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	// Getting coiniks from storage.
	coiniks, err := s.sStorageRepo.GetCoiniks(ctx, coin.Name)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coiniks", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}

	if coin.Count > 0 {
		if coin.SellOrderId == "" {
			var sum float64
			for i := 0; i < len(coin.Buy); i++ {
				sum += coin.Buy[i]
			}
			avg := sum / float64(len(coin.Buy))

			createReq := models.CreateOrderRequest{
				Category:    "spot",
				Side:        "Sell",
				Symbol:      coin.Name,
				OrderType:   "Limit",
				Qty:         fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", coin.Count),
				Price:       fmt.Sprintf("%."+strconv.Itoa(coiniks.PriceDecimals)+"f", avg*(1+user.Percent)),
				TimeInForce: "GTC",
			}

			createOrderResp, err := s.apiRepo.CreateOrder(ctx, createReq, user.ApiKey, user.SecretKey)
			if err != nil {
				slog.ErrorContext(ctx, "Error creating order", err)
				return err, eris
			}

			updateCoin := models.NewCoin(user.Id, coin.Name)
			updateCoin.SellOrderId = createOrderResp.Result.OrderID

			err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
			if err != nil {
				slog.ErrorContext(ctx, "Error updating coin", err)
				return err, eris
			}
		}
	}

	if !user.Buy {
		if coin.Count == 0 {
			if coin.BuyOrderId != "" {
				cancelReq := models.CancelOrderRequest{
					Category: "spot",
					OrderId:  coin.BuyOrderId,
					Symbol:   coin.Name,
				}

				_, err = s.apiRepo.CancelOrder(ctx, cancelReq, user.ApiKey, user.SecretKey)
				if err != nil {
					slog.ErrorContext(ctx, "Error canceling order", err)
					_, eris.File, eris.Line, _ = runtime.Caller(0)
					return err, eris
				}
			}
			err = s.DeleteCoin(ctx, user.Id, coin.Name)
			if err != nil {
				slog.ErrorContext(ctx, "Error delete coin", err)
				_, eris.File, eris.Line, _ = runtime.Caller(0)
				return err, eris
			}
		}
	}

	// If price becomes higher than entry price and amount of coin equals 0 we should raise entryPrice.
	if currentPrice > coin.EntryPrice && coin.Count == 0 {
		slog.DebugContext(ctx, "Raising entry price")

		err = s.HandleRaisingEntryPrice(ctx, currentPrice, coin, user, coiniks, candik)
		if err != nil {
			slog.ErrorContext(ctx, "Error handling entry price", err)
			_, eris.File, eris.Line, _ = runtime.Caller(0)
			return err, eris
		}
		return nil, eris
	}

	if coin.BuyOrderId != "" {
		// Checking if BUY ORDER has been fulfilled.
		getReq := make(models.GetOrderRequest)
		getReq["category"] = "spot"
		getReq["orderId"] = coin.BuyOrderId // <- BUY ORDER ID!
		getReq["symbol"] = coin.Name

		getOrderResp, err := s.apiRepo.GetOrder(ctx, getReq, user.ApiKey, user.SecretKey)
		if getOrderResp.Result.List == nil {
			fmt.Println("у меня сдохла мать бубубубуб\n\n\n\n\n\n\n\n\n я даун")
		}
		if err != nil {
			slog.ErrorContext(ctx, "Error getting order", err)
			_, eris.File, eris.Line, _ = runtime.Caller(0)
			return err, eris
		}
		ctx = logging.WithOrderId(ctx, getOrderResp.Result.List[0].OrderId)

		// If that buy order has no errors and order status is filled.
		if len(getOrderResp.Result.List) > 0 && getOrderResp.Result.List[0].OrderStatus == SuccessfulOrderStatus && getOrderResp.Result.List[0].Side == "Buy" {
			slog.DebugContext(ctx, "fulfilled BUY ORDER was found", "resp", getOrderResp.Result.List[0])

			err = s.HandleFilledBuyOrder(ctx, getOrderResp, coin, user, coiniks, actionChanMap, candik)
			if err != nil {
				slog.ErrorContext(ctx, "Error handling filled buy order", err)
				_, eris.File, eris.Line, _ = runtime.Caller(0)
				return err, eris
			}
			return nil, eris
		}
	}

	// If SELL ORDER does not exist then skip.
	if coin.SellOrderId == "" {
		return nil, eris
	}

	// Checking if SELL ORDER has been fulfilled.
	getReq := make(models.GetOrderRequest)
	getReq["category"] = "spot"
	getReq["orderId"] = coin.SellOrderId // <- SELL ORDER ID!
	getReq["symbol"] = coin.Name

	getOrderResp, err := s.apiRepo.GetOrder(ctx, getReq, user.ApiKey, user.SecretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting order", err)
		_, eris.File, eris.Line, _ = runtime.Caller(0)
		return err, eris
	}
	ctx = logging.WithOrderId(ctx, getOrderResp.Result.List[0].OrderId)

	// If that sell order has no errors and order status is filled.
	if len(getOrderResp.Result.List) > 0 {
		if getOrderResp.Result.List[0].OrderStatus == SuccessfulOrderStatus && getOrderResp.Result.List[0].Side == "Sell" {
			slog.DebugContext(ctx, "fulfilled SELL ORDER was found")

			err = s.HandleFilledSellOrder(ctx, getOrderResp, coin, user, coiniks, actionChanMap)
			if err != nil {
				slog.ErrorContext(ctx, "Error handling filled sell order", err)
				_, eris.File, eris.Line, _ = runtime.Caller(0)
				return err, eris
			}
		}
	}
	return nil, eris
}

func (s *Service) HandleRaisingEntryPrice(ctx context.Context, currentPrice float64, coin models.Coin, user models.User, coiniks models.Coiniks, candik bool) error {
	coin.EntryPrice = currentPrice

	err := s.sStorageRepo.ResetCoin(ctx, coin, user)
	if err != nil {
		slog.ErrorContext(ctx, "Error update coin", err)
		return err
	}

	resetedCoin, err := s.sStorageRepo.GetCoin(ctx, user.Id, coin.Name)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting reseted coin", err)
		return err
	}

	// Canceling buy order.
	if resetedCoin.BuyOrderId != "" {
		cancelReq := models.CancelOrderRequest{
			Category: "spot",
			OrderId:  resetedCoin.BuyOrderId,
			Symbol:   resetedCoin.Name,
		}

		_, err = s.apiRepo.CancelOrder(ctx, cancelReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error canceling order", err)
			return err
		}
	}

	// Creating new buy order.
	if candik {
		createReq := models.CreateOrderRequest{
			Category:    "spot",
			Side:        "Buy",
			Symbol:      coin.Name,
			OrderType:   "Limit",
			Qty:         fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", user.USDTBalance*0.015/currentPrice),
			MarketUint:  "baseCoin",
			PositionIdx: 0,
			Price:       fmt.Sprintf("%."+strconv.Itoa(coiniks.PriceDecimals)+"f", coin.EntryPrice-resetedCoin.Decrement),
			TimeInForce: "GTC",
		}

		createOrderResp, err := s.apiRepo.CreateOrder(ctx, createReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating order", err)
			return err
		}
		ctx = logging.WithOrderId(ctx, createOrderResp.Result.OrderID)

		updateCoin := models.NewCoin(user.Id, coin.Name)
		updateCoin.BuyOrderId = createOrderResp.Result.OrderID

		err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
		if err != nil {
			slog.ErrorContext(ctx, "Error updating coin", err)
			return err
		}
	}

	return nil
}

func (s *Service) HandleFilledBuyOrder(ctx context.Context, getOrderResp models.GetOrderResponse, coin models.Coin, user models.User, coiniks models.Coiniks, actionChanMap map[int64]chan models.Message, candik bool) error {
	price, err := strconv.ParseFloat(getOrderResp.Result.List[0].Price, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing price to float", err)
		return err
	}

	coin.Buy = append(coin.Buy, price)

	count, err := strconv.ParseFloat(getOrderResp.Result.List[0].Qty, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing count to float", err)
		return err
	}

	fee, err := strconv.ParseFloat(getOrderResp.Result.List[0].CumExecFee, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing fee to float", err)
		return err
	}

	count -= fee

	dotIdx := strings.Index(fmt.Sprintf("%f", count), ".")

	count, err = strconv.ParseFloat(fmt.Sprintf("%f", count)[:dotIdx+1+coiniks.QtyDecimals], 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing count to float", err)
		return err
	}

	coin.Count += count

	err = s.sStorageRepo.UpdateCoin(ctx, coin)
	if err != nil {
		slog.ErrorContext(ctx, "Error updating coin", err)
		return err
	}

	if candik {
		// Creating new buy order.
		createReq := models.CreateOrderRequest{
			Category:    "spot",
			Side:        "Buy",
			Symbol:      coin.Name,
			OrderType:   "Limit",
			Qty:         fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", coin.Count/float64(len(coin.Buy))),
			Price:       fmt.Sprintf("%."+strconv.Itoa(coiniks.PriceDecimals)+"f", price-coin.Decrement),
			TimeInForce: "GTC",
		}

		createOrderResp, err := s.apiRepo.CreateOrder(ctx, createReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating order", err)
			return err
		}

		updateCoin := models.NewCoin(user.Id, coin.Name)
		if createOrderResp.Result.OrderID == "" {
			return nil
		}
		updateCoin.BuyOrderId = createOrderResp.Result.OrderID

		err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
		if err != nil {
			slog.ErrorContext(ctx, "Error updating coin", err)
			return err
		}
	}

	// Canceling old sell order, if it exists.
	if coin.SellOrderId != "" {
		cancelReq := models.CancelOrderRequest{
			Category: "spot",
			OrderId:  coin.SellOrderId,
			Symbol:   coin.Name,
		}

		_, err = s.apiRepo.CancelOrder(ctx, cancelReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error canceling order", err)
			return err
		}
	}

	// Creating new sell order.
	var sum float64
	for i := 0; i < len(coin.Buy); i++ {
		sum += coin.Buy[i]
	}
	avg := sum / float64(len(coin.Buy))

	createReq := models.CreateOrderRequest{
		Category:    "spot",
		Side:        "Sell",
		Symbol:      coin.Name,
		OrderType:   "Limit",
		Qty:         fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", coin.Count),
		Price:       fmt.Sprintf("%."+strconv.Itoa(coiniks.PriceDecimals)+"f", avg*(1+user.Percent)),
		TimeInForce: "GTC",
	}

	createOrderResp, err := s.apiRepo.CreateOrder(ctx, createReq, user.ApiKey, user.SecretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Error creating order", err)
		return err
	}

	updateCoin := models.NewCoin(user.Id, coin.Name)
	updateCoin.SellOrderId = createOrderResp.Result.OrderID

	err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
	if err != nil {
		slog.ErrorContext(ctx, "Error updating coin", err)
		return err
	}

	// Sending message for goroutine from handler to notify user about buy
	msg := models.Message{
		User:   user,
		Coin:   coin,
		Action: bot.BuyAction,
	}

	actionChanMap[user.Id] <- msg

	return nil
}

func (s *Service) HandleFilledSellOrder(ctx context.Context, getOrderResp models.GetOrderResponse, coin models.Coin, user models.User, coiniks models.Coiniks, actionChanMap map[int64]chan models.Message) error {
	sellPrice, err := strconv.ParseFloat(getOrderResp.Result.List[0].Price, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing price to float", err)
		return err
	}

	err = s.sStorageRepo.SellCoin(user.Id, coin.Name, sellPrice)
	if err != nil {
		slog.ErrorContext(ctx, "Error updating coin", err)
		return err
	}

	// Canceling old buy order.
	cancelReq := models.CancelOrderRequest{
		Category: "spot",
		OrderId:  coin.BuyOrderId,
		Symbol:   coin.Name,
	}

	_, err = s.apiRepo.CancelOrder(ctx, cancelReq, user.ApiKey, user.SecretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Error canceling order", err)
	}

	updateCoin := models.NewCoin(user.Id, coin.Name)
	updateCoin.EntryPrice = sellPrice

	err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
	if err != nil {
		slog.ErrorContext(ctx, "Error updating coin", err)
		return err
	}

	resetedCoin, err := s.sStorageRepo.GetCoin(ctx, user.Id, coin.Name)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting reseted coin", err)
		return err
	}

	if user.Buy {
		// Creating new buy order.
		createReq := models.CreateOrderRequest{
			Category:    "spot",
			Side:        "Buy",
			Symbol:      coin.Name,
			OrderType:   "Limit",
			Qty:         fmt.Sprintf("%."+strconv.Itoa(coiniks.QtyDecimals)+"f", user.USDTBalance*0.015/sellPrice),
			MarketUint:  "baseCoin",
			PositionIdx: 0,
			Price:       fmt.Sprintf("%."+strconv.Itoa(coiniks.PriceDecimals)+"f", resetedCoin.EntryPrice-resetedCoin.Decrement),
			TimeInForce: "GTC",
		}

		createOrderResp, err := s.apiRepo.CreateOrder(ctx, createReq, user.ApiKey, user.SecretKey)
		if err != nil {
			slog.ErrorContext(ctx, "Error creating order", err)
			return err
		}

		updateCoin = models.NewCoin(user.Id, coin.Name)
		updateCoin.BuyOrderId = createOrderResp.Result.OrderID
		updateCoin.SellOrderId = "setNull"

		err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
		if err != nil {
			slog.ErrorContext(ctx, "Error updating coin", err)
			return err
		}
	}

	// Sending message for goroutine from handler to notify user about sell
	msg := models.Message{
		User:   user,
		Coin:   coin,
		Action: bot.SellAction,
	}

	var sum float64
	for i := 0; i < len(coin.Buy); i++ {
		sum += coin.Buy[i]
	}

	avg := sum / float64(len(coin.Buy))

	income := sellPrice*coin.Count - avg*coin.Count

	err = s.sStorageRepo.InsertIncome(user.Id, coin.Name, income, coin.Count)
	if err != nil {
		slog.ErrorContext(ctx, "Error inserting income", err)
		return err
	}

	coin.Count = 0

	msg.Coin.CurrentPrice = sellPrice
	msg.Coin.Income = income

	if !user.Buy {
		err = s.DeleteCoin(ctx, user.Id, coin.Name)
		return err
	}

	actionChanMap[user.Id] <- msg

	return nil
}
