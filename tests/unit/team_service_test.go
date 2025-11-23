package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/repository/mocks"
	"github.com/111zxc/pr-review-service/internal/service"
)

type TeamServiceTestSuite struct {
	mockTeamRepo *mocks.TeamRepository
	mockUserRepo *mocks.UserRepository
	teamService  *service.TeamService
}

func NewTeamServiceTestSuite() *TeamServiceTestSuite {
	mockTeamRepo := new(mocks.TeamRepository)
	mockUserRepo := new(mocks.UserRepository)
	teamService := service.NewTeamService(mockTeamRepo, mockUserRepo)

	return &TeamServiceTestSuite{
		mockTeamRepo: mockTeamRepo,
		mockUserRepo: mockUserRepo,
		teamService:  teamService,
	}
}

func CreateTestTeam() *domain.Team {
	return &domain.Team{
		Name: "backend",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}
}

func TestTeamService_CreateTeam_Success(t *testing.T) {
	suite := NewTeamServiceTestSuite()
	team := CreateTestTeam()

	suite.mockTeamRepo.On("Exists", "backend").Return(false, nil)
	suite.mockUserRepo.On("Create", mock.MatchedBy(func(user *domain.User) bool {
		return user.ID == "u1" && user.Username == "Alice"
	})).Return(nil)
	suite.mockUserRepo.On("Create", mock.MatchedBy(func(user *domain.User) bool {
		return user.ID == "u2" && user.Username == "Bob"
	})).Return(nil)
	suite.mockTeamRepo.On("Create", team).Return(nil)

	err := suite.teamService.CreateTeam(team)

	assert.NoError(t, err)
	suite.mockTeamRepo.AssertExpectations(t)
	suite.mockUserRepo.AssertExpectations(t)
}

func TestTeamService_CreateTeam_AlreadyExists(t *testing.T) {
	suite := NewTeamServiceTestSuite()
	team := CreateTestTeam()

	suite.mockTeamRepo.On("Exists", "backend").Return(true, nil)

	err := suite.teamService.CreateTeam(team)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrTeamExists, err)
	suite.mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeam_Success(t *testing.T) {
	suite := NewTeamServiceTestSuite()
	expectedTeam := CreateTestTeam()

	suite.mockTeamRepo.On("GetByName", "backend").Return(expectedTeam, nil)

	team, err := suite.teamService.GetTeam("backend")

	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
	suite.mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeam_NotFound(t *testing.T) {
	suite := NewTeamServiceTestSuite()

	suite.mockTeamRepo.On("GetByName", "nonexistent").Return(nil, domain.ErrTeamNotFound)

	team, err := suite.teamService.GetTeam("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Equal(t, domain.ErrTeamNotFound, err)
	suite.mockTeamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeam_RepositoryError(t *testing.T) {
	suite := NewTeamServiceTestSuite()
	expectedError := assert.AnError

	suite.mockTeamRepo.On("GetByName", "backend").Return(nil, expectedError)

	team, err := suite.teamService.GetTeam("backend")

	assert.Error(t, err)
	assert.Nil(t, team)
	assert.Equal(t, expectedError, err)
	suite.mockTeamRepo.AssertExpectations(t)
}
