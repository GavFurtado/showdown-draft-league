package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SpectatorRepository interface {
	Create(spectator *models.Spectator) (*models.Spectator, error)
	GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Spectator, error)
	GetByLeague(leagueID uuid.UUID) ([]models.Spectator, error)
	GetByUser(userID uuid.UUID) ([]models.Spectator, error)
	Delete(spectatorID uuid.UUID) error
}

type spectatorRepositoryImpl struct {
	db *gorm.DB
}

func NewSpectatorRepository(db *gorm.DB) SpectatorRepository {
	return &spectatorRepositoryImpl{db: db}
}

func (r *spectatorRepositoryImpl) Create(spectator *models.Spectator) (*models.Spectator, error) {
	err := r.db.Create(spectator).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: SpectatorRepo.Create) - failed to create spectator: %w", err)
	}
	return spectator, nil
}

func (r *spectatorRepositoryImpl) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Spectator, error) {
	var spectator models.Spectator
	err := r.db.Where("user_id = ? AND league_id = ?", userID, leagueID).First(&spectator).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: SpectatorRepo.GetByUserAndLeague) - failed: %w", err)
	}
	return &spectator, nil
}

func (r *spectatorRepositoryImpl) GetByLeague(leagueID uuid.UUID) ([]models.Spectator, error) {
	var spectators []models.Spectator
	err := r.db.Where("league_id = ?", leagueID).Find(&spectators).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: SpectatorRepo.GetByLeague) - failed: %w", err)
	}
	return spectators, nil
}

func (r *spectatorRepositoryImpl) GetByUser(userID uuid.UUID) ([]models.Spectator, error) {
	var spectators []models.Spectator
	err := r.db.Where("user_id = ?", userID).Find(&spectators).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: SpectatorRepo.GetByUser) - failed: %w", err)
	}
	return spectators, nil
}

func (r *spectatorRepositoryImpl) Delete(spectatorID uuid.UUID) error {
	err := r.db.Delete(&models.Spectator{}, "id = ?", spectatorID).Error
	if err != nil {
		return fmt.Errorf("(Error: SpectatorRepo.Delete) - failed to delete spectator: %w", err)
	}
	return nil
}