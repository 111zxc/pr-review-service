package handler

import (
	"encoding/json"
	"net/http"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/handler/dto"
	"github.com/111zxc/pr-review-service/internal/logger"
	"github.com/111zxc/pr-review-service/internal/service"
)

type PullRequestHandler struct {
	prService *service.PullRequestService
}

func NewPullRequestHandler(prService *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prService: prService}
}

func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", "error", err)
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	pr := &domain.PullRequest{
		ID:       req.ID,
		Name:     req.Name,
		AuthorID: req.AuthorID,
	}

	if err := h.prService.CreatePullRequest(pr); err != nil {
		switch err {
		case domain.ErrPullRequestExists:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			writeError(w, domain.NewErrorResponse("PR_EXISTS", "PR id already exists"))
		case domain.ErrUserNotFound, domain.ErrTeamNotFound:
			writeError(w, domain.NewErrorResponse("NOT_FOUND", "resource not found"))
		default:
			logger.Error("Failed to create PR", "error", err)
			writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": dto.PullRequestResponse{
			ID:                pr.ID,
			Name:              pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            pr.Status,
			AssignedReviewers: pr.AssignedReviewers,
			CreatedAt:         pr.CreatedAt,
		},
	})
	if err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}

func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	var req dto.MergePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", "error", err)
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	pr, err := h.prService.MergePullRequest(req.ID)
	if err != nil {
		switch err {
		case domain.ErrPullRequestNotFound:
			writeError(w, domain.NewErrorResponse("NOT_FOUND", "resource not found"))
		default:
			logger.Error("Failed to merge PR", "error", err)
			writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": dto.PullRequestResponse{
			ID:                pr.ID,
			Name:              pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            pr.Status,
			AssignedReviewers: pr.AssignedReviewers,
			CreatedAt:         pr.CreatedAt,
			MergedAt:          pr.MergedAt,
		},
	})
	if err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}

func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req dto.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", "error", err)
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		switch err {
		case domain.ErrPullRequestNotFound:
			writeError(w, domain.NewErrorResponse("NOT_FOUND", "resource not found"))
		case domain.ErrPullRequestMerged:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			writeError(w, domain.NewErrorResponse("PR_MERGED", "cannot reassign on merged PR"))
		case domain.ErrReviewerNotAssigned:
			writeError(w, domain.NewErrorResponse("NOT_FOUND", "resource not found"))
		case domain.ErrNoCandidate:
			writeError(w, domain.NewErrorResponse("NO_CANDIDATE", "no active replacement candidate in team"))
		default:
			logger.Error("Failed to reassign reviewer", "error", err)
			writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(dto.ReassignResponse{
		PR: dto.PullRequestResponse{
			ID:                pr.ID,
			Name:              pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            pr.Status,
			AssignedReviewers: pr.AssignedReviewers,
			CreatedAt:         pr.CreatedAt,
			MergedAt:          pr.MergedAt,
		},
		ReplacedBy: replacedBy,
	})
	if err != nil {
		logger.Error("failed to write json response", "error", err)
		return
	}
}
