package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/repository/mocks"
	"github.com/111zxc/pr-review-service/internal/service"
)

type StatsServiceTestSuite struct {
	mockStatsRepo *mocks.StatsRepository
	statsService  *service.StatsService
}

func NewStatsServiceTestSuite() *StatsServiceTestSuite {
	mockStatsRepo := new(mocks.StatsRepository)
	statsService := service.NewStatsService(mockStatsRepo)

	return &StatsServiceTestSuite{
		mockStatsRepo: mockStatsRepo,
		statsService:  statsService,
	}
}

func TestStatsService_GetStats_Success(t *testing.T) {
	suite := NewStatsServiceTestSuite()

	expectedStats := &domain.StatsResponse{
		EventCounts: map[domain.EventType]int{
			domain.EventTypePRCreated:          15,
			domain.EventTypePRMerged:           12,
			domain.EventTypeReviewerAssigned:   25,
			domain.EventTypeReviewerReassigned: 3,
			domain.EventTypeReviewerUnassigned: 0,
		},
		TotalEvents: 55,
	}

	suite.mockStatsRepo.On("GetEventStats").Return(expectedStats, nil)

	result, err := suite.statsService.GetStats()

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, result)
	assert.Equal(t, 55, result.TotalEvents)
	assert.Equal(t, 15, result.EventCounts[domain.EventTypePRCreated])
	assert.Equal(t, 12, result.EventCounts[domain.EventTypePRMerged])
	assert.Equal(t, 25, result.EventCounts[domain.EventTypeReviewerAssigned])
	assert.Equal(t, 3, result.EventCounts[domain.EventTypeReviewerReassigned])
	assert.Equal(t, 0, result.EventCounts[domain.EventTypeReviewerUnassigned])
	suite.mockStatsRepo.AssertExpectations(t)
}

func TestStatsService_GetStats_EmptyStats(t *testing.T) {
	suite := NewStatsServiceTestSuite()

	expectedStats := &domain.StatsResponse{
		EventCounts: map[domain.EventType]int{
			domain.EventTypePRCreated:          0,
			domain.EventTypePRMerged:           0,
			domain.EventTypeReviewerAssigned:   0,
			domain.EventTypeReviewerReassigned: 0,
			domain.EventTypeReviewerUnassigned: 0,
		},
		TotalEvents: 0,
	}

	suite.mockStatsRepo.On("GetEventStats").Return(expectedStats, nil)

	result, err := suite.statsService.GetStats()

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, result)
	assert.Equal(t, 0, result.TotalEvents)
	for eventType, count := range result.EventCounts {
		assert.Equal(t, 0, count, "Event type %s should have count 0", eventType)
	}
	suite.mockStatsRepo.AssertExpectations(t)
}

func TestStatsService_GetStats_RepositoryError(t *testing.T) {
	suite := NewStatsServiceTestSuite()
	expectedError := assert.AnError

	suite.mockStatsRepo.On("GetEventStats").Return(nil, expectedError)

	result, err := suite.statsService.GetStats()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedError, err)
	suite.mockStatsRepo.AssertExpectations(t)
}
