package stocks

import (
	"context"
)

type Repository interface {
	GetPrice(ctx context.Context, coinTag, apiKey string) (float64, error)
	ExistCoin(ctx context.Context, coinTag, apiKey string) (bool, error)
	CreateSignRequestAndGetRespBody(params, endPoint, method, apiKey, apiSecret string) ([]byte, error)
}
