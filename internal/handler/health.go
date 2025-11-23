package handler

import (
	"net/http"
	"time"

	"github.com/111zxc/pr-review-service/internal/logger"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Health check called", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	if err != nil {
		logger.Error("couldn't write health response", "error", err)
	}
}
