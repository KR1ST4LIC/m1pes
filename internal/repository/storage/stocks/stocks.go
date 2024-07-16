package stocks

import (
	"context"

	"m1pes/internal/models"
)

type Repository interface {
	GetCoin(ctx context.Context, userId int64, coin string) (models.Coin, error)
	GetCoiniks(ctx context.Context, coinName string) (models.Coiniks, error)
	EditBuy(ctx context.Context, userId int64, buy bool) error
	ExistCoin(ctx context.Context, coinTag string) (bool, error)
	GetCoinList(ctx context.Context, userId int64) ([]models.Coin, error)
	AddCoin(coin models.Coin) error
	UpdateCoin(ctx context.Context, coin models.Coin) error
	ResetCoin(ctx context.Context, coin models.Coin, user models.User) error
	UpdateCount(userID int64, count float64, coinTag string, decrement float64, buy []float64) error
	SellCoin(userID int64, coinTag string, sellPrice float64) error
	SetCoinToDefault(ctx context.Context, userId int64, coinTag string) error
	DeleteCoin(ctx context.Context, userID int64, coinTag string) error
	InsertIncome(userID int64, coinTag string, income, count float64) error
}
