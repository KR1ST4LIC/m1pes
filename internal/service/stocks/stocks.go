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
