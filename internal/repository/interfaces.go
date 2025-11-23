package repository

import (
	"github.com/111zxc/pr-review-service/internal/domain"
)

type (
	UserRepository interface {
		Create(user *domain.User) error
		GetByID(id string) (*domain.User, error)
		Update(user *domain.User) error
		GetByTeam(teamName string) ([]*domain.User, error)
	}

	TeamRepository interface {
		Create(team *domain.Team) error
		GetByName(name string) (*domain.Team, error)
		Exists(name string) (bool, error)
	}

	PullRequestRepository interface {
		Create(pr *domain.PullRequest) error
		GetByID(id string) (*domain.PullRequest, error)
		Update(pr *domain.PullRequest) error
		ListByReviewer(userID string) ([]*domain.PullRequest, error)
		Exists(id string) (bool, error)
	}

	PRStatusRepository interface {
		GetByCode(code string) (*domain.PRStatus, error)
		GetByID(id int) (*domain.PRStatus, error)
		ListAll() ([]*domain.PRStatus, error)
	}

	EventsRepository interface {
		CreateEvent(event *domain.Event) error
		GetEventsByType(eventType domain.EventType, limit int) ([]domain.Event, error)
		GetEventCountsByType() (map[domain.EventType]int, error)
	}

	StatsRepository interface {
		GetEventStats() (*domain.StatsResponse, error)
	}
)
