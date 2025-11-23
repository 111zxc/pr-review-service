package unit

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/repository/mocks"
	"github.com/111zxc/pr-review-service/internal/service"
)

type PRServiceTestSuite struct {
	mockPRRepo       *mocks.PullRequestRepository
	mockUserRepo     *mocks.UserRepository
	mockTeamRepo     *mocks.TeamRepository
	mockPRStatusRepo *mocks.PRStatusRepository
	mockEventsRepo   *mocks.EventsRepository
	prService        *service.PullRequestService
}

func NewPRServiceTestSuite() *PRServiceTestSuite {
	mockPRRepo := new(mocks.PullRequestRepository)
	mockUserRepo := new(mocks.UserRepository)
	mockTeamRepo := new(mocks.TeamRepository)
	mockPRStatusRepo := new(mocks.PRStatusRepository)
	mockEventsRepo := new(mocks.EventsRepository)

	prService := service.NewPullRequestService(
		mockPRRepo, mockUserRepo, mockTeamRepo, mockPRStatusRepo, mockEventsRepo,
	)

	return &PRServiceTestSuite{
		mockPRRepo:       mockPRRepo,
		mockUserRepo:     mockUserRepo,
		mockTeamRepo:     mockTeamRepo,
		mockPRStatusRepo: mockPRStatusRepo,
		mockEventsRepo:   mockEventsRepo,
		prService:        prService,
	}
}

func CreateTestPullRequest() *domain.PullRequest {
	return &domain.PullRequest{
		ID:       "pr-1",
		Name:     "Test PR",
		AuthorID: "u1",
		Status:   domain.PRStatusOpen,
	}
}

func TestPullRequestService_CreatePullRequest_Success(t *testing.T) {
	suite := NewPRServiceTestSuite()
	pr := CreateTestPullRequest()

	author := CreateTestUser()

	teamUsers := []*domain.User{
		{ID: "u2", Username: "Bob", IsActive: true, TeamName: "backend"},
		{ID: "u3", Username: "Charlie", IsActive: true, TeamName: "backend"},
		{ID: "u4", Username: "David", IsActive: true, TeamName: "backend"},
	}

	suite.mockPRRepo.On("Exists", "pr-1").Return(false, nil)
	suite.mockUserRepo.On("GetByID", "u1").Return(author, nil)
	suite.mockUserRepo.On("GetByTeam", "backend").Return(teamUsers, nil)
	suite.mockPRRepo.On("Create", pr).Return(nil)
	suite.mockEventsRepo.On("CreateEvent", mock.AnythingOfType("*domain.Event")).Return(nil)

	err := suite.prService.CreatePullRequest(pr)

	assert.NoError(t, err)
	assert.Len(t, pr.AssignedReviewers, 2)

	assert.NotContains(t, pr.AssignedReviewers, "u1")

	suite.mockPRRepo.AssertExpectations(t)
	suite.mockUserRepo.AssertExpectations(t)
	suite.mockTeamRepo.AssertExpectations(t)
}

func TestPullRequestService_CreatePullRequest_AlreadyExists(t *testing.T) {
	suite := NewPRServiceTestSuite()
	pr := CreateTestPullRequest()

	suite.mockPRRepo.On("Exists", "pr-1").Return(true, nil)

	err := suite.prService.CreatePullRequest(pr)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrPullRequestExists, err)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_CreatePullRequest_AuthorNotFound(t *testing.T) {
	suite := NewPRServiceTestSuite()
	pr := CreateTestPullRequest()

	suite.mockPRRepo.On("Exists", "pr-1").Return(false, nil)
	suite.mockUserRepo.On("GetByID", "u1").Return(nil, domain.ErrUserNotFound)

	err := suite.prService.CreatePullRequest(pr)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
	suite.mockPRRepo.AssertExpectations(t)
	suite.mockUserRepo.AssertExpectations(t)
}

func TestPullRequestService_CreatePullRequest_TeamNotFound(t *testing.T) {
	suite := NewPRServiceTestSuite()
	pr := CreateTestPullRequest()

	author := &domain.User{
		ID:       "u1",
		Username: "Alice",
		IsActive: true,
	}

	suite.mockPRRepo.On("Exists", "pr-1").Return(false, nil)
	suite.mockUserRepo.On("GetByID", "u1").Return(author, nil)
	suite.mockTeamRepo.On("GetByName", mock.Anything).Return(nil, domain.ErrTeamNotFound)

	err := suite.prService.CreatePullRequest(pr)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrTeamNotFound, err)
	suite.mockPRRepo.AssertExpectations(t)
	suite.mockUserRepo.AssertExpectations(t)
}

