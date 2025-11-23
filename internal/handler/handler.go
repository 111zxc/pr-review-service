package handler

import (
	"github.com/111zxc/pr-review-service/internal/service"
)

type Handler struct {
	Team   *TeamHandler
	User   *UserHandler
	PR     *PullRequestHandler
	Stats  *StatsHandler
	Health *HealthHandler
}

func New(team *service.TeamService, user *service.UserService, pr *service.PullRequestService, stats *service.StatsService) *Handler {
	return &Handler{
		Team:   NewTeamHandler(team),
		User:   NewUserHandler(user, pr),
		PR:     NewPullRequestHandler(pr),
		Stats:  NewStatsHandler(stats),
		Health: NewHealthHandler(),
	}
}
