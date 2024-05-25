package user

import (
	"context"
	"m1pes/internal/models"
)

type Repository interface {
	NewUser(ctx context.Context, user models.User) error
}
