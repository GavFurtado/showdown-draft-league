package services_test

import (
	"errors"
	"testing"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/mocks/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPokemonSpeciesService_GetPokemonSpeciesByID(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	pokemonID := int64(1)

	t.Run("Successfully gets pokemon species by ID", func(t *testing.T) {
		expectedPokemon := &models.PokemonSpecies{ID: pokemonID, Name: "Pikachu"}
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return(expectedPokemon, nil).Once()

		pokemon, err := service.GetPokemonSpeciesByID(pokemonID)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrPokemonSpeciesNotFound if not found", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()

		pokemon, err := service.GetPokemonSpeciesByID(pokemonID)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for other repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		pokemon, err := service.GetPokemonSpeciesByID(pokemonID)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for invalid ID", func(t *testing.T) {
		pokemon, err := service.GetPokemonSpeciesByID(0)
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByID")
	})
}

func TestPokemonSpeciesService_GetPokemonSpeciesByName(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	pokemonName := "Pikachu"

	t.Run("Successfully gets pokemon species by name", func(t *testing.T) {
		expectedPokemon := &models.PokemonSpecies{ID: 1, Name: pokemonName}
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", pokemonName).Return(expectedPokemon, nil).Once()

		pokemon, err := service.GetPokemonSpeciesByName(pokemonName)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrPokemonSpeciesNotFound if not found", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", pokemonName).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()

		pokemon, err := service.GetPokemonSpeciesByName(pokemonName)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for other repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", pokemonName).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		pokemon, err := service.GetPokemonSpeciesByName(pokemonName)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for empty name", func(t *testing.T) {
		pokemon, err := service.GetPokemonSpeciesByName("")
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "GetPokemonSpeciesByName")
	})
}

func TestPokemonSpeciesService_GetAllPokemonSpecies(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	t.Run("Successfully gets all pokemon species", func(t *testing.T) {
		expectedPokemon := []common.PokemonSpeciesListDTO{
			{ID: 1, Name: "Pikachu", Types: []string{"electric"}},
			{ID: 2, Name: "Charmander", Types: []string{"fire"}},
		}

		mockPokemonSpeciesRepo.On("GetAllPokemonSpecies").Return([]models.PokemonSpecies{
			{ID: 1, Name: "Pikachu", Types: models.StringArray{"electric"}},
			{ID: 2, Name: "Charmander", Types: models.StringArray{"fire"}},
		}, nil).Once()

		pokemon, err := service.GetAllPokemonSpecies()
		assert.NoError(t, err)

		assert.Equal(t, expectedPokemon, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns empty slice if no pokemon found", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetAllPokemonSpecies").Return([]models.PokemonSpecies{}, nil).Once()

		pokemon, err := service.GetAllPokemonSpecies()
		assert.NoError(t, err)
		assert.Empty(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockPokemonSpeciesRepo.On("GetAllPokemonSpecies").Return(([]models.PokemonSpecies)(nil), internalErr).Once()

		pokemon, err := service.GetAllPokemonSpecies()
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})
}

func TestPokemonSpeciesService_ListPokemonSpecies(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	filter := "Pika"

	t.Run("Successfully lists pokemon species with filter", func(t *testing.T) {
		expectedPokemon := []models.PokemonSpecies{
			{ID: 1, Name: "Pikachu"},
		}
		mockPokemonSpeciesRepo.On("FindPokemonSpecies", filter).Return(expectedPokemon, nil).Once()

		pokemon, err := service.ListPokemonSpecies(filter)
		assert.NoError(t, err)
		assert.Equal(t, expectedPokemon, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns empty slice if no pokemon found with filter", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("FindPokemonSpecies", filter).Return([]models.PokemonSpecies{}, nil).Once()

		pokemon, err := service.ListPokemonSpecies(filter)
		assert.NoError(t, err)
		assert.Empty(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInternalService for repository errors", func(t *testing.T) {
		internalErr := errors.New("database error")
		mockPokemonSpeciesRepo.On("FindPokemonSpecies", filter).Return(([]models.PokemonSpecies)(nil), internalErr).Once()

		pokemon, err := service.ListPokemonSpecies(filter)
		assert.ErrorIs(t, err, common.ErrInternalService)
		assert.Nil(t, pokemon)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})
}

func TestPokemonSpeciesService_CreatePokemonSpecies(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	newPokemon := &models.PokemonSpecies{ID: 3, Name: "Bulbasaur"}

	t.Run("Successfully creates pokemon species", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", newPokemon.ID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", newPokemon.Name).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()
		mockPokemonSpeciesRepo.On("CreatePokemonSpecies", newPokemon).Return(nil).Once()

		err := service.CreatePokemonSpecies(newPokemon)
		assert.NoError(t, err)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for invalid pokemon", func(t *testing.T) {
		err := service.CreatePokemonSpecies(nil)
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")

		err = service.CreatePokemonSpecies(&models.PokemonSpecies{ID: 0, Name: "Invalid"})
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")

		err = service.CreatePokemonSpecies(&models.PokemonSpecies{ID: 1, Name: ""})
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")
	})

	t.Run("Returns ErrConflict if pokemon with same ID exists", func(t *testing.T) {
		existingPokemon := &models.PokemonSpecies{ID: newPokemon.ID, Name: "Existing"}
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", newPokemon.ID).Return(existingPokemon, nil).Once()

		err := service.CreatePokemonSpecies(newPokemon)
		assert.ErrorIs(t, err, common.ErrConflict)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")
	})

	t.Run("Returns ErrConflict if pokemon with same name exists", func(t *testing.T) {
		existingPokemon := &models.PokemonSpecies{ID: 99, Name: newPokemon.Name}
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", newPokemon.ID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", newPokemon.Name).Return(existingPokemon, nil).Once()

		err := service.CreatePokemonSpecies(newPokemon)
		assert.ErrorIs(t, err, common.ErrConflict)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")
	})

	t.Run("Returns internal error if GetPokemonSpeciesByID fails", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", newPokemon.ID).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		err := service.CreatePokemonSpecies(newPokemon)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")
	})

	t.Run("Returns internal error if GetPokemonSpeciesByName fails", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", newPokemon.ID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", newPokemon.Name).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		err := service.CreatePokemonSpecies(newPokemon)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "CreatePokemonSpecies")
	})

	t.Run("Returns internal error if CreatePokemonSpecies fails", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", newPokemon.ID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByName", newPokemon.Name).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()
		mockPokemonSpeciesRepo.On("CreatePokemonSpecies", newPokemon).Return(internalErr).Once()

		err := service.CreatePokemonSpecies(newPokemon)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})
}

