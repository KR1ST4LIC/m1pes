package stocks

import (
	"context"

	"m1pes/internal/models"
	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
)

type Service struct {
	apiRepo     apiStock.Repository
	storageRepo storageStock.Repository
	stopCoinMap map[string]map[int64]chan struct{}
}

func New(stockRepo apiStock.Repository, storageRepo storageStock.Repository) *Service {
	return &Service{apiRepo: stockRepo, storageRepo: storageRepo, stopCoinMap: make(map[string]map[int64]chan struct{})}
}

func (s *Service) GetCoin(ctx context.Context, userId int64, coin string) {

}

func (s *Service) GetCoinList(ctx context.Context, userId int64) ([]models.Coin, error) {
	list, err := s.storageRepo.GetCoinList(ctx, userId)
	if err != nil {
		return list, err
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

func (s *Service) AddCoin(coin models.Coin) error {
	err := s.storageRepo.AddCoin(coin)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) InsertIncome(userID int64, coinTag string, income, count float64) error {
	err := s.storageRepo.InsertIncome(userID, coinTag, income, count)
	if err != nil {
		return err
	}
	return nil
}
