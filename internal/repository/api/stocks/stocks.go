package stocks

import (
	"context"
	"m1pes/internal/models"
)

type Repository interface {
	CreateOrder(ctx context.Context, orderReq models.CreateOrderRequest, apiKey, secretKey string) (models.CreateOrderResponse, error)
	CancelOrder(ctx context.Context, orderReq models.CancelOrderRequest, apiKey, secretKey string) (models.CancelOrderResponse, error)
	GetOrder(ctx context.Context, orderReq models.GetOrderRequest, apiKey, secretKey string) (models.GetOrderResponse, error)
	GetCoin(ctx context.Context, coinReq models.GetCoinRequest, apiKey, secretKey string) (models.GetCoinResponse, error)
	GetUserWalletBalance(ctx context.Context, req models.GetUserWalletRequest, apiKey, secretKey string) (models.GetUserWalletResponse, error)
	CreateSignRequestAndGetRespBody(params, endPoint, method, apiKey, apiSecret string) ([]byte, error)
}
