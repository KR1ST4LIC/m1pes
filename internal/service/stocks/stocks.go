package stocks

import (
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

func (s *Service) GetCoinList() ([]string, error) {
	list, err := s.storageRepo.GetCoinList()
	if err != nil {
		return nil, err
	}
	return list, nil
}