func TestPullRequestService_MergePullRequest_Success(t *testing.T) {
	suite := NewPRServiceTestSuite()

	pr := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	suite.mockPRRepo.On("GetByID", "pr-1").Return(pr, nil)
	suite.mockPRRepo.On("Update", mock.MatchedBy(func(p *domain.PullRequest) bool {
		return p.Status == domain.PRStatusMerged && p.MergedAt != nil
	})).Return(nil)
	suite.mockEventsRepo.On("CreateEvent", mock.AnythingOfType("*domain.Event")).Return(nil)

	result, err := suite.prService.MergePullRequest("pr-1")

	assert.NoError(t, err)
	assert.Equal(t, domain.PRStatusMerged, result.Status)
	assert.NotNil(t, result.MergedAt)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_MergePullRequest_AlreadyMerged(t *testing.T) {
	suite := NewPRServiceTestSuite()

	mergedAt := time.Now()
	pr := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            domain.PRStatusMerged,
		AssignedReviewers: []string{"u2", "u3"},
		MergedAt:          &mergedAt,
	}

	suite.mockPRRepo.On("GetByID", "pr-1").Return(pr, nil)

	result, err := suite.prService.MergePullRequest("pr-1")

	assert.NoError(t, err)
	assert.Equal(t, domain.PRStatusMerged, result.Status)
	assert.Equal(t, &mergedAt, result.MergedAt)
	suite.mockPRRepo.AssertExpectations(t)
	suite.mockPRRepo.AssertNotCalled(t, "Update")
}

