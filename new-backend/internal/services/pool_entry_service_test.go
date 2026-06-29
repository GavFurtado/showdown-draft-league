package services_test

import (
	"errors"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	mock_repos "github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupPoolEntryServiceTest() (services.PoolEntryService, *mock_repos.MockPoolEntryRepository, *mock_repos.MockLeagueRepository, *mock_repos.MockUserRepository, *mock_repos.MockPokemonSpeciesRepository) {
	mockPoolEntryRepo := new(mock_repos.MockPoolEntryRepository)
	mockLeagueRepo := new(mock_repos.MockLeagueRepository)
	mockUserRepo := new(mock_repos.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repos.MockPokemonSpeciesRepository)

	service := services.NewPoolEntryService(
		mockPoolEntryRepo,
		mockLeagueRepo,
		mockUserRepo,
		mockPokemonSpeciesRepo,
	)

	return service, mockPoolEntryRepo, mockLeagueRepo, mockUserRepo, mockPokemonSpeciesRepo
}

func TestPoolEntryService_GetByID(t *testing.T) {
	service, mockPoolEntryRepo, _, _, _ := setupPoolEntryServiceTest()

	t.Run("Success", func(t *testing.T) {
		expected := &models.PoolEntry{ID: uuid.New(), IsAvailable: true}
		mockPoolEntryRepo.On("GetByID", expected.ID).Return(expected, nil).Once()

		result, err := service.GetByID(expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockPoolEntryRepo.On("GetByID", id).Return((*models.PoolEntry)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPoolEntryNotFound, err)
		assert.Nil(t, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		id := uuid.New()
		mockPoolEntryRepo.On("GetByID", id).Return((*models.PoolEntry)(nil), errors.New("db error")).Once()

		result, err := service.GetByID(id)
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})
}

func TestPoolEntryService_GetByLeague(t *testing.T) {
	service, mockPoolEntryRepo, _, _, _ := setupPoolEntryServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.PoolEntry{{ID: uuid.New(), LeagueID: leagueID}}
		mockPoolEntryRepo.On("GetByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		leagueID := uuid.New()
		mockPoolEntryRepo.On("GetByLeague", leagueID).Return([]models.PoolEntry(nil), errors.New("db error")).Once()

		result, err := service.GetByLeague(leagueID)
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})
}

func TestPoolEntryService_GetAvailableByLeague(t *testing.T) {
	service, mockPoolEntryRepo, _, _, _ := setupPoolEntryServiceTest()

	t.Run("Success", func(t *testing.T) {
		leagueID := uuid.New()
		expected := []models.PoolEntry{{ID: uuid.New(), LeagueID: leagueID, IsAvailable: true}}
		mockPoolEntryRepo.On("GetAvailableByLeague", leagueID).Return(expected, nil).Once()

		result, err := service.GetAvailableByLeague(leagueID)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})

	t.Run("InternalError", func(t *testing.T) {
		leagueID := uuid.New()
		mockPoolEntryRepo.On("GetAvailableByLeague", leagueID).Return([]models.PoolEntry(nil), errors.New("db error")).Once()

		result, err := service.GetAvailableByLeague(leagueID)
		assert.Error(t, err)
		assert.Equal(t, types.ErrInternalService, err)
		assert.Nil(t, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})
}

func TestPoolEntryService_Create(t *testing.T) {
	service, mockPoolEntryRepo, mockLeagueRepo, _, mockPokemonSpeciesRepo := setupPoolEntryServiceTest()

	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		leagueID := uuid.New()
		speciesID := int64(1)
		cost := 50

		league := &models.League{ID: leagueID, Status: enums.LeagueStatusSetup}
		species := &models.PokemonSpecies{ID: speciesID}

		input := &requests.PoolEntryCreateRequestDTO{
			LeagueID:         leagueID,
			PokemonSpeciesID: speciesID,
			Cost:             &cost,
		}

		created := &models.PoolEntry{
			LeagueID:         leagueID,
			PokemonSpeciesID: speciesID,
			Cost:             &cost,
			IsAvailable:      true,
		}
		created.ID = uuid.New()

		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", speciesID).Return(species, nil).Once()
		mockPoolEntryRepo.On("Create", mock.AnythingOfType("*models.PoolEntry")).Return(created, nil).Once()

		result, err := service.Create(&models.User{ID: userID}, input)
		assert.NoError(t, err)
		assert.Equal(t, created, result)
		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPoolEntryRepo.AssertExpectations(t)
	})

	t.Run("LeagueNotFound", func(t *testing.T) {
		leagueID := uuid.New()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		input := &requests.PoolEntryCreateRequestDTO{LeagueID: leagueID}
		result, err := service.Create(&models.User{ID: uuid.New()}, input)
		assert.Error(t, err)
		assert.Equal(t, types.ErrLeagueNotFound, err)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("InvalidLeagueStatus", func(t *testing.T) {
		leagueID := uuid.New()
		league := &models.League{ID: leagueID, Status: enums.LeagueStatusDrafting}
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(league, nil).Once()

		input := &requests.PoolEntryCreateRequestDTO{LeagueID: leagueID}
		result, err := service.Create(&models.User{ID: uuid.New()}, input)
		assert.Error(t, err)
		assert.Equal(t, types.ErrInvalidState, err)
		assert.Nil(t, result)
		mockLeagueRepo.AssertExpectations(t)
	})
}

func TestPoolEntryService_Update(t *testing.T) {
	service, mockPoolEntryRepo, mockLeagueRepo, _, _ := setupPoolEntryServiceTest()

	t.Run("Success", func(t *testing.T) {
		poolEntryID := uuid.New()
		leagueID := uuid.New()
		cost := 60
		isAvailable := true

		existing := &models.PoolEntry{ID: poolEntryID, LeagueID: leagueID, Cost: &cost, IsAvailable: false}
		league := &models.League{ID: leagueID, Status: enums.LeagueStatusSetup}

		input := &requests.PoolEntryUpdateRequestDTO{
			PoolEntryID: poolEntryID,
			Cost:        &cost,
			IsAvailable: &isAvailable,
		}

		mockPoolEntryRepo.On("GetByID", poolEntryID).Return(existing, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", leagueID).Return(league, nil).Once()
		mockPoolEntryRepo.On("Update", mock.AnythingOfType("*models.PoolEntry")).Return(existing, nil).Once()

		result, err := service.Update(&models.User{ID: uuid.New()}, input)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockPoolEntryRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		mockPoolEntryRepo.On("GetByID", id).Return((*models.PoolEntry)(nil), gorm.ErrRecordNotFound).Once()

		input := &requests.PoolEntryUpdateRequestDTO{PoolEntryID: id}
		result, err := service.Update(&models.User{ID: uuid.New()}, input)
		assert.Error(t, err)
		assert.Equal(t, types.ErrPoolEntryNotFound, err)
		assert.Nil(t, result)
		mockPoolEntryRepo.AssertExpectations(t)
	})
}
