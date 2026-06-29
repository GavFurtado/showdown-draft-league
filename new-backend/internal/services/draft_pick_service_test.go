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

func setupDraftPickServiceTest() (services.DraftPickService, *mock_repos.MockDraftPickRepository, *mock_repos.MockDraftRepository) {
	mockDraftPickRepo := new(mock_repos.MockDraftPickRepository)
	mockDraftRepo := new(mock_repos.MockDraftRepository)

	service := services.NewDraftPickService(
		mockDraftPickRepo,
		mockDraftRepo,
	)

	return service, mockDraftPickRepo, mockDraftRepo
}

func TestDraftPickService_GetByID(t *testing.T) {
	service, mockDraftPickRepo, _ := setupDraftPickServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.DraftPick{ID: uuid.New()}
		mockDraftPickRepo.On("GetByID", expected.ID).Return(expected, nil).Once()

		result, err := service.GetByID(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockDraftPickRepo.On("GetByID", id).Return((*models.DraftPick)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrDraftPickNotFound, err)
		assert.Nil(t, result)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		id := uuid.New()
		mockDraftPickRepo.On("GetByID", id).Return((*models.DraftPick)(nil), errors.New("db error")).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockDraftPickRepo.AssertExpectations(t)
	})
}

func TestDraftPickService_GetByDraft(t *testing.T) {
	service, mockDraftPickRepo, _ := setupDraftPickServiceTest()

	t.Run("Success", func(t *testing.T) {
		draftID := uuid.New()
		expected := []models.DraftPick{{ID: uuid.New(), DraftID: draftID}}
		mockDraftPickRepo.On("GetByDraft", draftID).Return(expected, nil).Once()

		result, err := service.GetByDraft(draftID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockDraftPickRepo.On("GetByDraft", mock.Anything).Return([]models.DraftPick(nil), errors.New("db error")).Once()

		result, err := service.GetByDraft(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockDraftPickRepo.AssertExpectations(t)
	})
}

func TestDraftPickService_GetByPlayer(t *testing.T) {
	service, mockDraftPickRepo, _ := setupDraftPickServiceTest()

	t.Run("Success", func(t *testing.T) {
		playerID := uuid.New()
		expected := []models.DraftPick{{ID: uuid.New(), PlayerID: playerID}}
		mockDraftPickRepo.On("GetByPlayer", playerID).Return(expected, nil).Once()

		result, err := service.GetByPlayer(playerID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockDraftPickRepo.On("GetByPlayer", mock.Anything).Return([]models.DraftPick(nil), errors.New("db error")).Once()

		result, err := service.GetByPlayer(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockDraftPickRepo.AssertExpectations(t)
	})
}

func TestDraftPickService_GetCountByDraft(t *testing.T) {
	service, mockDraftPickRepo, _ := setupDraftPickServiceTest()

	t.Run("Success", func(t *testing.T) {
		draftID := uuid.New()
		mockDraftPickRepo.On("GetCountByDraft", draftID).Return(int64(5), nil).Once()

		count, err := service.GetCountByDraft(draftID)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockDraftPickRepo.On("GetCountByDraft", mock.Anything).Return(int64(0), errors.New("db error")).Once()

		count, err := service.GetCountByDraft(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.Equal(t, types.ErrInternalService, err)
		mockDraftPickRepo.AssertExpectations(t)
	})
}

func TestDraftPickService_GetHistory(t *testing.T) {
	service, mockDraftPickRepo, mockDraftRepo := setupDraftPickServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		draftID := uuid.New()
		draft := &models.Draft{ID: draftID}
		expected := []models.DraftPick{{ID: uuid.New(), DraftID: draftID}}

		mockDraftRepo.On("GetDraftByLeagueID", leagueID).Return(draft, nil).Once()
		mockDraftPickRepo.On("GetByDraft", draftID).Return(expected, nil).Once()

		result, err := service.GetHistory(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockDraftRepo.AssertExpectations(t)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("DraftNotFound", func(t *testing.T) {
		mockDraftRepo.On("GetDraftByLeagueID", mock.Anything).Return((*models.Draft)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetHistory(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, types.ErrDraftNotFound, err)
		assert.Nil(t, result)
		mockDraftRepo.AssertExpectations(t)
	})
}

func TestDraftPickService_GetNextPickNumber(t *testing.T) {
	service, mockDraftPickRepo, _ := setupDraftPickServiceTest()

	t.Run("Success", func(t *testing.T) {
		draftID := uuid.New()
		mockDraftPickRepo.On("GetCountByDraft", draftID).Return(int64(10), nil).Once()

		number, err := service.GetNextPickNumber(draftID)
		assert.NoError(t, err)
		assert.Equal(t, 11, number)
		mockDraftPickRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockDraftPickRepo.On("GetCountByDraft", mock.Anything).Return(int64(0), errors.New("db error")).Once()

		number, err := service.GetNextPickNumber(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, 0, number)
		assert.Equal(t, types.ErrInternalService, err)
		mockDraftPickRepo.AssertExpectations(t)
	})
}
