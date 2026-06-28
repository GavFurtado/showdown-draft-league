package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PoolEntryRepository interface {
	GetByID(id uuid.UUID) (*models.PoolEntry, error)
	GetByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error)
	GetAvailableByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error)
	GetByIDs(leagueID uuid.UUID, ids []uuid.UUID) ([]models.PoolEntry, error)
	GetBySpecies(leagueID uuid.UUID, speciesID int64) (*models.PoolEntry, error)
	GetByCostRange(leagueID uuid.UUID, minCost, maxCost int) ([]models.PoolEntry, error)
	IsAvailable(leagueID uuid.UUID, speciesID int64) (bool, error)
	GetCost(leagueID uuid.UUID, speciesID int64) (*int, error)
	GetAvailableCount(leagueID uuid.UUID) (int64, error)
	Create(entry *models.PoolEntry) (*models.PoolEntry, error)
	CreateBatch(entries []models.PoolEntry) ([]models.PoolEntry, error)
	Update(entry *models.PoolEntry) (*models.PoolEntry, error)
	MarkUnavailable(tx *gorm.DB, id uuid.UUID) error
	MarkAvailable(tx *gorm.DB, id uuid.UUID) error
	Delete(leagueID uuid.UUID, speciesID int64) error
	DeleteAllByLeague(leagueID uuid.UUID) error
}

type poolEntryRepositoryImpl struct {
	db *gorm.DB
}

func NewPoolEntryRepository(db *gorm.DB) PoolEntryRepository {
	return &poolEntryRepositoryImpl{db: db}
}

func (r *poolEntryRepositoryImpl) GetByID(id uuid.UUID) (*models.PoolEntry, error) {
	var entry models.PoolEntry
	err := r.db.Preload("PokemonSpecies").
		First(&entry, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetByID) - failed to get pool entry: %w", err)
	}
	return &entry, nil
}

func (r *poolEntryRepositoryImpl) GetByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error) {
	var entries []models.PoolEntry
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ?", leagueID).
		Order("cost ASC, pokemon_species_id ASC").
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetByLeague) - failed to get pool entries: %w", err)
	}
	return entries, nil
}

func (r *poolEntryRepositoryImpl) GetAvailableByLeague(leagueID uuid.UUID) ([]models.PoolEntry, error) {
	var entries []models.PoolEntry
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND is_available = ?", leagueID, true).
		Order("cost DESC, created_at ASC").
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetAvailableByLeague) - failed to get available pool entries: %w", err)
	}
	return entries, nil
}

func (r *poolEntryRepositoryImpl) GetByIDs(leagueID uuid.UUID, ids []uuid.UUID) ([]models.PoolEntry, error) {
	var entries []models.PoolEntry
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND id IN (?)", leagueID, ids).
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetByIDs) - failed to get pool entries by IDs: %w", err)
	}
	return entries, nil
}

func (r *poolEntryRepositoryImpl) GetBySpecies(leagueID uuid.UUID, speciesID int64) (*models.PoolEntry, error) {
	var entry models.PoolEntry
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND pokemon_species_id = ?", leagueID, speciesID).
		First(&entry).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetBySpecies) - failed to get pool entry by species: %w", err)
	}
	return &entry, nil
}

func (r *poolEntryRepositoryImpl) GetByCostRange(leagueID uuid.UUID, minCost, maxCost int) ([]models.PoolEntry, error) {
	var entries []models.PoolEntry
	err := r.db.Preload("PokemonSpecies").
		Where("league_id = ? AND cost BETWEEN ? AND ?", leagueID, minCost, maxCost).
		Order("cost ASC").
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetByCostRange) - failed to get pool entries by cost: %w", err)
	}
	return entries, nil
}

func (r *poolEntryRepositoryImpl) IsAvailable(leagueID uuid.UUID, speciesID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.PoolEntry{}).
		Where("league_id = ? AND pokemon_species_id = ? AND is_available = ?", leagueID, speciesID, true).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("(Error: PoolEntryRepo.IsAvailable) - failed to check availability: %w", err)
	}
	return count > 0, nil
}