func TestPullRequestService_MergePullRequest_NotFound(t *testing.T) {
	suite := NewPRServiceTestSuite()

	suite.mockPRRepo.On("GetByID", "pr-1").Return(nil, domain.ErrPullRequestNotFound)

	result, err := suite.prService.MergePullRequest("pr-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrPullRequestNotFound, err)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_ReassignReviewer_Success(t *testing.T) {
	suite := NewPRServiceTestSuite()

	pr := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	oldReviewer := &domain.User{
		ID:       "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	teamUsers := []*domain.User{
		{ID: "u3", Username: "Ivan", IsActive: true, TeamName: "backend"},
		{ID: "u4", Username: "Stepan", IsActive: true, TeamName: "backend"},
	}

	suite.mockPRRepo.On("GetByID", "pr-1").Return(pr, nil)
	suite.mockUserRepo.On("GetByID", "u2").Return(oldReviewer, nil)
	suite.mockUserRepo.On("GetByTeam", "backend").Return(teamUsers, nil)
	suite.mockPRRepo.On("Update", mock.MatchedBy(func(p *domain.PullRequest) bool {
		return slices.Contains(p.AssignedReviewers, "u4") && !slices.Contains(p.AssignedReviewers, "u2")
	})).Return(nil)
	suite.mockEventsRepo.On("CreateEvent", mock.AnythingOfType("*domain.Event")).Return(nil)

	result, newReviewer, err := suite.prService.ReassignReviewer("pr-1", "u2")

	assert.NoError(t, err)
	assert.Equal(t, "u4", newReviewer)
	assert.Contains(t, result.AssignedReviewers, "u4")
	assert.NotContains(t, result.AssignedReviewers, "u2")
	suite.mockPRRepo.AssertExpectations(t)
	suite.mockUserRepo.AssertExpectations(t)
}

func TestPullRequestService_ReassignReviewer_MergedPR(t *testing.T) {
	suite := NewPRServiceTestSuite()

	pr := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            domain.PRStatusMerged,
		AssignedReviewers: []string{"u2", "u3"},
	}

	suite.mockPRRepo.On("GetByID", "pr-1").Return(pr, nil)

	result, newReviewer, err := suite.prService.ReassignReviewer("pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "", newReviewer)
	assert.Equal(t, domain.ErrPullRequestMerged, err)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_ReassignReviewer_ReviewerNotAssigned(t *testing.T) {
	suite := NewPRServiceTestSuite()

	pr := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u3", "u4"},
	}

	suite.mockPRRepo.On("GetByID", "pr-1").Return(pr, nil)

	result, newReviewer, err := suite.prService.ReassignReviewer("pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "", newReviewer)
	assert.Equal(t, domain.ErrReviewerNotAssigned, err)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_ReassignReviewer_NoReplacementCandidate(t *testing.T) {
	suite := NewPRServiceTestSuite()

	pr := &domain.PullRequest{
		ID:                "pr-1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            domain.PRStatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	oldReviewer := &domain.User{
		ID:       "u2",
		Username: "Bob",
		TeamName: "backend",
		IsActive: true,
	}

	teamUsers := []*domain.User{}

	suite.mockPRRepo.On("GetByID", "pr-1").Return(pr, nil)
	suite.mockUserRepo.On("GetByID", "u2").Return(oldReviewer, nil)
	suite.mockUserRepo.On("GetByTeam", "backend").Return(teamUsers, nil)

	result, newReviewer, err := suite.prService.ReassignReviewer("pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "", newReviewer)
	assert.Equal(t, domain.ErrNoCandidate, err)
	suite.mockPRRepo.AssertExpectations(t)
	suite.mockUserRepo.AssertExpectations(t)
}

func TestPullRequestService_GetUserReviews_Success(t *testing.T) {
	suite := NewPRServiceTestSuite()
	userID := "u1"

	expectedPRs := []*domain.PullRequestShort{
		{
			ID:       "pr-1",
			Name:     "Add feature",
			AuthorID: "u2",
			Status:   domain.PRStatusOpen,
		},
		{
			ID:       "pr-2",
			Name:     "Fix bug",
			AuthorID: "u3",
			Status:   domain.PRStatusMerged,
		},
	}

	suite.mockPRRepo.On("ListByReviewer", userID).Return([]*domain.PullRequest{
		{
			ID:       "pr-1",
			Name:     "Add feature",
			AuthorID: "u2",
			Status:   domain.PRStatusOpen,
		},
		{
			ID:       "pr-2",
			Name:     "Fix bug",
			AuthorID: "u3",
			Status:   domain.PRStatusMerged,
		},
	}, nil)

	result, err := suite.prService.GetUserReviews(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPRs, result)
	assert.Len(t, result, 2)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_GetUserReviews_EmptyList(t *testing.T) {
	suite := NewPRServiceTestSuite()
	userID := "u1"

	suite.mockPRRepo.On("ListByReviewer", userID).Return([]*domain.PullRequest{}, nil)

	result, err := suite.prService.GetUserReviews(userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Len(t, result, 0)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_GetUserReviews_RepositoryError(t *testing.T) {
	suite := NewPRServiceTestSuite()
	userID := "u1"
	expectedError := assert.AnError

	suite.mockPRRepo.On("ListByReviewer", userID).Return(nil, expectedError)

	result, err := suite.prService.GetUserReviews(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	suite.mockPRRepo.AssertExpectations(t)
}

func TestPullRequestService_GetUserReviews_WithMultiplePRsSameAuthor(t *testing.T) {
	suite := NewPRServiceTestSuite()
	userID := "u1"

	expectedPRs := []*domain.PullRequestShort{
		{
			ID:       "pr-1",
			Name:     "Feature A",
			AuthorID: "u2",
			Status:   domain.PRStatusOpen,
		},
		{
			ID:       "pr-2",
			Name:     "Feature B",
			AuthorID: "u2",
			Status:   domain.PRStatusOpen,
		},
	}

	suite.mockPRRepo.On("ListByReviewer", userID).Return([]*domain.PullRequest{
		{
			ID:       "pr-1",
			Name:     "Feature A",
			AuthorID: "u2",
			Status:   domain.PRStatusOpen,
		},
		{
			ID:       "pr-2",
			Name:     "Feature B",
			AuthorID: "u2",
			Status:   domain.PRStatusOpen,
		},
	}, nil)

	result, err := suite.prService.GetUserReviews(userID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, expectedPRs, result)
	assert.Equal(t, "u2", result[0].AuthorID)
	assert.Equal(t, "u2", result[1].AuthorID)
	suite.mockPRRepo.AssertExpectations(t)
}
