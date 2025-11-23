package service

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/logger"
	"github.com/111zxc/pr-review-service/internal/repository"
)

type PullRequestService struct {
	prRepo       repository.PullRequestRepository
	userRepo     repository.UserRepository
	teamRepo     repository.TeamRepository
	prStatusRepo repository.PRStatusRepository
	eventsRepo   repository.EventsRepository
	rng          *rand.Rand
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
	prStatusRepo repository.PRStatusRepository,
	eventsRepo repository.EventsRepository,
) *PullRequestService {
	return &PullRequestService{
		prRepo:       prRepo,
		userRepo:     userRepo,
		teamRepo:     teamRepo,
		prStatusRepo: prStatusRepo,
		eventsRepo:   eventsRepo,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PullRequestService) CreatePullRequest(pr *domain.PullRequest) error {
	exists, err := s.prRepo.Exists(pr.ID)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrPullRequestExists
	}

	author, err := s.userRepo.GetByID(pr.AuthorID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	var authorTeam string
	if author.TeamName != "" {
		authorTeam = author.TeamName
	} else {
		teams, err := s.findUserTeams(pr.AuthorID)
		if err != nil || len(teams) == 0 {
			return domain.ErrTeamNotFound
		}
		authorTeam = teams[0].Name
	}

	reviewers, err := s.assignReviewers(authorTeam, pr.AuthorID)
	if err != nil {
		return err
	}
	pr.AssignedReviewers = reviewers

	if err := s.prRepo.Create(pr); err != nil {
		return err
	}

	eventData, err := json.Marshal(domain.PRCreatedData{
		PRName:    pr.Name,
		CreatedAt: time.Now(),
	})
	if err != nil {
		logger.Error("Failed to marshal created PR data",
			"error", err, "prname", pr.Name)
		return err
	}

	event := &domain.Event{
		EventType:      domain.EventTypePRCreated,
		PRID:           pr.ID,
		UserID:         pr.AuthorID,
		AdditionalData: eventData,
	}
	if err := s.eventsRepo.CreateEvent(event); err != nil {
		logger.Error("Failed to create PR created event",
			"error", err, "pr_id", pr.ID)
	}

	for _, reviewerID := range pr.AssignedReviewers {
		assignmentData, err := json.Marshal(domain.ReviewerAssignedData{
			AssignedAt: time.Now(),
		})
		if err != nil {
			logger.Error("Failed to marshal reviewer assignment data",
				"error", err)
			return err
		}

		assignmentEvent := &domain.Event{
			EventType:      domain.EventTypeReviewerAssigned,
			PRID:           pr.ID,
			UserID:         reviewerID,
			AdditionalData: assignmentData,
		}
		if err := s.eventsRepo.CreateEvent(assignmentEvent); err != nil {
			logger.Error("Failed to create reviewer assigned event",
				"error", err, "pr_id", pr.ID, "reviewer_id", reviewerID)
		}
	}

	return nil
}

func (s *PullRequestService) MergePullRequest(prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, err
	}

	if pr.IsMerged() {
		return pr, nil
	}

	pr.Status = domain.PRStatusMerged
	now := time.Now()
	pr.MergedAt = &now

	if err := s.prRepo.Update(pr); err != nil {
		return nil, err
	}

	mergeData, err := json.Marshal(domain.PRMergedData{
		MergedAt: time.Now(),
	})
	if err != nil {
		logger.Error("Failed to marshal PR merge data",
			"error", err)
	}

	event := &domain.Event{
		EventType:      domain.EventTypePRMerged,
		PRID:           prID,
		UserID:         pr.AuthorID,
		AdditionalData: mergeData,
	}
	if err := s.eventsRepo.CreateEvent(event); err != nil {
		logger.Error("Failed to create PR merged event",
			"error", err, "pr_id", prID)
	}

	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(prID, oldUserID string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, "", err
	}

	if pr.IsMerged() {
		return nil, "", domain.ErrPullRequestMerged
	}

	if !s.isUserAssigned(pr.AssignedReviewers, oldUserID) {
		logger.Error("Reviewer not assigned",
			"pr_id", prID,
			"old_reviewer_id", oldUserID,
			"assigned_reviewers", pr.AssignedReviewers)
		return nil, "", domain.ErrReviewerNotAssigned
	}

	newReviewer, err := s.findReplacementReviewer(prID, oldUserID)
	if err != nil {
		return nil, "", err
	}

	for i, reviewer := range pr.AssignedReviewers {
		if reviewer == oldUserID {
			pr.AssignedReviewers[i] = newReviewer
			break
		}
	}

	if err := s.prRepo.Update(pr); err != nil {
		return nil, "", err
	}

	reassignData, err := json.Marshal(domain.ReviewerReassignedData{
		OldUserID:    oldUserID,
		NewUserID:    newReviewer,
		ReassignedAt: time.Now(),
	})
	if err != nil {
		logger.Error("Failed to marshal reassigned reviewer data",
			"error", err)
	}

	event := &domain.Event{
		EventType:      domain.EventTypeReviewerReassigned,
		PRID:           prID,
		UserID:         newReviewer,
		AdditionalData: reassignData,
	}
	if err := s.eventsRepo.CreateEvent(event); err != nil {
		logger.Error("Failed to create reviewer reassigned event",
			"error", err, "pr_id", prID, "old_user_id", oldUserID, "new_user_id", newReviewer)
	}

	return pr, newReviewer, nil
}

