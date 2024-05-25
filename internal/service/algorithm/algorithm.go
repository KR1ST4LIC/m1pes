package algorithm

import (
	"context"
	"fmt"
	"log/slog"

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

func (s *Service) StartTrading(ctx context.Context, userId int64) error {
	coinList, err := s.sStoRepo.GetCoinList(ctx, userId)
	if err != nil {
		return err
	}

	for _, coin := range coinList {
		// init map that stores coin name as key and map2 as value
		// map2 stores userId as key and struct{} as value
		s.stopCoinMap[coin] = make(map[int64]chan struct{})
		s.stopCoinMap[coin][userId] = make(chan struct{})
		go func(funcCoin string) {
			for {
				select {
				case <-s.stopCoinMap[funcCoin][userId]:
					return
				default:
					// here is code for algorithm
					currentPrice, err := s.apiRepo.GetPrice(ctx, funcCoin)
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

					change := algorithm.Algorithm(currentPrice, &coin, &user)

					fmt.Println(coin.Name, currentPrice, coin.Count)

					if change {
						err = s.sStoRepo.UpdateCoin(userId, coin.Name, coin.EntryPrice)
					}
					err = s.sStoRepo.AddCoin(coin)
					if err != nil {
						slog.ErrorContext(ctx, "Error adding coin", err)
						return
					}

					//time.Sleep(time.Millisecond * 500)
					//fmt.Println(funcCoin)
				}
			}
		}(coin)
	}
	return nil
}

func (s *Service) StopTradingCoin(ctx context.Context, userId int64, coin string) error {
	s.stopCoinMap[coin][userId] <- struct{}{}
	return nil
}
