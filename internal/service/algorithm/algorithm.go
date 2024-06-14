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

					coin, err = s.sStorageRepo.GetCoin(ctx, user.Id, coin.Name)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting coin from storage", err)
					}

					params := make(models.GetCoinRequest)
					params["category"] = "spot"
					params["symbol"] = coin.Name

					byteParams, err := json.Marshal(params)
					if err != nil {
						slog.ErrorContext(ctx, "Error marshalling params to json", err)
					}

					var getCoinResp models.GetCoinResponse
					body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(byteParams), GetCoinEndpoint, http.MethodGet, user.ApiKey, user.SecretKey)
					if err != nil {
						slog.ErrorContext(ctx, "Error creating get coin request", err)
					}

					err = json.Unmarshal(body, &getCoinResp)
					if err != nil {
						slog.ErrorContext(ctx, "Error unmarshalling get coin response", err)
					}

					if getCoinResp.RetMsg != "OK" {
						slog.ErrorContext(ctx, "Error getting coin response: ", err)
					}

					currentPrice, err := strconv.ParseFloat(getCoinResp.Result.List[0].Price, 64)
					if err != nil {
						slog.ErrorContext(ctx, "Error parsing current price to float", err)
					}

					coiniks, err := s.sStorageRepo.GetCoiniks(ctx, coin.Name)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting coiniks", err)
						return
					}

					// If price becomes higher than entry price and amount of coin equals 0 we should raise entryPrice.
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

						// Canceling buy order.
						if resetedCoin.BuyOrderId != "" {
							cancelReq := map[string]interface{}{
								"category": "spot",
								"symbol":   resetedCoin.Name,
								"orderId":  resetedCoin.BuyOrderId,
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
						z := "%." + strconv.Itoa(coiniks.QtyDecimals) + "f"
						y := "%." + strconv.Itoa(coiniks.PriceDecimals) + "f"
						createReq := map[string]interface{}{
							"category":    "spot",
							"side":        "Buy",
							"symbol":      coin.Name,
							"orderType":   "Limit",
							"qty":         fmt.Sprintf(z, user.Balance*0.015/currentPrice),
							"marketUint":  "baseCoin",
							"positionIdx": 0,
							"price":       fmt.Sprintf(y, coin.EntryPrice-coin.Decrement),
							"timeInForce": "GTC",
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

						if createOrderResp.RetMsg != "OK" {
							slog.ErrorContext(ctx, "Error creating order: ", createOrderResp.RetMsg)
						}

						updateCoin := models.NewCoin(user.Id, coin.Name)
						updateCoin.BuyOrderId = createOrderResp.Result.OrderID

						err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
						if err != nil {
							slog.ErrorContext(ctx, "Error updating coin", err)
						}
						continue
					}

					// Checking if BUY ORDER has been fulfilled.
					getReq := make(models.GetOrderRequest)
					getReq["category"] = "spot"
					getReq["orderId"] = coin.BuyOrderId // <- BUY ORDER ID!
					getReq["symbol"] = coin.Name

					jsonData, err := json.Marshal(getReq)
					if err != nil {
						slog.ErrorContext(ctx, "Error marshaling create order", err)
					}

					body, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), GetOrderEndpoint, http.MethodGet, user.ApiKey, user.SecretKey)
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

					if len(getOrderResp.Result.List) > 0 && getOrderResp.Result.List[0].OrderStatus == SuccessfulOrderStatus {
						if getOrderResp.Result.List[0].Side == "Buy" {
							slog.DebugContext(ctx, "fulfilled BUY ORDER was found")

							z := "%." + strconv.Itoa(coiniks.PriceDecimals) + "f"
							price, err := strconv.ParseFloat(fmt.Sprintf(z, coin.EntryPrice-coin.Decrement), 64)
							if err != nil {
								fmt.Println(err)
							}

							coin.Buy = append(coin.Buy, price)

							count, _ := strconv.ParseFloat(getOrderResp.Result.List[0].Qty, 64)
							coin.Count += count

							err = s.sStorageRepo.UpdateCoin(ctx, coin)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Creating new buy order
							z = "%." + strconv.Itoa(coiniks.QtyDecimals) + "f"
							g := "%." + strconv.Itoa(coiniks.PriceDecimals) + "f"
							createReq := map[string]interface{}{
								"category":    "spot",
								"Side":        "Buy",
								"symbol":      coin.Name,
								"qrderType":   "Limit",
								"qty":         fmt.Sprintf(z, coin.Count/float64(len(coin.Buy))),
								"price":       fmt.Sprintf(g, price-coin.Decrement),
								"timeInForce": "GTC",
							}

							jsonData, err = json.Marshal(createReq)
							if err != nil {
								slog.ErrorContext(ctx, "Error marshaling create order", err)
							}

							body, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), CreateOrderEndpoint, http.MethodPost, user.ApiKey, user.SecretKey)
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
							cancelReq := map[string]interface{}{
								"category": "spot",
								"orderId":  coin.SellOrderId,
								"symbol":   coin.Name,
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

							createReq = map[string]interface{}{
								"category":    "spot",
								"side":        "Sell",
								"symbol":      coin.Name,
								"orderType":   "Limit",
								"qty":         fmt.Sprintf("%f", coin.Count),
								"price":       fmt.Sprintf("%f", avg),
								"timeInForce": "GTC",
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

							continue
						} else {
							fmt.Println(fmt.Sprintf("броо тут какая то хуйня : %s", getOrderResp.Result.List[0].Side))
						}

						//switch getOrderResp.List[0].Side {
						//case "Sell":
						//	sellPrice, _ := strconv.ParseFloat(getOrderResp.List[0].Price, 64)
						//
						//	err = s.sStorageRepo.SellCoin(user.Id, coin.Name, sellPrice)
						//	if err != nil {
						//		slog.ErrorContext(ctx, "Error updating coin", err)
						//	}
						//
						//	// Sending message for goroutine from handler to notify user about sell
						//	msg := models.Message{
						//		User:   user,
						//		Coin:   coin,
						//		Action: algorithm.SellAction,
						//	}
						//
						//	coin.Count = 0
						//	msg.Coin.CurrentPrice = currentPrice
						//	actionChanMap[userId] <- msg
						//}
					}

					// Checking if sell order has been executed.
					getReq = make(models.GetOrderRequest)
					getReq["category"] = "spot"
					getReq["orderId"] = coin.SellOrderId // <- SELL ORDER ID!
					getReq["symbol"] = coin.Name

					jsonData, err = json.Marshal(getReq)
					if err != nil {
						slog.ErrorContext(ctx, "Error marshaling create order", err)
					}

					body, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), GetOrderEndpoint, http.MethodGet, user.ApiKey, user.SecretKey)
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

					if len(getOrderResp.Result.List) > 0 && getOrderResp.Result.List[0].OrderStatus == SuccessfulOrderStatus {
						slog.DebugContext(ctx, "successfully SELL ORDER was found")

						if getOrderResp.Result.List[0].Side == "Sell" {
							sellPrice, _ := strconv.ParseFloat(getOrderResp.Result.List[0].Price, 64)
							err = s.sStorageRepo.SellCoin(user.Id, coin.Name, sellPrice)
							if err != nil {
								slog.ErrorContext(ctx, "Error updating coin", err)
							}

							// Canceling old buy order
							cancelReq := map[string]interface{}{
								"category": "spot",
								"orderID":  coin.BuyOrderId,
								"symbol":   coin.Name,
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
							z := "%." + strconv.Itoa(coiniks.QtyDecimals) + "f"
							y := "%." + strconv.Itoa(coiniks.PriceDecimals) + "f"

							createReq := map[string]interface{}{
								"category":    "spot",
								"side":        "Buy",
								"symbol":      coin.Name,
								"orderType":   "Limit",
								"qty":         fmt.Sprintf(z, user.Balance*0.015/sellPrice),
								"marketUint":  "baseCoin",
								"positionIdx": 0,
								"price":       fmt.Sprintf(y, coin.EntryPrice-coin.Decrement),
								"timeInForce": "GTC",
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

							err = s.sStorageRepo.InsertIncome(user.Id, coin.Name, income, coin.Count)
							if err != nil {
								slog.ErrorContext(ctx, "Error inserting income", err)
							}

							coin.Count = 0

							msg.Coin.CurrentPrice = sellPrice
							msg.Coin.Income = income

							actionChanMap[userId] <- msg
						} else {
							fmt.Println(fmt.Sprintf("броо тут какая то хуйня Sell: %s", getOrderResp.Result.List[0].Side))
						}
					}
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
