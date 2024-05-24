package user

import "m1pes/internal/models"

type Repository interface {
	NewUser(user models.User) error
}
