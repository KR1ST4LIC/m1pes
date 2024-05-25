package stocks

import "context"

type Repository interface {
	GetPrice(ctx context.Context, coinTag string) (float64, error)
	ExistCoin(ctx context.Context, coinTag string) (bool, error)
}
