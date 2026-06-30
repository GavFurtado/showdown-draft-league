package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockClaimRepository struct {
	mock.Mock
}

func (m *MockClaimRepository) Create(claim *models.Claim) (*models.Claim, error) {
	args := m.Called(claim)
	var result *models.Claim
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Claim)
	}
	return result, args.Error(1)
}

func (m *MockClaimRepository) GetByID(id uuid.UUID) (*models.Claim, error) {
	args := m.Called(id)
	var result *models.Claim
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Claim)
	}
	return result, args.Error(1)
}

func (m *MockClaimRepository) GetActiveByPlayerAndSpecies(playerID uuid.UUID, speciesID int64) (*models.Claim, error) {
	args := m.Called(playerID, speciesID)
	var result *models.Claim
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Claim)
	}
	return result, args.Error(1)
}

func (m *MockClaimRepository) GetActiveByPlayer(playerID uuid.UUID) ([]models.Claim, error) {
	args := m.Called(playerID)
	return args.Get(0).([]models.Claim), args.Error(1)
}

func (m *MockClaimRepository) GetActiveByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.Claim), args.Error(1)
}

func (m *MockClaimRepository) GetReleasedByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.Claim), args.Error(1)
}

func (m *MockClaimRepository) GetActiveCountByPlayer(playerID uuid.UUID) (int64, error) {
	args := m.Called(playerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClaimRepository) GetActiveCountByLeague(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockClaimRepository) IsSpeciesClaimedInLeague(leagueID uuid.UUID, speciesID int64) (bool, error) {
	args := m.Called(leagueID, speciesID)
	return args.Bool(0), args.Error(1)
}

func (m *MockClaimRepository) Update(claim *models.Claim) (*models.Claim, error) {
	args := m.Called(claim)
	var result *models.Claim
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Claim)
	}
	return result, args.Error(1)
}

func (m *MockClaimRepository) ReleaseTx(tx *gorm.DB, claim *models.Claim, member *models.LeagueMember, dropCost int, releasedWeek int, poolEntryID uuid.UUID) error {
	args := m.Called(tx, claim, member, dropCost, releasedWeek, poolEntryID)
	return args.Error(0)
}

func (m *MockClaimRepository) PickupFreeAgentTx(tx *gorm.DB, member *models.LeagueMember, newClaim *models.Claim, poolEntry *models.PoolEntry, pickupCost int) error {
	args := m.Called(tx, member, newClaim, poolEntry, pickupCost)
	return args.Error(0)
}
