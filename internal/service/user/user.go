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

func (s *Service) UpdateUser(ctx context.Context, user models.User) error {
	err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetAllUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) GetIncomeLastDay(ctx context.Context, userID int64) (float64, error) {
	income, err := s.userRepo.GetIncome(ctx, userID)
	if err != nil {
		return 0, err
	}
	return income, nil
}

func (s *Service) NewUser(ctx context.Context, user models.User) error {
	err := s.userRepo.NewUser(ctx, user)
	if err != nil {
		return logging.WrapError(ctx, err)
	}
	return nil
}

func (s *Service) GetUser(ctx context.Context, userId int64) (models.User, error) {
	u, err := s.userRepo.GetUser(ctx, userId)
	if err != nil {
		return models.User{}, logging.WrapError(ctx, err)
	}
	return u, nil
}
