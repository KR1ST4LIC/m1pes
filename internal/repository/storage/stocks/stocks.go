package stocks

import (
	"context"

	"m1pes/internal/models"
)

type Repository interface {
	GetCoin(ctx context.Context, userId int64, coin string) (models.Coin, error)
	GetCoinList(ctx context.Context, userId int64) ([]string, error)
	AddCoin(coin models.Coin) error
	CheckStatus(userId int64) (string, error)
	UpdateStatus(userID int64, status string) error
	UpdatePercent(userID int64, percent float64) error
	UpdateCoin(userID int64, coinTag string, entryPrice, percent float64) error
	UpdateCount(userID, count int64, coinTag string, decrement float64, buy []float64) error
	SellAction(userID int64, coinTag string, currentPrice, decrement float64) error
}
