package user

import (
	"m1pes/internal/models"
	"m1pes/internal/repository/storage/user"
)

type Service struct {
	userRepo user.Repository
}

func New(userRepo user.Repository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) NewUser(user models.User) error {
	err := s.userRepo.NewUser(user)
	if err != nil {
		return err
	}
	return nil
}
