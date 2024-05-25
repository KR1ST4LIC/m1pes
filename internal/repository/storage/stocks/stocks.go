package stocks

import "context"

type Repository interface {
	GetCoinList(ctx context.Context, userId int64) ([]string, error)
}
