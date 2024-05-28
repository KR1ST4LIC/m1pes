package algorithm

import (
	"context"
	"log/slog"
	"m1pes/internal/models"

	"m1pes/internal/algorithm"
	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
	storageUser "m1pes/internal/repository/storage/user"
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

	for _, coin := range coinList.Name {
		// init map that stores coin name as key and map2 as value
		// map2 stores userId as key and struct{} as value
		if _, ok := s.stopCoinMap[userId][coin]; ok {
			continue
		}

		if _, ok := s.stopCoinMap[userId]; !ok {
			s.stopCoinMap[userId] = make(map[string]chan struct{})
		}
		s.stopCoinMap[userId][coin] = make(chan struct{})
		go func(funcCoin string) {
			for {
				select {
				case <-s.stopCoinMap[userId][funcCoin]:
					delete(s.stopCoinMap[userId], funcCoin)
					return
				default:
					currentPrice, err := s.apiRepo.GetPrice(ctx, funcCoin)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting price from api", err)
						return
					}

					user, err := s.uStorageRepo.GetUser(ctx, userId)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting user from algorithm", err)
						return
					}

					coin, err := s.sStorageRepo.GetCoin(ctx, userId, funcCoin)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting coin from algorithm", err)
						return
					}

					status := algorithm.Algorithm(currentPrice, &coin, &user)

					msg := models.Message{
						User: user,
						Coin: coin,
					}

					switch status {
					case algorithm.ChangeAction:
						err = s.sStorageRepo.ResetCoin(ctx, coin, user)
						if err != nil {
							slog.ErrorContext(ctx, "Error update coin", err)
							return
						}
					case algorithm.WaitAction:
						continue
					case algorithm.BuyAction:
						updateCoin := models.NewCoin(user.Id, coin.Name)
						updateCoin.Buy = coin.Buy
						updateCoin.Count = coin.Count
						updateCoin.Decrement = coin.Decrement

						err = s.sStorageRepo.UpdateCoin(ctx, updateCoin)
						if err != nil {
							slog.ErrorContext(ctx, "Error update count", err)
							return
						}
						msg.Action = algorithm.BuyAction
						actionChanMap[userId] <- msg
					case algorithm.SellAction:
						coin.Decrement = currentPrice * user.Percent
						err = s.sStorageRepo.SellCoin(userId, coin.Name, currentPrice, coin.Decrement)
						if err != nil {
							slog.ErrorContext(ctx, "Error update SellAction", err)
							return
						}

						coin.Count = 0
						msg.Coin.CurrentPrice = currentPrice
						msg.Action = algorithm.SellAction
						actionChanMap[userId] <- msg
					}
				}
			}
		}(coin)
	}

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

	currentPrice, err := s.apiRepo.GetPrice(ctx, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting price from api", err)
		return err
	}

	coin, err := s.sStorageRepo.GetCoin(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin", err)
		return err
	}

	user, err := s.uStorageRepo.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user from algorithm", err)
		return err
	}

	spentMoney := user.Balance * user.Percent * float64(len(coin.Buy))
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
