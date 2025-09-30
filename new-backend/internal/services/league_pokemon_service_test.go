package services_test

import (
	"errors"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestLeaguePokemonService_CreatePokemonForLeague(t *testing.T) {
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)

	service := services.NewLeaguePokemonService(
		mockLeaguePokemonRepo,
		mockLeagueRepo,
		mockUserRepo,
		mockPokemonSpeciesRepo,
	)

	testLeagueID := uuid.New()
	testUserID := uuid.New()
	testPokemonSpeciesID := int64(1)
	testCost := 100

	currentUser := &models.User{ID: testUserID}
	input := &common.LeaguePokemonCreateRequest{
		LeagueID:         testLeagueID,
		PokemonSpeciesID: testPokemonSpeciesID,
		Cost:             &testCost,
	}

	t.Run("Successfully creates a single league pokemon", func(t *testing.T) {
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}
		pokemonSpecies := &models.PokemonSpecies{ID: testPokemonSpeciesID, Name: "Pikachu"}
		expectedLeaguePokemon := &models.LeaguePokemon{
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &testCost,
			IsAvailable:      true,
		}

		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return(pokemonSpecies, nil).Once()
		mockLeaguePokemonRepo.On("CreateLeaguePokemon", expectedLeaguePokemon).Return(expectedLeaguePokemon, nil).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedLeaguePokemon, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertExpectations(t)
	})

	t.Run("Fails if league not found", func(t *testing.T) {
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.ErrorIs(t, err, common.ErrLeagueNotFound)
		assert.Nil(t, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByID")
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemon")
	})

	t.Run("Fails if league is not in Setup status", func(t *testing.T) {
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusDrafting} // Not Setup
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInvalidState)
		assert.Nil(t, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByID")
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemon")
	})

	t.Run("Fails if pokemon species not found", func(t *testing.T) {
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)
		assert.Nil(t, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemon")
	})

	t.Run("Fails if league repository returns internal error", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return((*models.League)(nil), internalErr).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByID")
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemon")
	})

	t.Run("Fails if pokemon species repository returns internal error", func(t *testing.T) {
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}
		internalErr := errors.New("db error")
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemon")
	})

	t.Run("Fails if league pokemon repository returns internal error", func(t *testing.T) {
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}
		pokemonSpecies := &models.PokemonSpecies{ID: testPokemonSpeciesID, Name: "Pikachu"}
		leaguePokemon := &models.LeaguePokemon{
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &testCost,
			IsAvailable:      true,
		}
		internalErr := errors.New("db error")

		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", testPokemonSpeciesID).Return(pokemonSpecies, nil).Once()
		mockLeaguePokemonRepo.On("CreateLeaguePokemon", leaguePokemon).Return((*models.LeaguePokemon)(nil), internalErr).Once()

		result, err := service.CreatePokemonForLeague(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, result)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertExpectations(t)
	})
}

