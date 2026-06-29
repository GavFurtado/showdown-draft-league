package services_test

import (
	"errors"
	"testing"

	mock_repos "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupClaimServiceTest() (services.ClaimService, *mock_repos.MockClaimRepository) {
	mockClaimRepo := new(mock_repos.MockClaimRepository)
	service := services.NewClaimService(mockClaimRepo)
	return service, mockClaimRepo
}

func TestClaimService_GetByID(t *testing.T) {
	service, mockClaimRepo := setupClaimServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.Claim{ID: uuid.New()}
		mockClaimRepo.On("GetByID", expected.ID).Return(expected, nil).Once()

		result, err := service.GetByID(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockClaimRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockClaimRepo.On("GetByID", id).Return((*models.Claim)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrClaimNotFound, err)
		assert.Nil(t, result)
		mockClaimRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		id := uuid.New()
		mockClaimRepo.On("GetByID", id).Return((*models.Claim)(nil), errors.New("db error")).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockClaimRepo.AssertExpectations(t)
	})
}

func TestClaimService_GetActiveByPlayer(t *testing.T) {
	service, mockClaimRepo := setupClaimServiceTest()

	t.Run("Success", func(t *testing.T) {
		playerID := uuid.New()
		expected := []models.Claim{{ID: uuid.New(), PlayerID: playerID, IsActive: true}}
		mockClaimRepo.On("GetActiveByPlayer", playerID).Return(expected, nil).Once()

		result, err := service.GetActiveByPlayer(playerID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockClaimRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockClaimRepo.On("GetActiveByPlayer", mock.Anything).Return([]models.Claim(nil), errors.New("db error")).Once()

		result, err := service.GetActiveByPlayer(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockClaimRepo.AssertExpectations(t)
	})
}

func TestClaimService_GetActiveByLeague(t *testing.T) {
	service, mockClaimRepo := setupClaimServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.Claim{{ID: uuid.New(), LeagueID: leagueID, IsActive: true}}
		mockClaimRepo.On("GetActiveByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetActiveByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockClaimRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockClaimRepo.On("GetActiveByLeague", mock.Anything).Return([]models.Claim(nil), errors.New("db error")).Once()

		result, err := service.GetActiveByLeague(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockClaimRepo.AssertExpectations(t)
	})
}

func TestClaimService_GetReleasedByLeague(t *testing.T) {
	service, mockClaimRepo := setupClaimServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.Claim{{ID: uuid.New(), LeagueID: leagueID, IsActive: false}}
		mockClaimRepo.On("GetReleasedByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetReleasedByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockClaimRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockClaimRepo.On("GetReleasedByLeague", mock.Anything).Return([]models.Claim(nil), errors.New("db error")).Once()

		result, err := service.GetReleasedByLeague(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockClaimRepo.AssertExpectations(t)
	})
}
