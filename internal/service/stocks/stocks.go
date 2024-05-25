package stocks

import (
	"context"
	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
)

type Service struct {
	stockRepo   apiStock.Repository
	storageRepo storageStock.Repository
}

func New(stockRepo apiStock.Repository, storageRepo storageStock.Repository) *Service {
	return &Service{stockRepo: stockRepo, storageRepo: storageRepo}
}

func (s *Service) GetCoinList(ctx context.Context, userId int64) ([]string, error) {
	list, err := s.storageRepo.GetCoinList(ctx, userId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) ExistCoin(ctx context.Context, coinTag string) (bool, error) {
	list, err := s.stockRepo.ExistCoin(ctx, coinTag)
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
