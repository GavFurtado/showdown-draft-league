package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockLeaguePokemonRepository struct {
	mock.Mock
}

func (m *MockLeaguePokemonRepository) GetLeaguePokemonByIDs(leagueID uuid.UUID, leaguePokemonIDs []uuid.UUID) ([]models.LeaguePokemon, error) {
	args := m.Called(leagueID, leaguePokemonIDs)
	var result []models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).([]models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) CreateLeaguePokemon(leaguePokemon *models.LeaguePokemon) (*models.LeaguePokemon, error) {
	args := m.Called(leaguePokemon)
	var result *models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) CreateLeaguePokemonBatch(leaguePokemon []models.LeaguePokemon) ([]models.LeaguePokemon, error) {
	args := m.Called(leaguePokemon)
	var result []models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).([]models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) GetAllPokemonByLeague(leagueID uuid.UUID) ([]models.LeaguePokemon, error) {
	args := m.Called(leagueID)
	var result []models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).([]models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) GetAvailablePokemonByLeague(leagueID uuid.UUID) ([]models.LeaguePokemon, error) {
	args := m.Called(leagueID)
	var result []models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).([]models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) GetLeaguePokemonByID(id uuid.UUID) (*models.LeaguePokemon, error) {
	args := m.Called(id)
	var result *models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) UpdateLeaguePokemon(leaguePokemon *models.LeaguePokemon) (*models.LeaguePokemon, error) {
	args := m.Called(leaguePokemon)
	var result *models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) MarkPokemonUnavailable(leagueID, pokemonSpeciesID uuid.UUID) error {
	args := m.Called(leagueID, pokemonSpeciesID)
	return args.Error(0)
}

func (m *MockLeaguePokemonRepository) GetPokemonByCostRange(leagueID uuid.UUID, minCost, maxCost int) ([]models.LeaguePokemon, error) {
	args := m.Called(leagueID, minCost, maxCost)
	var result []models.LeaguePokemon
	if args.Get(0) != nil {
		result = args.Get(0).([]models.LeaguePokemon)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) IsPokemonAvailable(leagueID, pokemonSpeciesID uuid.UUID) (bool, error) {
	args := m.Called(leagueID, pokemonSpeciesID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLeaguePokemonRepository) GetPokemonCost(leagueID, pokemonSpeciesID uuid.UUID) (*int, error) {
	args := m.Called(leagueID, pokemonSpeciesID)
	var result *int
	if args.Get(0) != nil {
		result = args.Get(0).(*int)
	}
	return result, args.Error(1)
}

func (m *MockLeaguePokemonRepository) DeleteLeaguePokemon(leagueID, pokemonSpeciesID uuid.UUID) error {
	args := m.Called(leagueID, pokemonSpeciesID)
	return args.Error(0)
}

func (m *MockLeaguePokemonRepository) GetAvailablePokemonCount(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLeaguePokemonRepository) DeleteAllLeaguePokemon(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}