package algorithm

import (
	"context"
	"errors"
	"log/slog"

	"m1pes/internal/models"

	"m1pes/internal/algorithm"
	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
	storageUser "m1pes/internal/repository/storage/user"
)

type Service struct {
	apiRepo     apiStock.Repository
	sStoRepo    storageStock.Repository
	uStoRepo    storageUser.Repository
	stopCoinMap map[string]map[int64]chan struct{}
}

func New(apiRepo apiStock.Repository, sStoRepo storageStock.Repository, uStoRepo storageUser.Repository) *Service {
	return &Service{apiRepo, sStoRepo, uStoRepo, make(map[string]map[int64]chan struct{})}
}

func (s *Service) StartTrading(ctx context.Context, userId int64, actionChanMap map[int64]chan models.Message) error {
	coinList, err := s.sStoRepo.GetCoinList(ctx, userId)
	if err != nil {
		return err
	}

	for _, coin := range coinList.Name {
		// init map that stores coin name as key and map2 as value
		// map2 stores userId as key and struct{} as value
		if _, ok := s.stopCoinMap[coin][userId]; ok {
			continue
		}
		s.stopCoinMap[coin] = make(map[int64]chan struct{})
		s.stopCoinMap[coin][userId] = make(chan struct{})
		go func(funcCoin string) {
			for {
				select {
				case <-s.stopCoinMap[funcCoin][userId]:
					delete(s.stopCoinMap[funcCoin], userId)
					return
				default:
					currentPrice, err := s.apiRepo.GetPrice(ctx, funcCoin)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting price from api", err)
						return
					}

					user, err := s.uStoRepo.GetUser(ctx, userId)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting user from algorithm", err)
						return
					}

					coin, err := s.sStoRepo.GetCoin(ctx, userId, funcCoin)
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
						err = s.sStoRepo.UpdateCoin(userId, coin.Name, coin.EntryPrice, user.Percent)
						if err != nil {
							slog.ErrorContext(ctx, "Error update coin", err)
							return
						}
					case algorithm.WaitAction:
						continue
					case algorithm.BuyAction:
						err = s.sStoRepo.UpdateCount(userId, coin.Count, coin.Name, coin.Decrement, coin.Buy)
						if err != nil {
							slog.ErrorContext(ctx, "Error update count", err)
							return
						}
						msg.Action = algorithm.BuyAction
						actionChanMap[userId] <- msg
					case algorithm.SellAction:
						coin.Decrement = currentPrice * user.Percent
						err = s.sStoRepo.SellCoin(userId, coin.Name, currentPrice, coin.Decrement)
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
	return nil
}

func (s *Service) DeleteCoin(ctx context.Context, userId int64, coinTag string) error {
	if _, ok := s.stopCoinMap[coinTag][userId]; !ok {
		return errors.New("coin does not exist")
	}
	s.stopCoinMap[coinTag][userId] <- struct{}{}

	currentPrice, err := s.apiRepo.GetPrice(ctx, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting price from api", err)
		return err
	}

	coin, err := s.sStoRepo.GetCoin(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting coin", err)
		return err
	}

	user, err := s.uStoRepo.GetUser(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user from algorithm", err)
		return err
	}

	spentMoney := user.Balance * user.Percent * float64(len(coin.Buy))
	earnMoney := currentPrice * coin.Count

	income := earnMoney - spentMoney

	err = s.sStoRepo.DeleteCoin(ctx, userId, coinTag)
	if err != nil {
		slog.ErrorContext(ctx, "Error delete coin", err)
		return err
	}

	err = s.uStoRepo.ChangeBalance(ctx, userId, income)
	if err != nil {
		slog.ErrorContext(ctx, "Error change balance", err)
		return err
	}

	err = s.sStoRepo.InsertIncome(userId, coinTag, income, coin.Count)
	if err != nil {
		slog.ErrorContext(ctx, "Error insert income", err)
		return err
	}

	return nil
}
