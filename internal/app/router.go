package app

import (
	"net/http"

	"github.com/111zxc/pr-review-service/internal/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", h.Team.CreateTeam)
	mux.HandleFunc("/team/get", h.Team.GetTeam)

	mux.HandleFunc("/users/setIsActive", h.User.SetUserActive)
	mux.HandleFunc("/users/getReview", h.User.GetUserReviews)

	mux.HandleFunc("/pullRequest/create", h.PR.CreatePullRequest)
	mux.HandleFunc("/pullRequest/merge", h.PR.MergePullRequest)
	mux.HandleFunc("/pullRequest/reassign", h.PR.ReassignReviewer)

	mux.HandleFunc("/stats", h.Stats.GetStats)

	mux.HandleFunc("/health", h.Health.Health)

	return mux
}
