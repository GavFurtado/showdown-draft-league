package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DraftPickRepository defines the interface for managing draft pick event log entries.
type DraftPickRepository interface {
	// CreateDraftPick creates a single draft pick record.
	CreateDraftPick(pick *models.DraftPick) (*models.DraftPick, error)
	// BatchCreateDraftPicks creates multiple draft pick records in a single transaction.
	BatchCreateDraftPicks(picks []models.DraftPick) error
	// GetDraftPickByID retrieves a draft pick by its ID with relationships.
	GetDraftPickByID(id uuid.UUID) (*models.DraftPick, error)
	// GetDraftPicksByDraft retrieves all picks for a given draft, ordered by pick number.
	GetDraftPicksByDraft(draftID uuid.UUID) ([]models.DraftPick, error)
	// GetDraftPicksByPlayer retrieves all picks made by a specific player in a draft.
	GetDraftPicksByPlayer(playerID uuid.UUID) ([]models.DraftPick, error)
	// GetDraftPickCountByDraft returns the total number of picks made in a draft.
	GetDraftPickCountByDraft(draftID uuid.UUID) (int64, error)
}

type draftPickRepositoryImpl struct {
	db *gorm.DB
}

func NewDraftPickRepository(db *gorm.DB) DraftPickRepository {
	return &draftPickRepositoryImpl{db: db}
}

func (r *draftPickRepositoryImpl) CreateDraftPick(pick *models.DraftPick) (*models.DraftPick, error) {
	if err := r.db.Create(pick).Error; err != nil {
		return nil, fmt.Errorf("(Error: CreateDraftPick) - failed to create draft pick: %w", err)
	}
	return pick, nil
}

func (r *draftPickRepositoryImpl) BatchCreateDraftPicks(picks []models.DraftPick) error {
	if len(picks) == 0 {
		return nil
	}
	if err := r.db.CreateInBatches(picks, 100).Error; err != nil {
		return fmt.Errorf("(Error: BatchCreateDraftPicks) - failed to batch create draft picks: %w", err)
	}
	return nil
}

func (r *draftPickRepositoryImpl) GetDraftPickByID(id uuid.UUID) (*models.DraftPick, error) {
	var pick models.DraftPick
	err := r.db.Preload("Draft").
		Preload("Player").
		Preload("Player.User").
		Preload("PoolEntry").
		Preload("PoolEntry.PokemonSpecies").
		First(&pick, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftPickByID) - failed to get draft pick: %w", err)
	}
	return &pick, nil
}

func (r *draftPickRepositoryImpl) GetDraftPicksByDraft(draftID uuid.UUID) ([]models.DraftPick, error) {
	var picks []models.DraftPick
	err := r.db.Preload("Player").
		Preload("Player.User").
		Preload("PoolEntry").
		Preload("PoolEntry.PokemonSpecies").
		Where("draft_id = ?", draftID).
		Order("pick_number ASC").
		Find(&picks).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftPicksByDraft) - failed to get draft picks: %w", err)
	}
	return picks, nil
}

func (r *draftPickRepositoryImpl) GetDraftPicksByPlayer(playerID uuid.UUID) ([]models.DraftPick, error) {
	var picks []models.DraftPick
	err := r.db.Preload("Draft").
		Preload("PoolEntry").
		Preload("PoolEntry.PokemonSpecies").
		Where("player_id = ?", playerID).
		Order("pick_number ASC").
		Find(&picks).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: GetDraftPicksByPlayer) - failed to get draft picks by player: %w", err)
	}
	return picks, nil
}

func (r *draftPickRepositoryImpl) GetDraftPickCountByDraft(draftID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.DraftPick{}).
		Where("draft_id = ?", draftID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: GetDraftPickCountByDraft) - failed to count draft picks: %w", err)
	}
	return count, nil
}
