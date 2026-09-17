package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlayerRepository interface {
	Create(player *models.Player) (*models.Player, error)
	GetByID(id uuid.UUID) (*models.Player, error)
	GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error)
	GetByLeague(leagueID uuid.UUID) ([]models.Player, error)
	GetByLeagueAndGroup(leagueID uuid.UUID, groupNumber int) ([]models.Player, error)
	GetByUser(userID uuid.UUID) ([]models.Player, error)
	Update(player *models.Player) (*models.Player, error)
	UpdateDraftPoints(playerID uuid.UUID, points int) error
	UpdateRecord(playerID uuid.UUID, wins, losses int) error
	UpdateDraftPosition(playerID uuid.UUID, position int) error
	UpdateRole(playerID uuid.UUID, role rbac.MemberRole) error
	GetCountByLeague(leagueID uuid.UUID) (int64, error)
	Delete(playerID uuid.UUID) error
	IsUserInLeague(userID, leagueID uuid.UUID) (bool, error)
	GetWithFullRoster(playerID uuid.UUID) (*models.Player, error)
	FindByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error)
	FindByInLeagueName(name string, leagueID uuid.UUID) (*models.Player, error)
	FindByTeamName(name string, leagueID uuid.UUID) (*models.Player, error)
}

type playerRepositoryImpl struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepositoryImpl{db: db}
}

func (r *playerRepositoryImpl) Create(player *models.Player) (*models.Player, error) {
	err := r.db.Create(player).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}
	return player, nil
}

func (r *playerRepositoryImpl) GetByID(id uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.First(&player, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	return &player, nil
}

func (r *playerRepositoryImpl) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&player).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get player by user and league: %w", err)
	}
	return &player, nil
}

func (r *playerRepositoryImpl) GetByLeague(leagueID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := r.db.
		Where("league_id = ?", leagueID).
		Order("draft_position ASC").
		Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get players for league: %w", err)
	}
	return players, nil
}

func (r *playerRepositoryImpl) GetByLeagueAndGroup(leagueID uuid.UUID, groupNumber int) ([]models.Player, error) {
	var players []models.Player
	err := r.db.
		Where("league_id = ? AND group_number = ?", leagueID, groupNumber).
		Order("draft_position ASC").
		Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get players for league group: %w", err)
	}
	return players, nil
}

func (r *playerRepositoryImpl) GetByUser(userID uuid.UUID) ([]models.Player, error) {
	var players []models.Player
	err := r.db.
		Where("user_id = ?", userID).
		Find(&players).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get players for user: %w", err)
	}
	return players, nil
}

func (r *playerRepositoryImpl) Update(player *models.Player) (*models.Player, error) {
	err := r.db.Model(player).Select("*").Updates(player).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update player: %w", err)
	}
	return r.GetByID(player.ID)
}

func (r *playerRepositoryImpl) UpdateDraftPoints(playerID uuid.UUID, points int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("draft_points", points).Error
	if err != nil {
		return fmt.Errorf("failed to update player draft points: %w", err)
	}
	return nil
}

func (r *playerRepositoryImpl) UpdateRecord(playerID uuid.UUID, wins, losses int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Updates(map[string]any{
			"wins":   wins,
			"losses": losses,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update player record: %w", err)
	}
	return nil
}

func (r *playerRepositoryImpl) UpdateDraftPosition(playerID uuid.UUID, position int) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("draft_position", position).Error
	if err != nil {
		return fmt.Errorf("failed to update player draft position: %w", err)
	}
	return nil
}

func (r *playerRepositoryImpl) UpdateRole(playerID uuid.UUID, role rbac.MemberRole) error {
	err := r.db.Model(&models.Player{}).
		Where("id = ?", playerID).
		Update("role", role).Error
	if err != nil {
		return fmt.Errorf("failed to update player role: %w", err)
	}
	return nil
}

func (r *playerRepositoryImpl) GetCountByLeague(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Player{}).
		Where("league_id = ?", leagueID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count players for league: %w", err)
	}
	return count, nil
}

func (r *playerRepositoryImpl) Delete(playerID uuid.UUID) error {
	err := r.db.Delete(&models.Player{}, "id = ?", playerID).Error
	if err != nil {
		return fmt.Errorf("failed to delete player: %w", err)
	}
	return nil
}

func (r *playerRepositoryImpl) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Player{}).
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check player membership in league: %w", err)
	}
	return count > 0, nil
}

func (r *playerRepositoryImpl) GetWithFullRoster(playerID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.Preload("User").
		Preload("League").
		Preload("Claims").
		First(&player, "id = ?", playerID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get player with full roster: %w", err)
	}
	return &player, nil
}

func (r *playerRepositoryImpl) FindByUserAndLeague(userID, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&player).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find player by user and league: %w", err)
	}
	return &player, nil
}

func (r *playerRepositoryImpl) FindByInLeagueName(name string, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("in_league_name = ? AND league_id = ?", name, leagueID).
		First(&player).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find player by in-league name: %w", err)
	}
	return &player, nil
}

func (r *playerRepositoryImpl) FindByTeamName(name string, leagueID uuid.UUID) (*models.Player, error) {
	var player models.Player
	err := r.db.
		Where("team_name = ? AND league_id = ?", name, leagueID).
		First(&player).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find player by team name: %w", err)
	}
	return &player, nil
}
