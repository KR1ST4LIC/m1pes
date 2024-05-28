package user

import (
	"context"

	"m1pes/internal/models"
)

type Repository interface {
	NewUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, userId int64) (models.User, error)
	ChangeBalance(ctx context.Context, userId int64, amount float64) error
}
