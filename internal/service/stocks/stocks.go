package stocks

import (
	"context"
	"encoding/json"

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

func (s *Service) ExistCoin(ctx context.Context, coinTag, apiKey string) (bool, error) {
	list, err := s.apiRepo.ExistCoin(ctx, coinTag, apiKey)
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

func (s *Service) CreateOrder(endPoint, apiKey, apiSecret string, order models.OrderCreate) ([]byte, error) {
	postParams := map[string]interface{}{
		"category":    "spot",
		"symbol":      order.Symbol,
		"side":        order.Side,
		"positionIdx": 0,
		"orderType":   "Limit",
		"qty":         order.Qty,
		"price":       order.Price,
		"timeInForce": "GTC",
	}
	jsonData, err := json.Marshal(postParams)
	if err != nil {
		return nil, err
	}
	data, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), endPoint, "", apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetBalanceFromBybit(endPoint, apiKey, apiSecret string, order models.OrderCreate) (float64, error) {
	postParams := map[string]interface{}{
		"category":    "spot",
		"symbol":      order.Symbol,
		"side":        order.Side,
		"positionIdx": 0,
		"orderType":   "Limit",
		"qty":         order.Qty,
		"price":       order.Price,
		"timeInForce": "GTC",
	}
	jsonData, err := json.Marshal(postParams)
	if err != nil {
		return 0, err
	}
	_, err = s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), endPoint, "/v5/account/wallet-balance", apiKey, apiSecret)
	if err != nil {
		return 0, err
	}
	return 0, nil
}
