package algorithm

import (
	"context"
	"errors"
	"fmt"
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

	for _, coin := range coinList {
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
					fmt.Scan(&currentPrice)
					if err != nil {
						slog.ErrorContext(ctx, "Error getting price from algorithm", err)
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

					fmt.Println(coin.Name, coin.EntryPrice, coin.Count)

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
						err = s.sStoRepo.SellAction(userId, coin.Name, currentPrice, coin.Decrement)
						if err != nil {
							slog.ErrorContext(ctx, "Error update SellAction", err)
							return
						}

						msg.Action = algorithm.SellAction
						actionChanMap[userId] <- msg
					}
				}
			}
		}(coin)
	}
	return nil
}

func (s *Service) StopTradingCoin(ctx context.Context, userId int64, coin string) error {
	if _, ok := s.stopCoinMap[coin][userId]; !ok {
		return errors.New("coin does not exist")
	}
	s.stopCoinMap[coin][userId] <- struct{}{}
	return nil
}
