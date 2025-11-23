package handler

import (
	"encoding/json"
	"net/http"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/handler/dto"
	"github.com/111zxc/pr-review-service/internal/logger"
	"github.com/111zxc/pr-review-service/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	prService   *service.PullRequestService
}

func NewUserHandler(userService *service.UserService, prService *service.PullRequestService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
	}
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	var req dto.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", "error", err)
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	user, err := h.userService.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			writeError(w, domain.NewErrorResponse("NOT_FOUND", "resource not found"))
		default:
			logger.Error("Failed to set user active", "error", err)
			writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"user": dto.UserResponse{
			UserID:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	})
	if err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "user_id is required"))
		return
	}

	prs, err := h.prService.GetUserReviews(userID)
	if err != nil {
		logger.Error("Failed to get user reviews", "error", err)
		writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		return
	}

	var shortPRs []dto.PullRequestShortResponse
	for _, pr := range prs {
		shortPRs = append(shortPRs, dto.PullRequestShortResponse{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(dto.UserReviewsResponse{
		UserID:       userID,
		PullRequests: shortPRs,
	})
	if err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}