func TestLeaguePokemonService_BatchCreatePokemonForLeague(t *testing.T) {
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)

	service := services.NewLeaguePokemonService(
		mockLeaguePokemonRepo,
		mockLeagueRepo,
		mockUserRepo,
		mockPokemonSpeciesRepo,
	)

	testLeagueID := uuid.New()
	testUserID := uuid.New()
	currentUser := &models.User{ID: testUserID}

	t.Run("Successfully creates multiple league pokemon", func(t *testing.T) {
		inputs := []*common.LeaguePokemonCreateRequest{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0]},
			{LeagueID: testLeagueID, PokemonSpeciesID: 2, Cost: &[]int{150}[0]},
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}
		pokemonSpecies1 := &models.PokemonSpecies{ID: 1, Name: "Pikachu"}
		pokemonSpecies2 := &models.PokemonSpecies{ID: 2, Name: "Charmander"}

		expectedBatch := []models.LeaguePokemon{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0], IsAvailable: true},
			{LeagueID: testLeagueID, PokemonSpeciesID: 2, Cost: &[]int{150}[0], IsAvailable: true},
		}

		// League is cached, so only called once
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", int64(1)).Return(pokemonSpecies1, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", int64(2)).Return(pokemonSpecies2, nil).Once()
		mockLeaguePokemonRepo.On("CreateLeaguePokemonBatch", expectedBatch).Return(nil).Once()

		results, err := service.BatchCreatePokemonForLeague(currentUser, inputs)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, testLeagueID, results[0].LeagueID)
		assert.Equal(t, int64(1), results[0].PokemonSpeciesID)
		assert.Equal(t, testLeagueID, results[1].LeagueID)
		assert.Equal(t, int64(2), results[1].PokemonSpeciesID)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertExpectations(t)
	})

	t.Run("Returns empty slice for empty input", func(t *testing.T) {
		inputs := []*common.LeaguePokemonCreateRequest{}

		results, err := service.BatchCreatePokemonForLeague(currentUser, inputs)
		assert.NoError(t, err)
		assert.Empty(t, results)

		// No repository calls should be made
		mockLeagueRepo.AssertNotCalled(t, "GetLeagueByID")
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByID")
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemonBatch")
	})

	t.Run("Fails if any league in batch is not found", func(t *testing.T) {
		inputs := []*common.LeaguePokemonCreateRequest{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0]},
			{LeagueID: uuid.New(), PokemonSpeciesID: 2, Cost: &[]int{150}[0]}, // Second league not found
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}

		// First league lookup succeeds
		mockLeagueRepo.On("GetLeagueByID", inputs[0].LeagueID).Return(league, nil).Once()
		// First pokemon species lookup succeeds
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", inputs[0].PokemonSpeciesID).Return(&models.PokemonSpecies{ID: inputs[0].PokemonSpeciesID}, nil).Once()
		// Second league lookup fails - this should cause early return
		mockLeagueRepo.On("GetLeagueByID", inputs[1].LeagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		results, err := service.BatchCreatePokemonForLeague(currentUser, inputs)
		assert.ErrorIs(t, err, common.ErrLeagueNotFound)
		assert.Nil(t, results)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		// No batch creation should happen on validation failure
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemonBatch")
	})

	t.Run("Fails if any league in batch is not in Setup status", func(t *testing.T) {
		inputs := []*common.LeaguePokemonCreateRequest{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0]},
			{LeagueID: testLeagueID, PokemonSpeciesID: 2, Cost: &[]int{150}[0]},
		}
		leagueDrafting := &models.League{ID: testLeagueID, Status: enums.LeagueStatusDrafting} // Not Setup

		// League is cached, so only called once
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(leagueDrafting, nil).Once()

		results, err := service.BatchCreatePokemonForLeague(currentUser, inputs)
		assert.ErrorIs(t, err, common.ErrInvalidState)
		assert.Nil(t, results)

		mockLeagueRepo.AssertExpectations(t)
		// Should fail on first league status check, no pokemon species lookup
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByID")
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemonBatch")
	})

	t.Run("Fails if any pokemon species in batch not found", func(t *testing.T) {
		inputs := []*common.LeaguePokemonCreateRequest{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0]},
			{LeagueID: testLeagueID, PokemonSpeciesID: 999, Cost: &[]int{150}[0]}, // PokemonSpeciesID 999 not found
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}

		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", int64(1)).Return(&models.PokemonSpecies{ID: 1}, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", int64(999)).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()

		results, err := service.BatchCreatePokemonForLeague(currentUser, inputs)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)
		assert.Nil(t, results)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		// Should fail during validation, no batch creation
		mockLeaguePokemonRepo.AssertNotCalled(t, "CreateLeaguePokemonBatch")
	})

	t.Run("Fails if batch create operation returns internal error", func(t *testing.T) {
		inputs := []*common.LeaguePokemonCreateRequest{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0]},
			{LeagueID: testLeagueID, PokemonSpeciesID: 2, Cost: &[]int{150}[0]},
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}
		pokemonSpecies1 := &models.PokemonSpecies{ID: 1, Name: "Pikachu"}
		pokemonSpecies2 := &models.PokemonSpecies{ID: 2, Name: "Charmander"}
		internalErr := errors.New("db error")

		expectedBatch := []models.LeaguePokemon{
			{LeagueID: testLeagueID, PokemonSpeciesID: 1, Cost: &[]int{100}[0], IsAvailable: true},
			{LeagueID: testLeagueID, PokemonSpeciesID: 2, Cost: &[]int{150}[0], IsAvailable: true},
		}

		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", int64(1)).Return(pokemonSpecies1, nil).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", int64(2)).Return(pokemonSpecies2, nil).Once()
		mockLeaguePokemonRepo.On("CreateLeaguePokemonBatch", expectedBatch).Return(internalErr).Once()

		results, err := service.BatchCreatePokemonForLeague(currentUser, inputs)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, results)

		mockLeagueRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertExpectations(t)
	})
}