func TestPokemonSpeciesService_UpdatePokemonSpecies(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	updatedPokemon := &models.PokemonSpecies{ID: 1, Name: "UpdatedPikachu"}

	t.Run("Successfully updates pokemon species", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", updatedPokemon.ID).Return(&models.PokemonSpecies{ID: 1, Name: "Pikachu"}, nil).Once()
		mockPokemonSpeciesRepo.On("UpdatePokemonSpecies", updatedPokemon).Return(nil).Once()

		err := service.UpdatePokemonSpecies(updatedPokemon)
		assert.NoError(t, err)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for invalid pokemon", func(t *testing.T) {
		err := service.UpdatePokemonSpecies(nil)
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "UpdatePokemonSpecies")

		err = service.UpdatePokemonSpecies(&models.PokemonSpecies{ID: 0, Name: "Invalid"})
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "UpdatePokemonSpecies")
	})

	t.Run("Returns ErrPokemonSpeciesNotFound if pokemon not found", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", updatedPokemon.ID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()

		err := service.UpdatePokemonSpecies(updatedPokemon)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "UpdatePokemonSpecies")
	})

	t.Run("Returns internal error if GetPokemonSpeciesByID fails", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", updatedPokemon.ID).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		err := service.UpdatePokemonSpecies(updatedPokemon)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "UpdatePokemonSpecies")
	})

	t.Run("Returns internal error if UpdatePokemonSpecies fails", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", updatedPokemon.ID).Return(&models.PokemonSpecies{ID: 1, Name: "Pikachu"}, nil).Once()
		internalErr := errors.New("db update error")
		mockPokemonSpeciesRepo.On("UpdatePokemonSpecies", updatedPokemon).Return(internalErr).Once()

		err := service.UpdatePokemonSpecies(updatedPokemon)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})
}

func TestPokemonSpeciesService_DeletePokemonSpecies(t *testing.T) {
	mockPokemonSpeciesRepo := new(mock_repositories.MockPokemonSpeciesRepository)
	service := services.NewPokemonSpeciesService(mockPokemonSpeciesRepo)

	pokemonID := int64(1)

	t.Run("Successfully deletes pokemon species", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return(&models.PokemonSpecies{ID: pokemonID, Name: "Pikachu"}, nil).Once()
		mockPokemonSpeciesRepo.On("DeletePokemonSpecies", pokemonID).Return(nil).Once()

		err := service.DeletePokemonSpecies(pokemonID)
		assert.NoError(t, err)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})

	t.Run("Returns ErrInvalidInput for invalid ID", func(t *testing.T) {
		err := service.DeletePokemonSpecies(0)
		assert.ErrorIs(t, err, common.ErrInvalidInput)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "DeletePokemonSpecies")
	})

	t.Run("Returns ErrPokemonSpeciesNotFound if pokemon not found", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return((*models.PokemonSpecies)(nil), gorm.ErrRecordNotFound).Once()

		err := service.DeletePokemonSpecies(pokemonID)
		assert.ErrorIs(t, err, common.ErrPokemonSpeciesNotFound)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "DeletePokemonSpecies")
	})

	t.Run("Returns internal error if GetPokemonSpeciesByID fails", func(t *testing.T) {
		internalErr := errors.New("db error")
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return((*models.PokemonSpecies)(nil), internalErr).Once()

		err := service.DeletePokemonSpecies(pokemonID)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
		mockPokemonSpeciesRepo.AssertNotCalled(t, "DeletePokemonSpecies")
	})

	t.Run("Returns internal error if DeletePokemonSpecies fails", func(t *testing.T) {
		mockPokemonSpeciesRepo.On("GetPokemonSpeciesByID", pokemonID).Return(&models.PokemonSpecies{ID: pokemonID, Name: "Pikachu"}, nil).Once()
		internalErr := errors.New("db delete error")
		mockPokemonSpeciesRepo.On("DeletePokemonSpecies", pokemonID).Return(internalErr).Once()

		err := service.DeletePokemonSpecies(pokemonID)
		assert.ErrorIs(t, err, common.ErrInternalService)
		mockPokemonSpeciesRepo.AssertExpectations(t)
	})
}