func (s *PullRequestService) GetUserReviews(userID string) ([]*domain.PullRequestShort, error) {
	prs, err := s.prRepo.ListByReviewer(userID)
	if err != nil {
		return nil, err
	}

	var shortPRs []*domain.PullRequestShort
	for _, pr := range prs {
		shortPRs = append(shortPRs, &domain.PullRequestShort{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	return shortPRs, nil
}

func (s *PullRequestService) assignReviewers(teamName, authorID string) ([]string, error) {
	teamUsers, err := s.userRepo.GetByTeam(teamName)
	if err != nil {
		return nil, err
	}

	var candidates []string
	for _, user := range teamUsers {
		if user.IsActive && user.ID != authorID {
			candidates = append(candidates, user.ID)
		}
	}

	if len(candidates) == 0 {
		return []string{}, nil
	}

	maxReviewers := min(2, len(candidates))
	selected := make([]string, maxReviewers)

	s.rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	copy(selected, candidates[:maxReviewers])

	return selected, nil
}

func (s *PullRequestService) findUserTeams(userID string) ([]*domain.Team, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user.TeamName == "" {
		return []*domain.Team{}, nil
	}

	team, err := s.teamRepo.GetByName(user.TeamName)
	if err != nil {
		return nil, err
	}

	return []*domain.Team{team}, nil
}

func (s *PullRequestService) isUserAssigned(reviewers []string, userID string) bool {
	for _, reviewer := range reviewers {
		if reviewer == userID {
			return true
		}
	}
	return false
}

func (s *PullRequestService) findReplacementReviewer(prID, oldUserID string) (string, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return "", err
	}

	oldReviewer, err := s.userRepo.GetByID(oldUserID)
	if err != nil {
		return "", err
	}

	if oldReviewer.TeamName == "" {
		return "", domain.ErrNoCandidate
	}

	teamUsers, err := s.userRepo.GetByTeam(oldReviewer.TeamName)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, user := range teamUsers {
		if user.IsActive &&
			user.ID != pr.AuthorID &&
			user.ID != oldUserID &&
			!s.isUserAssigned(pr.AssignedReviewers, user.ID) {
			candidates = append(candidates, user.ID)
		}
	}

	if len(candidates) == 0 {
		return "", domain.ErrNoCandidate
	}

	return candidates[s.rng.Intn(len(candidates))], nil
}
