package stocks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"m1pes/internal/models"
	apiStock "m1pes/internal/repository/api/stocks"
	storageStock "m1pes/internal/repository/storage/stocks"
)

type CreateOrderResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	ExtCode string `json:"extCode"`
	ExtInfo string `json:"extInfo"`
	Result  struct {
		OrderID     string `json:"orderId"`
		OrderLinkID string `json:"orderLinkId"`
	} `json:"result"`
	TimeNow string `json:"timeNow"`
}

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
	list, err := s.storageRepo.ExistCoin(ctx, coinTag)
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

func (s *Service) CreateOrder(apiKey, apiSecret string, order models.OrderCreate) (string, error) {
	resp := models.CreateOrderResponse{}
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
		return "", err
	}

	data, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), "/v5/order/create", "POST", apiKey, apiSecret)
	if err != nil {
		return "", err
	}

	json.Unmarshal(data, &resp)
	return resp.Result.OrderID, nil
}

func (s *Service) GetBalanceFromBybit(apiKey, apiSecret string) (float64, error) {
	resp := models.Response{}
	params := map[string]interface{}{
		"accountType": "UNIFIED",
	}
	jsonData, err := json.Marshal(params)
	if err != nil {
		fmt.Println(err)
	}
	data, err := s.apiRepo.CreateSignRequestAndGetRespBody(string(jsonData), "/v5/account/wallet-balance", "GET", apiKey, apiSecret)
	if err != nil {
		return 0, err
	}
	json.Unmarshal(data, &resp)
	if resp.RetMsg == "OK" {
		if resp.Result.List[0].TotalEquity == "" {
			return 0, nil
		}
	} else {
		return 0, err
	}
	a, err := strconv.ParseFloat(resp.Result.List[0].TotalEquity, 64)
	if err != nil {
		return 0, err
	}
	return a, nil
}
