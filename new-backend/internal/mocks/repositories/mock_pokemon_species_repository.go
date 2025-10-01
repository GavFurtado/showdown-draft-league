package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockPokemonSpeciesRepository struct {
	mock.Mock
}

func (m *MockPokemonSpeciesRepository) GetAllPokemonSpecies() ([]models.PokemonSpecies, error) {
	args := m.Called()
	var result []models.PokemonSpecies
	if args.Get(0) != nil {
		result = args.Get(0).([]models.PokemonSpecies)
	}
	return result, args.Error(1)
}

func (m *MockPokemonSpeciesRepository) GetPokemonSpeciesByID(id int64) (*models.PokemonSpecies, error) {
	args := m.Called(id)
	var result *models.PokemonSpecies
	if args.Get(0) != nil {
		result = args.Get(0).(*models.PokemonSpecies)
	}
	return result, args.Error(1)
}

func (m *MockPokemonSpeciesRepository) GetPokemonSpeciesByName(name string) (*models.PokemonSpecies, error) {
	args := m.Called(name)
	var result *models.PokemonSpecies
	if args.Get(0) != nil {
		result = args.Get(0).(*models.PokemonSpecies)
	}
	return result, args.Error(1)
}

func (m *MockPokemonSpeciesRepository) FindPokemonSpecies(filter string) ([]models.PokemonSpecies, error) {
	args := m.Called(filter)
	var result []models.PokemonSpecies
	if args.Get(0) != nil {
		result = args.Get(0).([]models.PokemonSpecies)
	}
	return result, args.Error(1)
}

func (m *MockPokemonSpeciesRepository) CreatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	args := m.Called(pokemon)
	return args.Error(0)
}

func (m *MockPokemonSpeciesRepository) UpdatePokemonSpecies(pokemon *models.PokemonSpecies) error {
	args := m.Called(pokemon)
	return args.Error(0)
}

func (m *MockPokemonSpeciesRepository) DeletePokemonSpecies(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}
