package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClaimRepository interface {
	Create(claim *models.Claim) (*models.Claim, error)
	GetByID(id uuid.UUID) (*models.Claim, error)
	GetActiveByPlayerAndSpecies(playerID uuid.UUID, speciesID int64) (*models.Claim, error)
	GetActiveByPlayer(playerID uuid.UUID) ([]models.Claim, error)
	GetActiveByLeague(leagueID uuid.UUID) ([]models.Claim, error)
	GetReleasedByLeague(leagueID uuid.UUID) ([]models.Claim, error)
	GetActiveCountByPlayer(playerID uuid.UUID) (int64, error)
	GetActiveCountByLeague(leagueID uuid.UUID) (int64, error)
	IsSpeciesClaimedInLeague(leagueID uuid.UUID, speciesID int64) (bool, error)
	Update(claim *models.Claim) (*models.Claim, error)
	ReleaseTx(tx *gorm.DB, claim *models.Claim, member *models.Player, dropCost int, releasedWeek int, poolEntryID uuid.UUID) error
	PickupFreeAgentTx(tx *gorm.DB, member *models.Player, newClaim *models.Claim, poolEntry *models.PoolEntry, pickupCost int) error
}

type claimRepositoryImpl struct {
	db *gorm.DB
}

func NewClaimRepository(db *gorm.DB) ClaimRepository {
	return &claimRepositoryImpl{db: db}
}

func (r *claimRepositoryImpl) Create(claim *models.Claim) (*models.Claim, error) {
	if err := r.db.Create(claim).Error; err != nil {
		return nil, fmt.Errorf("(Error: ClaimRepo.Create) - failed to create claim: %w", err)
	}
	return claim, nil
}

func (r *claimRepositoryImpl) GetByID(id uuid.UUID) (*models.Claim, error) {
	var claim models.Claim
	err := r.db.Preload("League").
		Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		First(&claim, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: ClaimRepo.GetByID) - failed to get claim: %w", err)
	}
	return &claim, nil
}

func (r *claimRepositoryImpl) GetActiveByPlayerAndSpecies(playerID uuid.UUID, speciesID int64) (*models.Claim, error) {
	var claim models.Claim
	err := r.db.Where("player_id = ? AND species_id = ? AND is_active = ?", playerID, speciesID, true).
		First(&claim).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("(Error: ClaimRepo.GetActiveByPlayerAndSpecies) - failed: %w", err)
	}
	return &claim, nil
}

func (r *claimRepositoryImpl) GetActiveByPlayer(playerID uuid.UUID) ([]models.Claim, error) {
	var claims []models.Claim
	err := r.db.Preload("PokemonSpecies").
		Where("player_id = ? AND is_active = ?", playerID, true).
		Order("created_at ASC").
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: ClaimRepo.GetActiveByPlayer) - failed: %w", err)
	}
	return claims, nil
}

func (r *claimRepositoryImpl) GetActiveByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	var claims []models.Claim
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ? AND is_active = ?", leagueID, true).
		Order("created_at ASC").
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: ClaimRepo.GetActiveByLeague) - failed: %w", err)
	}
	return claims, nil
}

func (r *claimRepositoryImpl) GetReleasedByLeague(leagueID uuid.UUID) ([]models.Claim, error) {
	var claims []models.Claim
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PokemonSpecies").
		Where("league_id = ? AND is_active = ?", leagueID, false).
		Order("updated_at DESC").
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: ClaimRepo.GetReleasedByLeague) - failed: %w", err)
	}
	return claims, nil
}

func (r *claimRepositoryImpl) GetActiveCountByPlayer(playerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Claim{}).
		Where("player_id = ? AND is_active = ?", playerID, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: ClaimRepo.GetActiveCountByPlayer) - failed: %w", err)
	}
	return count, nil
}

func (r *claimRepositoryImpl) GetActiveCountByLeague(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Claim{}).
		Where("league_id = ? AND is_active = ?", leagueID, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: ClaimRepo.GetActiveCountByLeague) - failed: %w", err)
	}
	return count, nil
}

func (r *claimRepositoryImpl) IsSpeciesClaimedInLeague(leagueID uuid.UUID, speciesID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Claim{}).
		Where("league_id = ? AND species_id = ? AND is_active = ?", leagueID, speciesID, true).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("(Error: ClaimRepo.IsSpeciesClaimedInLeague) - failed: %w", err)
	}
	return count > 0, nil
}

func (r *claimRepositoryImpl) Update(claim *models.Claim) (*models.Claim, error) {
	if err := r.db.Save(claim).Error; err != nil {
		return nil, fmt.Errorf("(Error: ClaimRepo.Update) - failed: %w", err)
	}
	return claim, nil
}

func (r *claimRepositoryImpl) ReleaseTx(tx *gorm.DB, claim *models.Claim, member *models.Player, dropCost int, releasedWeek int, poolEntryID uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	if err := db.Model(claim).Updates(map[string]any{
		"is_active":     false,
		"released_week": releasedWeek,
	}).Error; err != nil {
		return fmt.Errorf("(Error: ClaimRepo.ReleaseTx) - failed to release claim: %w", err)
	}

	if poolEntryID != uuid.Nil {
		if err := db.Model(&models.PoolEntry{}).Where("id = ?", poolEntryID).Update("is_available", true).Error; err != nil {
			return fmt.Errorf("(Error: ClaimRepo.ReleaseTx) - failed to update pool entry: %w", err)
		}
	}

	member.TransferCredits -= dropCost
	if err := db.Save(member).Error; err != nil {
		return fmt.Errorf("(Error: ClaimRepo.ReleaseTx) - failed to update member credits: %w", err)
	}

	return nil
}

func (r *claimRepositoryImpl) PickupFreeAgentTx(tx *gorm.DB, member *models.Player, newClaim *models.Claim, poolEntry *models.PoolEntry, pickupCost int) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	member.TransferCredits -= pickupCost
	if err := db.Save(member).Error; err != nil {
		return fmt.Errorf("(Error: ClaimRepo.PickupFreeAgentTx) - failed to update member credits: %w", err)
	}

	if err := db.Create(newClaim).Error; err != nil {
		return fmt.Errorf("(Error: ClaimRepo.PickupFreeAgentTx) - failed to create claim: %w", err)
	}

	if err := db.Model(poolEntry).Update("is_available", false).Error; err != nil {
		return fmt.Errorf("(Error: ClaimRepo.PickupFreeAgentTx) - failed to update pool entry: %w", err)
	}

	return nil
}
