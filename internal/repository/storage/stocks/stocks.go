package stocks

import (
	"context"

	"m1pes/internal/models"
)

type Repository interface {
	GetCoin(ctx context.Context, userId int64, coin string) (models.Coin, error)
	GetCoinList(ctx context.Context, userId int64) (models.List, error)
	AddCoin(coin models.Coin) error
	UpdatePercent(userID int64, percent float64) error
	UpdateCoin(userID int64, coinTag string, entryPrice, percent float64) error
	UpdateCount(userID int64, count float64, coinTag string, decrement float64, buy []float64) error
	SellCoin(userID int64, coinTag string, currentPrice, decrement float64) error
	DeleteCoin(ctx context.Context, userID int64, coinTag string) error
	InsertIncome(userID int64, coinTag string, income, count float64) error
}