func (r *poolEntryRepositoryImpl) GetCost(leagueID uuid.UUID, speciesID int64) (*int, error) {
	var entry models.PoolEntry
	err := r.db.Select("cost").
		Where("league_id = ? AND pokemon_species_id = ?", leagueID, speciesID).
		First(&entry).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.GetCost) - failed to get cost: %w", err)
	}
	return entry.Cost, nil
}

func (r *poolEntryRepositoryImpl) GetAvailableCount(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.PoolEntry{}).
		Where("league_id = ? AND is_available = ?", leagueID, true).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: PoolEntryRepo.GetAvailableCount) - failed to count: %w", err)
	}
	return count, nil
}

func (r *poolEntryRepositoryImpl) Create(entry *models.PoolEntry) (*models.PoolEntry, error) {
	err := r.db.Create(entry).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.Create) - failed to create pool entry: %w", err)
	}
	err = r.db.Preload("PokemonSpecies").First(entry, "id = ?", entry.ID).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.Create) - failed to preload species: %w", err)
	}
	return entry, nil
}

func (r *poolEntryRepositoryImpl) CreateBatch(entries []models.PoolEntry) ([]models.PoolEntry, error) {
	if len(entries) == 0 {
		return entries, nil
	}

	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.CreateBatch) - failed to start tx: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.CreateInBatches(entries, 100).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("(Error: PoolEntryRepo.CreateBatch) - failed to create batch: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.CreateBatch) - failed to commit: %w", err)
	}

	var created []models.PoolEntry
	if len(entries) > 0 {
		err := r.db.Preload("PokemonSpecies").
			Where("league_id = ?", entries[0].LeagueID).
			Where("id IN (?)", func() []uuid.UUID {
				ids := make([]uuid.UUID, len(entries))
				for i, e := range entries {
					ids[i] = e.ID
				}
				return ids
			}()).
			Find(&created).Error
		if err != nil {
			return nil, fmt.Errorf("(Error: PoolEntryRepo.CreateBatch) - failed to retrieve created entries: %w", err)
		}
	}

	return created, nil
}

func (r *poolEntryRepositoryImpl) Update(entry *models.PoolEntry) (*models.PoolEntry, error) {
	err := r.db.Select("cost", "is_available", "updated_at").
		Updates(entry).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: PoolEntryRepo.Update) - failed to update pool entry: %w", err)
	}
	return entry, nil
}

func (r *poolEntryRepositoryImpl) MarkUnavailable(tx *gorm.DB, id uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.Model(&models.PoolEntry{}).
		Where("id = ?", id).
		Update("is_available", false).Error
	if err != nil {
		return fmt.Errorf("(Error: PoolEntryRepo.MarkUnavailable) - failed to mark unavailable: %w", err)
	}
	return nil
}

func (r *poolEntryRepositoryImpl) MarkAvailable(tx *gorm.DB, id uuid.UUID) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.Model(&models.PoolEntry{}).
		Where("id = ?", id).
		Update("is_available", true).Error
	if err != nil {
		return fmt.Errorf("(Error: PoolEntryRepo.MarkAvailable) - failed to mark available: %w", err)
	}
	return nil
}

func (r *poolEntryRepositoryImpl) Delete(leagueID uuid.UUID, speciesID int64) error {
	err := r.db.Where("league_id = ? AND pokemon_species_id = ?", leagueID, speciesID).
		Delete(&models.PoolEntry{}).Error
	if err != nil {
		return fmt.Errorf("(Error: PoolEntryRepo.Delete) - failed to delete: %w", err)
	}
	return nil
}

func (r *poolEntryRepositoryImpl) DeleteAllByLeague(leagueID uuid.UUID) error {
	err := r.db.Where("league_id = ?", leagueID).
		Delete(&models.PoolEntry{}).Error
	if err != nil {
		return fmt.Errorf("(Error: PoolEntryRepo.DeleteAllByLeague) - failed to delete all: %w", err)
	}
	return nil
}
