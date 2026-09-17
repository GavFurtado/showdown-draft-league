package services_test

import (
	"errors"
	"log/slog"
	"testing"

	mock_repos "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupSpectatorServiceTest() (services.SpectatorService, *mock_repos.MockSpectatorRepository, *mock_repos.MockPlayerRepository, *mock_repos.MockLeagueRepository) {
	mockSpectatorRepo := new(mock_repos.MockSpectatorRepository)
	mockPlayerRepo := new(mock_repos.MockPlayerRepository)
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)

	service := services.NewSpectatorService(
		slog.Default(),
		mockSpectatorRepo,
		mockPlayerRepo,
		mockLeagueRepo,
	)

	return service, mockSpectatorRepo, mockPlayerRepo, mockLeagueRepo
}

func TestSpectatorService_GetByLeague(t *testing.T) {
	service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.Spectator{{ID: uuid.New(), LeagueID: leagueID}}
		mockSpectatorRepo.On("GetByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockSpectatorRepo.On("GetByLeague", mock.Anything).Return([]models.Spectator(nil), errors.New("db error")).Once()

		result, err := service.GetByLeague(uuid.New())
		assert.ErrorIs(t, err, types.ErrInternalService)
		assert.Nil(t, result)
		mockSpectatorRepo.AssertExpectations(t)
	})
}

func TestSpectatorService_JoinLeague(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service, mockSpectatorRepo, mockPlayerRepo, mockLeagueRepo := setupSpectatorServiceTest()
		userID := uuid.New()
		leagueID := uuid.New()
		name := "spectating"
		created := &models.Spectator{ID: uuid.New(), UserID: userID, LeagueID: leagueID, Role: rbac.MRoleSpectator}

		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockPlayerRepo.On("FindByUserAndLeague", userID, leagueID).Return((*models.Player)(nil), nil).Once()
		mockSpectatorRepo.On("GetByUserAndLeague", userID, leagueID).Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()
		mockSpectatorRepo.On("Create", mock.AnythingOfType("*models.Spectator")).Return(created, nil).Once()

		result, err := service.JoinLeague(userID, leagueID, &name)
		assert.NoError(t, err)
		assert.Equal(t, created, result)
		mockLeagueRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("LeagueNotFound", func(t *testing.T) {
		service, _, _, mockLeagueRepo := setupSpectatorServiceTest()
		leagueID := uuid.New()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.JoinLeague(uuid.New(), leagueID, nil)
		assert.ErrorIs(t, err, types.ErrLeagueNotFound)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("AlreadyPlayer", func(t *testing.T) {
		service, _, mockPlayerRepo, mockLeagueRepo := setupSpectatorServiceTest()
		userID := uuid.New()
		leagueID := uuid.New()

		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockPlayerRepo.On("FindByUserAndLeague", userID, leagueID).Return(&models.Player{ID: uuid.New()}, nil).Once()

		result, err := service.JoinLeague(userID, leagueID, nil)
		assert.ErrorIs(t, err, types.ErrUserAlreadyInLeague)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
	})

	t.Run("AlreadySpectator", func(t *testing.T) {
		service, mockSpectatorRepo, mockPlayerRepo, mockLeagueRepo := setupSpectatorServiceTest()
		userID := uuid.New()
		leagueID := uuid.New()

		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(&models.League{ID: leagueID}, nil).Once()
		mockPlayerRepo.On("FindByUserAndLeague", userID, leagueID).Return((*models.Player)(nil), nil).Once()
		mockSpectatorRepo.On("GetByUserAndLeague", userID, leagueID).Return(&models.Spectator{ID: uuid.New()}, nil).Once()

		result, err := service.JoinLeague(userID, leagueID, nil)
		assert.ErrorIs(t, err, types.ErrUserAlreadySpectator)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
		mockPlayerRepo.AssertExpectations(t)
		mockSpectatorRepo.AssertExpectations(t)
	})
}

func TestSpectatorService_LeaveLeague(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()
		userID := uuid.New()
		leagueID := uuid.New()
		spectatorID := uuid.New()

		mockSpectatorRepo.On("GetByUserAndLeague", userID, leagueID).
			Return(&models.Spectator{ID: spectatorID, Role: rbac.MRoleSpectator}, nil).Once()
		mockSpectatorRepo.On("Delete", spectatorID).Return(nil).Once()

		err := service.LeaveLeague(userID, leagueID)
		assert.NoError(t, err)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()
		mockSpectatorRepo.On("GetByUserAndLeague", mock.Anything, mock.Anything).
			Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()

		err := service.LeaveLeague(uuid.New(), uuid.New())
		assert.ErrorIs(t, err, types.ErrSpectatorNotFound)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("OwnerCannotLeave", func(t *testing.T) {
		service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()
		mockSpectatorRepo.On("GetByUserAndLeague", mock.Anything, mock.Anything).
			Return(&models.Spectator{ID: uuid.New(), Role: rbac.MRoleSpectatorOwner}, nil).Once()

		err := service.LeaveLeague(uuid.New(), uuid.New())
		assert.ErrorIs(t, err, types.ErrCannotDemoteLeagueOwner)
		mockSpectatorRepo.AssertNotCalled(t, "Delete")
		mockSpectatorRepo.AssertExpectations(t)
	})
}

func TestSpectatorService_UpdateProfile(t *testing.T) {
	service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()

	t.Run("Success", func(t *testing.T) {
		spectatorID := uuid.New()
		existing := &models.Spectator{ID: spectatorID}
		name := "new name"
		updated := &models.Spectator{ID: spectatorID, InLeagueName: &name}

		mockSpectatorRepo.On("GetByID", spectatorID).Return(existing, nil).Once()
		mockSpectatorRepo.On("Update", existing).Return(updated, nil).Once()

		result, err := service.UpdateProfile(spectatorID, &name)
		assert.NoError(t, err)
		assert.Equal(t, updated, result)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSpectatorRepo.On("GetByID", mock.Anything).Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateProfile(uuid.New(), nil)
		assert.ErrorIs(t, err, types.ErrSpectatorNotFound)
		assert.Nil(t, result)
		mockSpectatorRepo.AssertExpectations(t)
	})
}

func TestSpectatorService_UpdateOwnProfile(t *testing.T) {
	service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		leagueID := uuid.New()
		existing := &models.Spectator{ID: uuid.New(), UserID: userID, LeagueID: leagueID}
		name := "self name"

		mockSpectatorRepo.On("GetByUserAndLeague", userID, leagueID).Return(existing, nil).Once()
		mockSpectatorRepo.On("Update", existing).Return(existing, nil).Once()

		result, err := service.UpdateOwnProfile(userID, leagueID, &name)
		assert.NoError(t, err)
		assert.Equal(t, existing, result)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSpectatorRepo.On("GetByUserAndLeague", mock.Anything, mock.Anything).
			Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateOwnProfile(uuid.New(), uuid.New(), nil)
		assert.ErrorIs(t, err, types.ErrSpectatorNotFound)
		assert.Nil(t, result)
		mockSpectatorRepo.AssertExpectations(t)
	})
}

