package stocks

import (
	"context"
	"fmt"
	"m1pes/internal/algorithm"
	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
	"time"
)

type Service struct {
	apiRepo     apiStock.Repository
	storageRepo storageStock.Repository
	stopCoinMap map[string]map[int64]chan struct{}
}

func New(stockRepo apiStock.Repository, storageRepo storageStock.Repository) *Service {
	return &Service{apiRepo: stockRepo, storageRepo: storageRepo, stopCoinMap: make(map[string]map[int64]chan struct{})}
}

func (s *Service) StopTradingCoin(_ context.Context, userId int64, coin string) error {
	s.stopCoinMap[coin][userId] <- struct{}{}
	return nil
}

func (s *Service) StartTrading(ctx context.Context, userId int64) error {
	coinList, err := s.storageRepo.GetCoinList(ctx, userId)
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
					currentPrice, err := s.apiRepo.GetPrice(ctx, coin)
					if err != nil {
						return
					}

					algorithm.Algorithm()
					time.Sleep(time.Millisecond * 500)
					fmt.Println(funcCoin)
				}
			}
		}(coin)
	}
	return nil
}

func (s *Service) GetCoinList(ctx context.Context, userId int64) ([]string, error) {
	list, err := s.storageRepo.GetCoinList(ctx, userId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) ExistCoin(ctx context.Context, coinTag string) (bool, error) {
	list, err := s.apiRepo.ExistCoin(ctx, coinTag)
	if err != nil {
		return false, err
	}
	return list, nil
}

func (s *Service) AddCoin(userId int64, coinTag string) error {
	err := s.storageRepo.AddCoin(userId, coinTag)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) CheckStatus(userId int64) (string, error) {
	status, err := s.storageRepo.CheckStatus(userId)
	if err != nil {
		return "", err
	}
	return status, nil
}
func (s *Service) UpdateStatus(userID int64, status string) error {
	err := s.storageRepo.UpdateStatus(userID, status)
	if err != nil {
		return err
	}
	return nil
}
func (s *Service) UpdatePercent(userID int64, percent float64) error {
	err := s.storageRepo.UpdatePercent(userID, percent)
	if err != nil {
		return err
	}
	return nil
}
