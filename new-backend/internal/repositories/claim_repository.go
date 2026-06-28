package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ClaimRepository defines the interface for managing ownership claims on Pokemon species.
type ClaimRepository interface {
	// CreateClaim creates a new claim record.
	CreateClaim(claim *models.Claim) (*models.Claim, error)
	// GetClaimByID retrieves a claim by its ID with relationships.
	GetClaimByID(id uuid.UUID) (*models.Claim, error)
	// GetActiveClaimByPlayerAndSpecies checks if a player has an active claim for a species.
	GetActiveClaimByPlayerAndSpecies(playerID uuid.UUID, speciesID int64) (*models.Claim, error)
	// GetActiveClaimsByPlayer retrieves all active claims for a player in a league.
	GetActiveClaimsByPlayer(playerID uuid.UUID) ([]models.Claim, error)
	// GetActiveClaimsByLeague retrieves all active claims in a league.
	GetActiveClaimsByLeague(leagueID uuid.UUID) ([]models.Claim, error)
	// GetReleasedClaimsByLeague retrieves all released (inactive) claims in a league.
	GetReleasedClaimsByLeague(leagueID uuid.UUID) ([]models.Claim, error)
	// GetActiveClaimCountByPlayer returns the count of active claims for a player.
	GetActiveClaimCountByPlayer(playerID uuid.UUID) (int64, error)
	// GetActiveClaimCountByLeague returns the count of active claims in a league.
	GetActiveClaimCountByLeague(leagueID uuid.UUID) (int64, error)
	// IsSpeciesClaimedInLeague checks if a Pokemon species has an active claim in a league.
	IsSpeciesClaimedInLeague(leagueID uuid.UUID, speciesID int64) (bool, error)
	// UpdateClaim updates an existing claim record.
	UpdateClaim(claim *models.Claim) (*models.Claim, error)
	// ReleaseClaimTransaction performs a transactional release of a claim:
	// marks claim as inactive, sets released week, returns pokemon to pool.
	ReleaseClaimTransaction(claim *models.Claim, player *models.Player, dropCost int, releasedWeek int, poolEntryID uuid.UUID) error
	// PickupFreeAgentTransaction performs a transactional free-agent pickup:
	// creates a new claim, deducts credits, marks pool entry unavailable.
	PickupFreeAgentTransaction(player *models.Player, newClaim *models.Claim, poolEntry *models.PoolEntry, pickupCost int) error
}

type claimRepositoryImpl struct {
	db *gorm.DB
}

func NewClaimRepository(db *gorm.DB) ClaimRepository {
	return &claimRepositoryImpl{db: db}
}

func (r *claimRepositoryImpl) CreateClaim(claim *models.Claim) (*models.Claim, error) {
	if err := r.db.Create(claim).Error; err != nil {
		return nil, fmt.Errorf("(Error: CreateClaim) - failed to create claim: %w", err)
	}
	return claim, nil
}

func (r *claimRepositoryImpl) GetClaimByID(id uuid.UUID) (*models.Claim, error) {
	var claim models.Claim
	err := r.db.Preload("League").
		Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		First(&claim, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetClaimByID) - failed to get claim: %w", err)
	}
	return &claim, nil
}

func (r *claimRepositoryImpl) GetActiveClaimByPlayerAndSpecies(playerID uuid.UUID, speciesID int64) (*models.Claim, error) {
	var claim models.Claim
	err := r.db.Where("player_id = ? AND species_id = ? AND is_active = ?", playerID, speciesID, true).
		First(&claim).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("(Error: GetActiveClaimByPlayerAndSpecies) - failed to get active claim: %w", err)
	}
	return &claim, nil
}

func (r *claimRepositoryImpl) GetActiveClaimsByPlayer(playerID uuid.UUID) ([]models.Claim, error) {
	var claims []models.Claim
	err := r.db.Preload("PokemonSpecies").
		Where("player_id = ? AND is_active = ?", playerID, true).
		Order("created_at ASC").
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetActiveClaimsByPlayer) - failed to get active claims: %w", err)
	}
	return claims, nil
}