func TestSpectatorService_UpdateRole(t *testing.T) {
	service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()

	t.Run("RejectsParticipantRole", func(t *testing.T) {
		result, err := service.UpdateRole(uuid.New(), rbac.MRoleRegular)
		assert.ErrorIs(t, err, types.ErrInvalidInput)
		assert.Nil(t, result)
		mockSpectatorRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("Success", func(t *testing.T) {
		spectatorID := uuid.New()
		expected := &models.Spectator{ID: spectatorID, Role: rbac.MRoleSpectatorModerator}

		mockSpectatorRepo.On("GetByID", spectatorID).Return(&models.Spectator{ID: spectatorID}, nil).Once()
		mockSpectatorRepo.On("UpdateRole", spectatorID, rbac.MRoleSpectatorModerator).Return(nil).Once()
		mockSpectatorRepo.On("GetByID", spectatorID).Return(expected, nil).Once()

		result, err := service.UpdateRole(spectatorID, rbac.MRoleSpectatorModerator)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSpectatorRepo.On("GetByID", mock.Anything).Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateRole(uuid.New(), rbac.MRoleSpectator)
		assert.ErrorIs(t, err, types.ErrSpectatorNotFound)
		assert.Nil(t, result)
		mockSpectatorRepo.AssertExpectations(t)
	})
}

func TestSpectatorService_DeleteSpectator(t *testing.T) {
	service, mockSpectatorRepo, _, _ := setupSpectatorServiceTest()

	t.Run("Success", func(t *testing.T) {
		spectatorID := uuid.New()
		mockSpectatorRepo.On("GetByID", spectatorID).Return(&models.Spectator{ID: spectatorID, Role: rbac.MRoleSpectator}, nil).Once()
		mockSpectatorRepo.On("Delete", spectatorID).Return(nil).Once()

		err := service.DeleteSpectator(spectatorID)
		assert.NoError(t, err)
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("OwnerCannotBeDeleted", func(t *testing.T) {
		spectatorID := uuid.New()
		mockSpectatorRepo.On("GetByID", spectatorID).Return(&models.Spectator{ID: spectatorID, Role: rbac.MRoleSpectatorOwner}, nil).Once()

		err := service.DeleteSpectator(spectatorID)
		assert.ErrorIs(t, err, types.ErrCannotDemoteLeagueOwner)
		mockSpectatorRepo.AssertNotCalled(t, "Delete")
		mockSpectatorRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockSpectatorRepo.On("GetByID", mock.Anything).Return((*models.Spectator)(nil), gorm.ErrRecordNotFound).Once()

		err := service.DeleteSpectator(uuid.New())
		assert.ErrorIs(t, err, types.ErrSpectatorNotFound)
		mockSpectatorRepo.AssertExpectations(t)
	})
}
