package stocks

import (
	"context"
	"encoding/json"
	"log/slog"
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

func (s *Service) GetCoinFromStockExchange(ctx context.Context, coinReq models.GetCoinRequest, apiKey, secretKey string) (models.GetCoinResponse, error) {
	coin, err := s.apiRepo.GetCoin(ctx, coinReq, apiKey, secretKey)
	if err != nil {
		return coin, err
	}
	return coin, nil
}

func (s *Service) DeleteCoin(ctx context.Context, coinTag string, userId int64) error {
	err := s.storageRepo.DeleteCoin(ctx, userId, coinTag)
	if err != nil {
		return err
	}
	return nil
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

func (s *Service) GetCoiniks(ctx context.Context, coinTag string) (models.Coiniks, error) {
	u, err := s.storageRepo.GetCoiniks(ctx, coinTag)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (s *Service) EditBuy(ctx context.Context, userId int64, buy bool) error {
	err := s.storageRepo.EditBuy(ctx, userId, buy)
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

func (s *Service) GetUserWalletBalance(ctx context.Context, apiKey, apiSecret string) (float64, error) {
	getUserWalletParams := make(models.GetUserWalletRequest)
	getUserWalletParams["accountType"] = "UNIFIED"

	getUserWalletResp, err := s.apiRepo.GetUserWalletBalance(ctx, getUserWalletParams, apiKey, apiSecret)
	if err != nil {
		slog.ErrorContext(ctx, "Error converting user USDT wallet balance to float", err)
		return 0, err
	}

	userUSDTBalance, err := strconv.ParseFloat(getUserWalletResp.Result.List[0].TotalEquity, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting user wallet balance", err)
		return 0, err
	}

	return userUSDTBalance, nil
}
