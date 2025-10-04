package services_test

import (
	"errors"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserService_GetMyProfileHandler(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := services.NewUserService(mockUserRepo)

	userID := uuid.New()

	t.Run("Successfully gets user profile", func(t *testing.T) {
		expectedUser := &models.User{ID: userID, DiscordUsername: "testuser"}
		mockUserRepo.On("GetUserByID", userID).Return(expectedUser, nil).Once()

		user, err := service.GetMyProfileHandler(userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound if user not found", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		user, err := service.GetMyProfileHandler(userID)
		assert.ErrorIs(t, err, common.ErrUserNotFound)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for other repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), internalErr).Once()

		user, err := service.GetMyProfileHandler(userID)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetMyDiscordDetailsHandler(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := services.NewUserService(mockUserRepo)

	userID := uuid.New()

	t.Run("Successfully gets Discord details", func(t *testing.T) {
		user := &models.User{ID: userID, DiscordUsername: "testdiscord", DiscordAvatarURL: "avatar.url"}
		expectedDiscordUser := &common.DiscordUser{ID: userID.String(), Username: "testdiscord", Avatar: "avatar.url"}
		mockUserRepo.On("GetUserByID", userID).Return(user, nil).Once()

		discordUser, err := service.GetMyDiscordDetailsHandler(userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedDiscordUser, discordUser)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound if user not found", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		discordUser, err := service.GetMyDiscordDetailsHandler(userID)
		assert.ErrorIs(t, err, common.ErrUserNotFound)
		assert.Nil(t, discordUser)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for other repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), internalErr).Once()

		discordUser, err := service.GetMyDiscordDetailsHandler(userID)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, discordUser)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateProfileHandler(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := services.NewUserService(mockUserRepo)

	userID := uuid.New()
	showdownName := "newshowdown"
	updateReq := common.UserUpdateProfileRequest{ShowdownName: &showdownName}

	t.Run("Successfully updates user profile", func(t *testing.T) {
		originalUser := &models.User{ID: userID, ShowdownUsername: "oldshowdown"}
		updatedUser := &models.User{ID: userID, ShowdownUsername: "newshowdown"}

		mockUserRepo.On("GetUserByID", userID).Return(originalUser, nil).Once()
		mockUserRepo.On("UpdateUser", updatedUser).Return(updatedUser, nil).Once()

		user, err := service.UpdateProfileHandler(userID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, updatedUser, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound if user not found", func(t *testing.T) {
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound).Once()

		user, err := service.UpdateProfileHandler(userID, updateReq)
		assert.ErrorIs(t, err, common.ErrUserNotFound)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService if GetUserByID fails", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockUserRepo.On("GetUserByID", userID).Return((*models.User)(nil), internalErr).Once()

		user, err := service.UpdateProfileHandler(userID, updateReq)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService if UpdateUser fails", func(t *testing.T) {
		originalUser := &models.User{ID: userID, ShowdownUsername: "oldshowdown"}
		updatedUser := &models.User{ID: userID, ShowdownUsername: "newshowdown"}
		internalErr := errors.New("db update error")

		mockUserRepo.On("GetUserByID", userID).Return(originalUser, nil).Once()
		mockUserRepo.On("UpdateUser", updatedUser).Return((*models.User)(nil), internalErr).Once()

		user, err := service.UpdateProfileHandler(userID, updateReq)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetMyLeaguesHandler(t *testing.T) {
	mockUserRepo := new(mock_repositories.MockUserRepository)
	service := services.NewUserService(mockUserRepo)

	userID := uuid.New()

	t.Run("Successfully gets user leagues", func(t *testing.T) {
		expectedLeagues := []models.League{
			{ID: uuid.New(), Name: "League 1"},
			{ID: uuid.New(), Name: "League 2"},
		}
		mockUserRepo.On("GetUserLeagues", userID).Return(expectedLeagues, nil).Once()

		leagues, err := service.GetMyLeaguesHandler(userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeagues, leagues)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrUserNotFound if user not found", func(t *testing.T) {
		mockUserRepo.On("GetUserLeagues", userID).Return(([]models.League)(nil), gorm.ErrRecordNotFound).Once()

		leagues, err := service.GetMyLeaguesHandler(userID)
		assert.ErrorIs(t, err, common.ErrUserNotFound)
		assert.Nil(t, leagues)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for other repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockUserRepo.On("GetUserLeagues", userID).Return(([]models.League)(nil), internalErr).Once()

		leagues, err := service.GetMyLeaguesHandler(userID)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, leagues)
		mockUserRepo.AssertExpectations(t)
	})
}
