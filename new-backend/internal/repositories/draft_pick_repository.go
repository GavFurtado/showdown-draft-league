package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DraftPickRepository interface {
	Create(pick *models.DraftPick) (*models.DraftPick, error)
	CreateBatch(picks []models.DraftPick) error
	GetByID(id uuid.UUID) (*models.DraftPick, error)
	GetByDraft(draftID uuid.UUID) ([]models.DraftPick, error)
	GetByPlayer(playerID uuid.UUID) ([]models.DraftPick, error)
	GetCountByDraft(draftID uuid.UUID) (int64, error)
}

type draftPickRepositoryImpl struct {
	db *gorm.DB
}

func NewDraftPickRepository(db *gorm.DB) DraftPickRepository {
	return &draftPickRepositoryImpl{db: db}
}

func (r *draftPickRepositoryImpl) Create(pick *models.DraftPick) (*models.DraftPick, error) {
	if err := r.db.Create(pick).Error; err != nil {
		return nil, fmt.Errorf("(Error: DraftPickRepo.Create) - failed to create draft pick: %w", err)
	}
	return pick, nil
}

func (r *draftPickRepositoryImpl) CreateBatch(picks []models.DraftPick) error {
	if len(picks) == 0 {
		return nil
	}
	if err := r.db.CreateInBatches(picks, 100).Error; err != nil {
		return fmt.Errorf("(Error: DraftPickRepo.CreateBatch) - failed to batch create draft picks: %w", err)
	}
	return nil
}

func (r *draftPickRepositoryImpl) GetByID(id uuid.UUID) (*models.DraftPick, error) {
	var pick models.DraftPick
	err := r.db.Preload("Draft").
		Preload("Player").
		Preload("Player.User").
		Preload("PoolEntry").
		Preload("PoolEntry.PokemonSpecies").
		First(&pick, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: DraftPickRepo.GetByID) - failed to get draft pick: %w", err)
	}
	return &pick, nil
}

func (r *draftPickRepositoryImpl) GetByDraft(draftID uuid.UUID) ([]models.DraftPick, error) {
	var picks []models.DraftPick
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PoolEntry").
		Preload("PoolEntry.PokemonSpecies").
		Where("draft_id = ?", draftID).
		Order("pick_number ASC").
		Find(&picks).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: DraftPickRepo.GetByDraft) - failed to get draft picks: %w", err)
	}
	return picks, nil
}

func (r *draftPickRepositoryImpl) GetByPlayer(playerID uuid.UUID) ([]models.DraftPick, error) {
	var picks []models.DraftPick
	err := r.db.Preload("Draft").
		Preload("PoolEntry").
		Preload("PoolEntry.PokemonSpecies").
		Where("player_id = ?", playerID).
		Order("pick_number ASC").
		Find(&picks).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: DraftPickRepo.GetByPlayer) - failed to get draft picks: %w", err)
	}
	return picks, nil
}

func (r *draftPickRepositoryImpl) GetCountByDraft(draftID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.DraftPick{}).
		Where("draft_id = ?", draftID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: DraftPickRepo.GetCountByDraft) - failed to count draft picks: %w", err)
	}
	return count, nil
}
