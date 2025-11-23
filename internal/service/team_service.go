package service

import (
	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/repository"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(team *domain.Team) error {
	exists, err := s.teamRepo.Exists(team.Name)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrTeamExists
	}

	for _, member := range team.Members {
		user := &domain.User{
			ID:       member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}

		if err := s.userRepo.Create(user); err != nil {
			return err
		}
	}

	return s.teamRepo.Create(team)
}

func (s *TeamService) GetTeam(name string) (*domain.Team, error) {
	return s.teamRepo.GetByName(name)
}
