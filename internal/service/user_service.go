package service

import (
	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) SetUserActive(userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	user.IsActive = isActive

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
