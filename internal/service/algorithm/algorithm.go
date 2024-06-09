package algorithm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"m1pes/internal/algorithm"
	"m1pes/internal/models"

	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
	storageUser "m1pes/internal/repository/storage/user"
)

const (
	CreateOrderEndpoint   = "/v5/order/create"
	CancelOrderEndpoint   = "/v5/order/cancel"
	GetOrderEndpoint      = "/v5/order/realtime"
	GetCoinEndpoint       = "/v5/market/tickers"
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
		return err
	}

	for _, coin := range coinList {
		// init map that stores coin name as key and map2 as value
		// map2 stores userId as key and struct{} as value
		if _, ok := s.stopCoinMap[userId][coin.Name]; ok {
			continue
		}

		if _, ok := s.stopCoinMap[userId]; !ok {
			s.stopCoinMap[userId] = make(map[string]chan struct{})
		}
		s.stopCoinMap[userId][coin.Name] = make(chan struct{})
		go func(coin models.Coin) {
			for {
				select {
				case <-s.stopCoinMap[userId][coin.Name]:
					delete(s.stopCoinMap[userId], coin.Name)
					return
				default:
					user, err := s.uStorageRepo.GetUser(ctx, userId)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting user from algorithm", err)
						return
					}

					//currentPrice, err := s.apiRepo.GetPrice(ctx, coin.Name, user.ApiKey)
					//if err != nil {
					//	slog.ErrorContext(ctx, "Error getting price from api", err)
					//	return
					//}

					//getCoinReq := models.GetCoinRequest{
					//	Category: "option",
					//	Symbol:   coin.Name,
					//}
					//jsonData, err := json.Marshal(getCoinReq)
					//if err != nil {
					//	slog.ErrorContext(ctx, "Error marshalling get coin request", err)
					//}

					weirdMap := make(map[string]interface{})
					weirdMap["category"] = "spot"
					weirdMap["symbol"] = coin.Name

					var params string
					var i int
					for key, val := range weirdMap {
						if i == 0 {
							params += fmt.Sprintf("%s=%v", key, val)
							i++
							continue
						}
						params += fmt.Sprintf("&%s=%v", key, val)
					}

					var getCoinResp models.GetCoinResponse
					body, err := s.apiRepo.CreateSignRequestAndGetRespBody(params, GetCoinEndpoint, http.MethodGet, user.ApiKey, user.SecretKey)
					if err != nil {
						slog.ErrorContext(ctx, "Error creating get coin request", err)
					}

					err = json.Unmarshal(body, &getCoinResp)
					if err != nil {
						slog.ErrorContext(ctx, "Error unmarshalling get coin response", err)
					}

					//fmt.Println(getCoinResp)

					currentPrice, err := strconv.ParseFloat(getCoinResp.Result.List[0].Price, 64)
					if err != nil {
						slog.ErrorContext(ctx, "Error parsing current price to float", err)
					}

					//fmt.Println("currentPrice:", currentPrice)

					coin, err = s.sStorageRepo.GetCoin(ctx, user.Id, coin.Name)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting coin from storage", err)
					}
					//fmt.Println(coin)

					// If price becomes higher than entry price and amount of coin equals 0 we should raise entryPrice

					//fmt.Println(fmt.Sprintf("curr price:%v\temtry price:%v", currentPrice, coin.EntryPrice))

					if currentPrice > coin.EntryPrice && coin.Count == 0 {
						slog.DebugContext(ctx, "Raising entry price")

						coin.EntryPrice = currentPrice

						err = s.sStorageRepo.ResetCoin(ctx, coin, user)
						if err != nil {
							slog.ErrorContext(ctx, "Error update coin", err)
							return
						}

						resetedCoin, err := s.sStorageRepo.GetCoin(ctx, userId, coin.Name)
						if err != nil {
							slog.ErrorContext(ctx, "Error getting reseted coin", err)
							return
						}

						// canceling buy order
						if resetedCoin.BuyOrderId != "" {
							cancelReq := models.CancelOrderRequest{
								Category: "spot",
								OrderId:  resetedCoin.BuyOrderId,
								Symbol:   resetedCoin.Name,
							}
							jsonData, err := json.Marshal(cancelReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshalling cancel order", err)
							}

							_, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CancelOrderEndpoint, http.MethodPost, user.ApiKey, user.SecretKey)
							if err != nil {
								slog.ErrorContext(ctx, "Error creating sign request", err)
							}
						}

						// creating new buy order
						z := "%." + strconv.Itoa(resetedCoin.QtyDecimals) + "f"
						y := "%." + strconv.Itoa(resetedCoin.PriceDecimals) + "f"
						createReq := models.CreateOrderRequest{
							Category:    "spot",
							Side:        "Buy",
							Symbol:      coin.Name,
							OrderType:   "Limit",
							Qty:         fmt.Sprintf(z, coin.Count/float64(len(coin.Buy))),
							MarketUint:  "baseCoin",
							PositionIdx: 0,
							Price:       fmt.Sprintf(y, coin.EntryPrice-coin.Decrement), //[:len(fmt.Sprintf("%.4f", coin.EntryPrice-coin.Decrement))-1],
							TimeInForce: "GTC",
						}
						jsonData, err := json.Marshal(createReq)
						if err != nil {
							slog.ErrorContext(ctx, "Error marshaling create order", err)
						}

						body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, "POST", user.ApiKey, user.SecretKey)
						if err != nil {
							slog.ErrorContext(ctx, "Error creating sign request", err)
						}

						var createOrderResp models.CreateOrderResponse
						err = json.Unmarshal(body, &createOrderResp)
						if err != nil {
							slog.ErrorContext(ctx, "Error unmarshalling sign request", err)
						}

						//fmt.Println("response:", createOrderResp)

						updateCoin := models.NewCoin(user.Id, coin.Name)
						updateCoin.BuyOrderId = createOrderResp.Result.OrderID

						err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
						if err != nil {
							slog.ErrorContext(ctx, "Error updating coin", err)
						}
						continue
					}

					//Check buy order
					getReq := models.GetOrderRequest{
						Category: "spot",
						OrderID:  coin.BuyOrderId,
						Symbol:   coin.Name,
					}

					jsonData, err := json.Marshal(getReq)
					if err != nil {
						slog.ErrorContext(ctx, "Error marshaling create order", err)
					}

					body, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), GetOrderEndpoint, http.MethodGet, user.ApiKey, user.ApiKey)
					if err != nil {
						slog.ErrorContext(ctx, "Error creating sign request", err)
					}

					if string(body) == "" {
						continue
					}

					var getOrderResp models.GetOrderResponse

					err = json.Unmarshal(body, &getOrderResp)
					if err != nil {
						slog.ErrorContext(ctx, "Error unmarshalling create order", err)
					}

					// If order is completed successfully we should to set new buy and sell order, else we are waiting
					if getOrderResp.List[0].OrderStatus == SuccessfulOrderStatus {
						slog.DebugContext(ctx, "successfully order was found")

						if getOrderResp.List[0].Side == "Buy" {
							z := "%." + strconv.Itoa(coin.PriceDecimals) + "f"
							price, err := strconv.ParseFloat(fmt.Sprintf(z, coin.EntryPrice-coin.Decrement), 64)
							if err != nil {
								fmt.Println(err)
							}

							coin.Buy = append(coin.Buy, price)

							count, _ := strconv.ParseFloat(getOrderResp.List[0].Qty, 64)
							coin.Count += count

							err = s.sStorageRepo.UpdateCoin(ctx, coin)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Creating new buy order
							z = "%." + strconv.Itoa(coin.QtyDecimals) + "f"
							g := "%." + strconv.Itoa(coin.PriceDecimals) + "f"
							createReq := models.CreateOrderRequest{
								Category:    "spot",
								Side:        "Buy",
								Symbol:      coin.Name,
								OrderType:   "Limit",
								Qty:         fmt.Sprintf(z, coin.Count/float64(len(coin.Buy))),
								Price:       fmt.Sprintf(g, price-coin.Decrement),
								TimeInForce: "GTC",
							}
							jsonData, err = json.Marshal(createReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshaling create order", err)
							}

							body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, "POST", user.ApiKey, user.SecretKey)
							if err != nil {
								slog.ErrorContext(ctx, "Error creating sign request", err)
							}

							var createOrderResp models.CreateOrderResponse
							err = json.Unmarshal(body, &createOrderResp)
							if err != nil {
								slog.ErrorContext(ctx, "Error unmarshalling sign request", err)
							}

							updateCoin := models.NewCoin(user.Id, coin.Name)
							updateCoin.BuyOrderId = createOrderResp.Result.OrderID

							err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Canceling old sell order
							cancelReq := models.CancelOrderRequest{
								Category: "spot",
								OrderId:  coin.SellOrderId,
								Symbol:   coin.Name,
							}
							jsonData, err := json.Marshal(cancelReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshalling cancel order", err)
							}

							_, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CancelOrderEndpoint, http.MethodPost, user.ApiKey, user.SecretKey)
							if err != nil {
								slog.ErrorContext(ctx, "Error creating sign request", err)
							}

							// Creating new sell order
							var sum float64
							for i := 0; i < len(coin.Buy); i++ {
								sum += coin.Buy[i]
							}
							avg := sum / float64(len(coin.Buy))

							createReq = models.CreateOrderRequest{
								Category:    "spot",
								Side:        "Sell",
								Symbol:      coin.Name,
								OrderType:   "Limit",
								Qty:         fmt.Sprintf("%f", coin.Count),
								Price:       fmt.Sprintf("%f", avg),
								TimeInForce: "GTC",
							}
							jsonData, err = json.Marshal(createReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshaling create order", err)
							}

							body, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, "POST", user.ApiKey, user.SecretKey)
							if err != nil {
								slog.ErrorContext(ctx, "Error creating sign request", err)
							}

							err = json.Unmarshal(body, &createOrderResp)
							if err != nil {
								slog.ErrorContext(ctx, "Error unmarshalling sign request", err)
							}

							updateCoin = models.NewCoin(user.Id, coin.Name)
							updateCoin.SellOrderId = createOrderResp.Result.OrderID

							err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Sending message for goroutine from handler to notify user about buy
							msg := models.Message{
								User:   user,
								Coin:   coin,
								Action: algorithm.BuyAction,
							}

							actionChanMap[user.Id] <- msg
						} else {
							fmt.Println(fmt.Sprintf("броо тут какая то хуйня : %s", getOrderResp.List[0].Side))
						}

						switch getOrderResp.List[0].Side {
						case "Sell":
							sellPrice, _ := strconv.ParseFloat(getOrderResp.List[0].Price, 64)

							err = s.sStorageRepo.SellCoin(user.Id, coin.Name, sellPrice)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Sending message for goroutine from handler to notify user about sell
							msg := models.Message{
								User:   user,
								Coin:   coin,
								Action: algorithm.SellAction,
							}

							coin.Count = 0
							msg.Coin.CurrentPrice = currentPrice
							actionChanMap[userId] <- msg
						}
					}

					//Check sell order
					getReq = models.GetOrderRequest{
						Category: "spot",
						OrderID:  coin.SellOrderId,
						Symbol:   coin.Name,
					}

					jsonData, err = json.Marshal(getReq)
					if err != nil {
						slog.ErrorContext(ctx, "Error marshaling create order", err)
					}

					body, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), GetOrderEndpoint, http.MethodGet, user.ApiKey, user.ApiKey)
					if err != nil {
						slog.ErrorContext(ctx, "Error creating sign request", err)
					}

					if string(body) == "" {
						continue
					}

					var getOrderResponse models.GetOrderResponse

					err = json.Unmarshal(body, &getOrderResponse)
					if err != nil {
						slog.ErrorContext(ctx, "Error unmarshalling create order", err)
					}

					if getOrderResp.List[0].OrderStatus == SuccessfulOrderStatus {
						slog.DebugContext(ctx, "successfully order was found")

						if getOrderResp.List[0].Side == "Sell" {
							sellPrice, _ := strconv.ParseFloat(getOrderResp.List[0].Price, 64)
							err = s.sStorageRepo.SellCoin(user.Id, coin.Name, sellPrice)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Canceling old buy order
							cancelReq := models.CancelOrderRequest{
								Category: "spot",
								OrderId:  coin.BuyOrderId,
								Symbol:   coin.Name,
							}
							jsonData, err := json.Marshal(cancelReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshalling cancel order", err)
							}

							_, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CancelOrderEndpoint, http.MethodPost, user.ApiKey, user.SecretKey)
							if err != nil {
								slog.ErrorContext(ctx, "Error creating sign request", err)
							}

							// creating new buy order
							z := "%." + strconv.Itoa(coin.QtyDecimals) + "f"
							y := "%." + strconv.Itoa(coin.PriceDecimals) + "f"
							createReq := models.CreateOrderRequest{
								Category:    "spot",
								Side:        "Buy",
								Symbol:      coin.Name,
								OrderType:   "Limit",
								Qty:         fmt.Sprintf(z, user.Balance*0.015/sellPrice),
								MarketUint:  "baseCoin",
								PositionIdx: 0,
								Price:       fmt.Sprintf(y, coin.EntryPrice-coin.Decrement),
								TimeInForce: "GTC",
							}

							jsonData, err = json.Marshal(createReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshaling create order", err)
							}

							body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, "POST", user.ApiKey, user.SecretKey)
							if err != nil {
								slog.ErrorContext(ctx, "Error creating sign request", err)
							}

							var createOrderResp models.CreateOrderResponse
							err = json.Unmarshal(body, &createOrderResp)
							if err != nil {
								slog.ErrorContext(ctx, "Error unmarshalling sign request", err)
							}

							//fmt.Println("response:", createOrderResp)

							updateCoin := models.NewCoin(user.Id, coin.Name)
							updateCoin.BuyOrderId = createOrderResp.Result.OrderID

							err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Sending message for goroutine from handler to notify user about sell
							msg := models.Message{
								User:   user,
								Coin:   coin,
								Action: algorithm.SellAction,
							}
							var sum float64
							for i := 0; i < len(coin.Buy); i++ {
								sum += coin.Buy[i]
							}
							avg := sum / float64(len(coin.Buy))
							income := sellPrice*coin.Count - avg*coin.Count
							s.sStorageRepo.InsertIncome(user.Id, coin.Name, income, coin.Count)
							coin.Count = 0
							msg.Coin.CurrentPrice = sellPrice
							msg.Coin.Income = income
							actionChanMap[userId] <- msg
						} else {
							fmt.Println(fmt.Sprintf("броо тут какая то хуйня Sell: %s", getOrderResp.List[0].Side))
						}
					}

					//switch status {
					//case algorithm.ChangeAction:
					//	err = s.sStorageRepo.ResetCoin(ctx, coin, user)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error update coin", err)
					//		return
					//	}
					//
					//	// canceling buy order
					//	cancelReq := models.CancelOrderRequest{
					//		Category: "spot",
					//		OrderId:  coin.BuyOrderId,
					//		Symbol:   coin.Name,
					//	}
					//	jsonData, err = json.Marshal(cancelReq)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error marshalling cancel order", err)
					//	}
					//
					//	_, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CancelOrderEndpoint, "POST", "ZD2sewIvQg6deMclTN", "4FtzrEEpz8UYRxDyg3vzcLw0SR48KpOdO5A5")
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error creating sign request", err)
					//	}
					//
					//	// creating new buy order
					//	qty, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", coin.Count+user.Balance*0.02/currentPrice), 64)
					//
					//	createReq := models.CreateOrderRequest{
					//		Category:    "spot",
					//		Side:        "Buy",
					//		Symbol:      coin.Name,
					//		OrderType:   "Limit",
					//		Qty:         qty,
					//		Price:       coin.EntryPrice - coin.Decrement,
					//		TimeInForce: "GTC",
					//	}
					//	jsonData, err = json.Marshal(createReq)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error marshaling create order", err)
					//	}
					//
					//	body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, "POST", "ZD2sewIvQg6deMclTN", "4FtzrEEpz8UYRxDyg3vzcLw0SR48KpOdO5A5")
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error creating sign request", err)
					//	}
					//
					//	var resp models.CreateOrderResponse
					//	err = json.Unmarshal(body, &resp)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error unmarshalling sign request", err)
					//	}
					//
					//	updateCoin := models.NewCoin(user.Id, coin.Name)
					//	updateCoin.BuyOrderId = resp.Result.OrderID
					//	err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error updating coin", err)
					//	}
					//case algorithm.WaitAction:
					//	continue
					//case algorithm.BuyAction:
					//	updateCoin := models.NewCoin(user.Id, coin.Name)
					//	updateCoin.Buy = coin.Buy
					//	updateCoin.Count = coin.Count
					//	updateCoin.Decrement = coin.Decrement
					//
					//	err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error update count", err)
					//		return
					//	}
					//	msg.Action = algorithm.BuyAction
					//	actionChanMap[userId] <- msg
					//case algorithm.SellAction:
					//	coin.Decrement = currentPrice * user.Percent
					//	err = s.sStorageRepo.SellCoin(userId, coin.Name, currentPrice, coin.Decrement)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error update SellAction", err)
					//		return
					//	}
					//
					//	coin.Count = 0
					//	msg.Coin.CurrentPrice = currentPrice
					//	msg.Action = algorithm.SellAction
					//	actionChanMap[userId] <- msg
					//}

					//status := algorithm.Algorithm(currentPrice, &coin, &user)
					//
					//msg := models.Message{
					//	User: user,
					//	Coin: coin,
					//}
					//
					//switch status {
					//case algorithm.ChangeAction:
					//	err = s.sStorageRepo.ResetCoin(ctx, coin, user)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error update coin", err)
					//		return
					//	}
					//case algorithm.WaitAction:
					//	continue
					//case algorithm.BuyAction:
					//	updateCoin := models.NewCoin(user.Id, coin.Name)
					//	updateCoin.Buy = coin.Buy
					//	updateCoin.Count = coin.Count
					//	updateCoin.Decrement = coin.Decrement
					//
					//	err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error update count", err)
					//		return
					//	}
					//
					//	createReq := models.CreateOrderRequest{
					//		Category:    "spot",
					//		Side:        "Buy",
					//		Symbol:      coin.Name,
					//		OrderType:   "Market",
					//		Qty:         fmt.Sprintf("%.2f", coin.Count/float64(len(coin.Buy))),
					//		TimeInForce: "GTC",
					//	}
					//	jsonData, err := json.Marshal(createReq)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error marshaling create order", err)
					//	}
					//
					//	body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, http.MethodPost, user.ApiKey, user.SecretKey)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error creating sign request", err)
					//	}
					//
					//	fmt.Println(string(body))
					//
					//	msg.Action = algorithm.BuyAction
					//	actionChanMap[userId] <- msg
					//case algorithm.SellAction:
					//	coin.Decrement = currentPrice * user.Percent
					//
					//	createReq := models.CreateOrderRequest{
					//		Category:    "spot",
					//		Side:        "Sell",
					//		Symbol:      coin.Name,
					//		OrderType:   "Market",
					//		Qty:         fmt.Sprintf("%.2f", coin.Count),
					//		TimeInForce: "GTC",
					//	}
					//	jsonData, err := json.Marshal(createReq)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error marshaling create order", err)
					//	}
					//
					//	body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, http.MethodPost, user.ApiKey, user.SecretKey)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error creating sign request", err)
					//	}
					//
					//	fmt.Println(string(body))
					//
					//	err = s.sStorageRepo.SellCoin(userId, coin.Name, currentPrice, coin.Decrement)
					//	if err != nil {
					//		slog.ErrorContext(ctx, "Error update SellAction", err)
					//		return
					//	}
					//
					//	coin.Count = 0
					//	msg.Coin.CurrentPrice = currentPrice
					//	msg.Action = algorithm.SellAction
					//	actionChanMap[userId] <- msg
					//}
				}
			}
		}(coin)
	}

	// Indicate that the user has started trading
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
		return err
	}

	for coinName := range s.stopCoinMap[userID] {
		s.stopCoinMap[userID][coinName] <- struct{}{}
	}
	return nil
}

