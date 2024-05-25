package user

import (
	"context"
	"m1pes/internal/logging"
	"m1pes/internal/models"
	"m1pes/internal/repository/storage/user"
)

type Service struct {
	userRepo user.Repository
}

func New(userRepo user.Repository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) NewUser(ctx context.Context, user models.User) error {
	err := s.userRepo.NewUser(ctx, user)
	if err != nil {
		return logging.WrapError(ctx, err)
	}
	return nil
}
