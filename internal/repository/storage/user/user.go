package user

import (
	"context"
	"m1pes/internal/models"
)

type Repository interface {
	NewUser(ctx context.Context, user models.User) error
	IncrementBalance(ctx context.Context, userId, amount int64) error
}
