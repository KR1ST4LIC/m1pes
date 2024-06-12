package stocks

import (
	"context"
)

type Repository interface {
	GetPrice(ctx context.Context, coinTag, apiKey string) (float64, error)
	CreateSignRequestAndGetRespBody(params, endPoint, method, apiKey, apiSecret string) ([]byte, error)
}
