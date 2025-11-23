package handler

import (
	"encoding/json"
	"net/http"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/handler/dto"
	"github.com/111zxc/pr-review-service/internal/logger"
	"github.com/111zxc/pr-review-service/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", "error", err)
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	team := &domain.Team{
		Name: req.Name,
	}

	for _, member := range req.Members {
		team.Members = append(team.Members, domain.TeamMember{
			UserID:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		})
	}

	if err := h.teamService.CreateTeam(team); err != nil {
		switch err {
		case domain.ErrTeamExists:
			writeError(w, domain.NewErrorResponse("TEAM_EXISTS", "team_name already exists"))
		default:
			logger.Error("Failed to create team", "error", err)
			writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"team": team,
	})
	if err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, domain.NewErrorResponse("INVALID_INPUT", "team_name is required"))
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		switch err {
		case domain.ErrTeamNotFound:
			writeError(w, domain.NewErrorResponse("NOT_FOUND", "resource not found"))
		default:
			logger.Error("Failed to get team", "error", err)
			writeError(w, domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(team); err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}

func writeError(w http.ResponseWriter, errResp domain.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")

	switch errResp.Error.Code {
	case "TEAM_EXISTS":
		w.WriteHeader(http.StatusBadRequest)
	case "NOT_FOUND":
		w.WriteHeader(http.StatusNotFound)
	case "INVALID_INPUT":
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("failed to write JSON response", "error", err)
		return
	}
}
