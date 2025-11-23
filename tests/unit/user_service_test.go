package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/111zxc/pr-review-service/internal/domain"
	"github.com/111zxc/pr-review-service/internal/repository/mocks"
	"github.com/111zxc/pr-review-service/internal/service"
)

type UserServiceTestSuite struct {
	mockUserRepo *mocks.UserRepository
	userService  *service.UserService
}

func NewUserServiceTestSuite() *UserServiceTestSuite {
	mockUserRepo := new(mocks.UserRepository)
	userService := service.NewUserService(mockUserRepo)

	return &UserServiceTestSuite{
		mockUserRepo: mockUserRepo,
		userService:  userService,
	}
}

func CreateTestUser() *domain.User {
	return &domain.User{
		ID:       "u1",
		Username: "Alice",
		IsActive: true,
		TeamName: "backend",
	}
}

func TestUserService_SetUserActive_Success(t *testing.T) {
	suite := NewUserServiceTestSuite()
	user := CreateTestUser()

	suite.mockUserRepo.On("GetByID", "u1").Return(user, nil)
	suite.mockUserRepo.On("Update", mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u1" && u.IsActive == false
	})).Return(nil)

	result, err := suite.userService.SetUserActive("u1", false)

	assert.NoError(t, err)
	assert.False(t, result.IsActive)
	suite.mockUserRepo.AssertExpectations(t)
}

func TestUserService_SetUserActive_UserNotFound(t *testing.T) {
	suite := NewUserServiceTestSuite()

	suite.mockUserRepo.On("GetByID", "u1").Return(nil, domain.ErrUserNotFound)

	result, err := suite.userService.SetUserActive("u1", false)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrUserNotFound, err)
	suite.mockUserRepo.AssertExpectations(t)
}