func TestLeaguePokemonService_UpdateLeaguePokemon(t *testing.T) {
	mockLeaguePokemonRepo := new(mock_repositories.MockLeaguePokemonRepository)
	mockLeagueRepo := new(mock_repositories.MockLeagueRepository)
	mockUserRepo := new(mock_repositories.MockUserRepository)
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)

	service := services.NewLeaguePokemonService(
		mockLeaguePokemonRepo,
		mockLeagueRepo,
		mockUserRepo,
		mockPokemonSpeciesRepo,
	)

	testLeaguePokemonID := uuid.New()
	testLeagueID := uuid.New()
	testUserID := uuid.New()
	testPokemonSpeciesID := int64(1)
	currentUser := &models.User{ID: testUserID}

	t.Run("Successfully updates league pokemon cost and availability", func(t *testing.T) {
		originalCost := 100
		originalIsAvailable := true
		newCost := 120
		newIsAvailable := false

		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &originalCost,
			IsAvailable:      originalIsAvailable,
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}

		input := &common.LeaguePokemonUpdateRequest{
			LeaguePokemonID: testLeaguePokemonID,
			Cost:            &newCost,
			IsAvailable:     newIsAvailable,
		}

		updatedLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &newCost,
			IsAvailable:      newIsAvailable,
		}

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockLeaguePokemonRepo.On("UpdateLeaguePokemon", updatedLeaguePokemon).Return(updatedLeaguePokemon, nil).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.NoError(t, err)
		assert.Equal(t, updatedLeaguePokemon, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Successfully updates only cost", func(t *testing.T) {
		originalCost := 100
		originalIsAvailable := true
		newCost := 120

		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &originalCost,
			IsAvailable:      originalIsAvailable,
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}

		input := &common.LeaguePokemonUpdateRequest{
			LeaguePokemonID: testLeaguePokemonID,
			Cost:            &newCost,
			IsAvailable:     originalIsAvailable, // Same as original
		}

		updatedLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &newCost,
			IsAvailable:      originalIsAvailable,
		}

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockLeaguePokemonRepo.On("UpdateLeaguePokemon", updatedLeaguePokemon).Return(updatedLeaguePokemon, nil).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.NoError(t, err)
		assert.Equal(t, updatedLeaguePokemon, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Successfully updates only availability", func(t *testing.T) {
		originalCost := 100
		originalIsAvailable := true
		newIsAvailable := false

		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &originalCost,
			IsAvailable:      originalIsAvailable,
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}

		input := &common.LeaguePokemonUpdateRequest{
			LeaguePokemonID: testLeaguePokemonID,
			Cost:            &originalCost,
			IsAvailable:     newIsAvailable,
		}

		updatedLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &originalCost,
			IsAvailable:      newIsAvailable,
		}

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockLeaguePokemonRepo.On("UpdateLeaguePokemon", updatedLeaguePokemon).Return(updatedLeaguePokemon, nil).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.NoError(t, err)
		assert.Equal(t, updatedLeaguePokemon, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fails if league pokemon not found", func(t *testing.T) {
		input := &common.LeaguePokemonUpdateRequest{LeaguePokemonID: testLeaguePokemonID}
		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return((*models.LeaguePokemon)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.ErrorIs(t, err, common.ErrLeaguePokemonNotFound)
		assert.Nil(t, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertNotCalled(t, "GetLeagueByID")
	})

	t.Run("Fails if getting league pokemon returns internal error", func(t *testing.T) {
		internalErr := errors.New("db error")
		input := &common.LeaguePokemonUpdateRequest{LeaguePokemonID: testLeaguePokemonID}
		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return((*models.LeaguePokemon)(nil), internalErr).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertNotCalled(t, "GetLeagueByID")
	})

	t.Run("Fails if league not found (unreachable code path)", func(t *testing.T) {
		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &[]int{100}[0],
			IsAvailable:      true,
		}
		input := &common.LeaguePokemonUpdateRequest{LeaguePokemonID: testLeaguePokemonID}

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return((*models.League)(nil), gorm.ErrRecordNotFound).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.ErrorIs(t, err, common.ErrLeagueNotFound)
		assert.Nil(t, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fails if getting league returns internal error", func(t *testing.T) {
		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &[]int{100}[0],
			IsAvailable:      true,
		}
		internalErr := errors.New("db error")
		input := &common.LeaguePokemonUpdateRequest{LeaguePokemonID: testLeaguePokemonID}

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return((*models.League)(nil), internalErr).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fails if league is not in Setup or Drafting status", func(t *testing.T) {
		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &[]int{100}[0],
			IsAvailable:      true,
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusCompleted} // Not Setup or Drafting
		input := &common.LeaguePokemonUpdateRequest{LeaguePokemonID: testLeaguePokemonID}

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInvalidState)
		assert.Nil(t, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
		mockLeaguePokemonRepo.AssertNotCalled(t, "UpdateLeaguePokemon")
	})

	t.Run("Fails if update operation returns internal error", func(t *testing.T) {
		originalCost := 100
		originalIsAvailable := true
		newCost := 120
		newIsAvailable := false

		existingLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &originalCost,
			IsAvailable:      originalIsAvailable,
		}
		league := &models.League{ID: testLeagueID, Status: enums.LeagueStatusSetup}

		input := &common.LeaguePokemonUpdateRequest{
			LeaguePokemonID: testLeaguePokemonID,
			Cost:            &newCost,
			IsAvailable:     newIsAvailable,
		}

		updatedLeaguePokemon := &models.LeaguePokemon{
			ID:               testLeaguePokemonID,
			LeagueID:         testLeagueID,
			PokemonSpeciesID: testPokemonSpeciesID,
			Cost:             &newCost,
			IsAvailable:      newIsAvailable,
		}
		internalErr := errors.New("db update error")

		mockLeaguePokemonRepo.On("GetLeaguePokemonByID", testLeaguePokemonID).Return(existingLeaguePokemon, nil).Once()
		mockLeagueRepo.On("GetLeagueByID", testLeagueID).Return(league, nil).Once()
		mockLeaguePokemonRepo.On("UpdateLeaguePokemon", updatedLeaguePokemon).Return((*models.LeaguePokemon)(nil), internalErr).Once()

		result, err := service.UpdateLeaguePokemon(currentUser, input)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, result)

		mockLeaguePokemonRepo.AssertExpectations(t)
		mockLeagueRepo.AssertExpectations(t)
	})
}
