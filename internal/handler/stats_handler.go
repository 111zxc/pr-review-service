package handler

import (
	"encoding/json"
	"net/http"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/logger"
	"github.com/111zxc/pr-review-service/internal/service"
)

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := json.NewEncoder(w).Encode(domain.ErrorResponse{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "Only GET method is allowed",
			},
		})
		if err != nil {
			logger.Error("failed to write JSON response", "error", err)
			return
		}
		return
	}

	stats, err := h.statsService.GetStats()
	if err != nil {
		logger.Error("Failed to get stats", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(domain.ErrorResponse{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to get statistics",
			},
		})
		if err != nil {
			logger.Error("failed to write JSON response", "error", err)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(stats); err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}
