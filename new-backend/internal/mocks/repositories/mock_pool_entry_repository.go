package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockPoolEntryRepository struct {
	mock.Mock
}

func (m *MockPoolEntryRepository) GetByID(id uuid.UUID) (*models.PoolEntry, error) {
	args := m.Called(id)
	var result *models.PoolEntry
	if args.Get(0) != nil {
		result = args.Get(0).(*models.PoolEntry)
	}
	return result, args.Error(1)
}

func (m *MockPoolEntryRepository) GetByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.PoolEntry), args.Error(1)
}

func (m *MockPoolEntryRepository) GetAvailableByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.PoolEntry), args.Error(1)
}

func (m *MockPoolEntryRepository) GetByIDs(leagueID uuid.UUID, ids []uuid.UUID) ([]models.PoolEntry, error) {
	args := m.Called(leagueID, ids)
	return args.Get(0).([]models.PoolEntry), args.Error(1)
}

func (m *MockPoolEntryRepository) GetBySpecies(leagueID uuid.UUID, speciesID int64) (*models.PoolEntry, error) {
	args := m.Called(leagueID, speciesID)
	var result *models.PoolEntry
	if args.Get(0) != nil {
		result = args.Get(0).(*models.PoolEntry)
	}
	return result, args.Error(1)
}

func (m *MockPoolEntryRepository) GetByCostRange(leagueID uuid.UUID, minCost, maxCost int) ([]models.PoolEntry, error) {
	args := m.Called(leagueID, minCost, maxCost)
	return args.Get(0).([]models.PoolEntry), args.Error(1)
}

func (m *MockPoolEntryRepository) IsAvailable(leagueID uuid.UUID, speciesID int64) (bool, error) {
	args := m.Called(leagueID, speciesID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPoolEntryRepository) GetCost(leagueID uuid.UUID, speciesID int64) (*int, error) {
	args := m.Called(leagueID, speciesID)
	var result *int
	if args.Get(0) != nil {
		result = args.Get(0).(*int)
	}
	return result, args.Error(1)
}

func (m *MockPoolEntryRepository) GetAvailableCount(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPoolEntryRepository) Create(entry *models.PoolEntry) (*models.PoolEntry, error) {
	args := m.Called(entry)
	var result *models.PoolEntry
	if args.Get(0) != nil {
		result = args.Get(0).(*models.PoolEntry)
	}
	return result, args.Error(1)
}

func (m *MockPoolEntryRepository) CreateBatch(entries []models.PoolEntry) ([]models.PoolEntry, error) {
	args := m.Called(entries)
	return args.Get(0).([]models.PoolEntry), args.Error(1)
}

func (m *MockPoolEntryRepository) Update(entry *models.PoolEntry) (*models.PoolEntry, error) {
	args := m.Called(entry)
	var result *models.PoolEntry
	if args.Get(0) != nil {
		result = args.Get(0).(*models.PoolEntry)
	}
	return result, args.Error(1)
}

func (m *MockPoolEntryRepository) MarkUnavailable(tx *gorm.DB, id uuid.UUID) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockPoolEntryRepository) MarkAvailable(tx *gorm.DB, id uuid.UUID) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockPoolEntryRepository) Delete(leagueID uuid.UUID, speciesID int64) error {
	args := m.Called(leagueID, speciesID)
	return args.Error(0)
}

func (m *MockPoolEntryRepository) DeleteAllByLeague(leagueID uuid.UUID) error {
	args := m.Called(leagueID)
	return args.Error(0)
}