func (s *Service) DeleteCoin(ctx context.Context, userId int64, coinTag string) error {
	if _, ok := s.stopCoinMap[userId][coinTag]; ok {
		s.stopCoinMap[userId][coinTag] <- struct{}{}
	}

	user, err := s.uStorageRepo.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user", err)
		return err
	}
	currentPrice, err := s.apiRepo.GetPrice(ctx, coinTag, user.ApiKey)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting price from api", err)
		return err
	}

	coin, err := s.sStorageRepo.GetCoin(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin", err)
		return err
	}

	var money float64
	for i := 0; i < len(coin.Buy); i++ {
		money += coin.Buy[i]
	}
	avg := money / float64(len(coin.Buy))
	spentMoney := avg * coin.Count
	earnMoney := currentPrice * coin.Count

	income := earnMoney - spentMoney

	err = s.sStorageRepo.DeleteCoin(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error delete coin", err)
		return err
	}

	err = s.uStorageRepo.ChangeBalance(ctx, userId, income)
	if err != nil {
		slog.ErrorContext(ctx, "Error change balance", err)
		return err
	}

	err = s.sStorageRepo.InsertIncome(userId, coinTag, income, coin.Count)
	if err != nil {
		slog.ErrorContext(ctx, "Error insert income", err)
		return err
	}

	return nil
}