func (r *claimRepositoryImpl) GetActiveClaimsByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	var claims []models.Claim
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ? AND is_active = ?", leagueID, true).
		Order("created_at ASC").
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetActiveClaimsByLeague) - failed to get active claims: %w", err)
	}
	return claims, nil
}

func (r *claimRepositoryImpl) GetReleasedClaimsByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	var claims []models.Claim
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ? AND is_active = ?", leagueID, false).
		Order("updated_at DESC").
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetReleasedClaimsByLeague) - failed to get released claims: %w", err)
	}
	return claims, nil
}

func (r *claimRepositoryImpl) GetActiveClaimCountByPlayer(playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Claim{}).
		Where("player_id = ? AND is_active = ?", playerID, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: GetActiveClaimCountByPlayer) - failed to count active claims: %w", err)
	}
	return count, nil
}

func (r *claimRepositoryImpl) GetActiveClaimCountByLeague(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Claim{}).
		Where("league_id = ? AND is_active = ?", leagueID, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: GetActiveClaimCountByLeague) - failed to count active claims: %w", err)
	}
	return count, nil
}

func (r *claimRepositoryImpl) IsSpeciesClaimedInLeague(leagueID uuid.UUID, speciesID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Claim{}).
		Where("league_id = ? AND species_id = ? AND is_active = ?", leagueID, speciesID, true).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("(Error: IsSpeciesClaimedInLeague) - failed to check species claim: %w", err)
	}
	return count > 0, nil
}

func (r *claimRepositoryImpl) UpdateClaim(claim *models.Claim) (*models.Claim, error) {
	if err := r.db.Save(claim).Error; err != nil {
		return nil, fmt.Errorf("(Error: UpdateClaim) - failed to update claim: %w", err)
	}
	return claim, nil
}

// ReleaseClaimTransaction performs a transactional release of a claim:
// marks claim as inactive, sets released week, returns pokemon to pool.
func (r *claimRepositoryImpl) ReleaseClaimTransaction(claim *models.Claim, player *models.Player, dropCost int, releasedWeek int, poolEntryID uuid.UUID) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: ReleaseClaimTransaction) - failed to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Mark the claim as released
	if err := tx.Model(claim).Updates(map[string]any{
		"is_active":     false,
		"released_week": releasedWeek,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: ReleaseClaimTransaction) - failed to release claim: %w", err)
	}

	// 2. Mark the pool entry as available again
	if poolEntryID != uuid.Nil {
		if err := tx.Model(&models.PoolEntry{}).Where("id = ?", poolEntryID).Update("is_available", true).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("(Error: ReleaseClaimTransaction) - failed to update pool entry availability: %w", err)
		}
	}

	// 3. Decrement player's TransferCredits
	player.TransferCredits -= dropCost
	if err := tx.Save(player).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: ReleaseClaimTransaction) - failed to update player transfer credits: %w", err)
	}

	return tx.Commit().Error
}

// PickupFreeAgentTransaction performs a transactional free-agent pickup:
// creates a new claim, deducts credits, marks pool entry unavailable.
func (r *claimRepositoryImpl) PickupFreeAgentTransaction(player *models.Player, newClaim *models.Claim, poolEntry *models.PoolEntry, pickupCost int) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Decrement player's TransferCredits
	player.TransferCredits -= pickupCost
	if err := tx.Save(player).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to update player transfer credits: %w", err)
	}

	// 2. Create new claim entry
	if err := tx.Create(newClaim).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to create claim: %w", err)
	}

	// 3. Mark pool entry as unavailable
	if err := tx.Model(poolEntry).Update("is_available", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("(Error: PickupFreeAgentTransaction) - failed to update pool entry availability: %w", err)
	}

	return tx.Commit().Error
}
