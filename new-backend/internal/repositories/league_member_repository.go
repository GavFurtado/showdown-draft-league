package repositories

import (
	"fmt"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeagueMemberRepository interface {
	Create(member *models.LeagueMember) (*models.LeagueMember, error)
	GetByID(id uuid.UUID) (*models.LeagueMember, error)
	GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error)
	GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error)
	GetByLeagueAndGroup(leagueID uuid.UUID, groupNumber int) ([]models.LeagueMember, error)
	GetByUser(userID uuid.UUID) ([]models.LeagueMember, error)
	Update(member *models.LeagueMember) (*models.LeagueMember, error)
	UpdateDraftPoints(memberID uuid.UUID, points int) error
	UpdateRecord(memberID uuid.UUID, wins, losses int) error
	UpdateDraftPosition(memberID uuid.UUID, position int) error
	UpdateRole(memberID uuid.UUID, role rbac.MemberRole) error
	GetCountByLeague(leagueID uuid.UUID) (int64, error)
	Delete(memberID uuid.UUID) error
	IsUserInLeague(userID, leagueID uuid.UUID) (bool, error)
	GetWithFullRoster(memberID uuid.UUID) (*models.LeagueMember, error)
	FindByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error)
	FindByInLeagueName(name string, leagueID uuid.UUID) (*models.LeagueMember, error)
	FindByTeamName(name string, leagueID uuid.UUID) (*models.LeagueMember, error)
}

type leagueMemberRepositoryImpl struct {
	db *gorm.DB
}

func NewLeagueMemberRepository(db *gorm.DB) LeagueMemberRepository {
	return &leagueMemberRepositoryImpl{db: db}
}

func (r *leagueMemberRepositoryImpl) Create(member *models.LeagueMember) (*models.LeagueMember, error) {
	err := r.db.Create(member).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.Create) - failed to create member: %w", err)
	}
	return member, nil
}

func (r *leagueMemberRepositoryImpl) GetByID(id uuid.UUID) (*models.LeagueMember, error) {
	var member models.LeagueMember
	err := r.db.First(&member, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.GetByID) - failed to get member: %w", err)
	}
	return &member, nil
}

func (r *leagueMemberRepositoryImpl) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error) {
	var member models.LeagueMember
	err := r.db.
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&member).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.GetByUserAndLeague) - failed: %w", err)
	}
	return &member, nil
}

func (r *leagueMemberRepositoryImpl) GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error) {
	var members []models.LeagueMember
	err := r.db.
		Where("league_id = ?", leagueID).
		Order("draft_position ASC").
		Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.GetByLeague) - failed: %w", err)
	}
	return members, nil
}

func (r *leagueMemberRepositoryImpl) GetByLeagueAndGroup(leagueID uuid.UUID, groupNumber int) ([]models.LeagueMember, error) {
	var members []models.LeagueMember
	err := r.db.
		Where("league_id = ? AND group_number = ?", leagueID, groupNumber).
		Order("draft_position ASC").
		Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.GetByLeagueAndGroup) - failed: %w", err)
	}
	return members, nil
}

func (r *leagueMemberRepositoryImpl) GetByUser(userID uuid.UUID) ([]models.LeagueMember, error) {
	var members []models.LeagueMember
	err := r.db.
		Where("user_id = ?", userID).
		Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.GetByUser) - failed: %w", err)
	}
	return members, nil
}

func (r *leagueMemberRepositoryImpl) Update(member *models.LeagueMember) (*models.LeagueMember, error) {
	err := r.db.Model(member).Select("*").Updates(member).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.Update) - failed: %w", err)
	}
	return r.GetByID(member.ID)
}

func (r *leagueMemberRepositoryImpl) UpdateDraftPoints(memberID uuid.UUID, points int) error {
	err := r.db.Model(&models.LeagueMember{}).
		Where("id = ?", memberID).
		Update("draft_points", points).Error
	if err != nil {
		return fmt.Errorf("(Error: LeagueMemberRepo.UpdateDraftPoints) - failed: %w", err)
	}
	return nil
}

func (r *leagueMemberRepositoryImpl) UpdateRecord(memberID uuid.UUID, wins, losses int) error {
	err := r.db.Model(&models.LeagueMember{}).
		Where("id = ?", memberID).
		Updates(map[string]any{
			"wins":   wins,
			"losses": losses,
		}).Error
	if err != nil {
		return fmt.Errorf("(Error: LeagueMemberRepo.UpdateRecord) - failed: %w", err)
	}
	return nil
}

func (r *leagueMemberRepositoryImpl) UpdateDraftPosition(memberID uuid.UUID, position int) error {
	err := r.db.Model(&models.LeagueMember{}).
		Where("id = ?", memberID).
		Update("draft_position", position).Error
	if err != nil {
		return fmt.Errorf("(Error: LeagueMemberRepo.UpdateDraftPosition) - failed: %w", err)
	}
	return nil
}

func (r *leagueMemberRepositoryImpl) UpdateRole(memberID uuid.UUID, role rbac.MemberRole) error {
	err := r.db.Model(&models.LeagueMember{}).
		Where("id = ?", memberID).
		Update("role", role).Error
	if err != nil {
		return fmt.Errorf("(Error: LeagueMemberRepo.UpdateRole) - failed: %w", err)
	}
	return nil
}

func (r *leagueMemberRepositoryImpl) GetCountByLeague(leagueID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.LeagueMember{}).
		Where("league_id = ?", leagueID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("(Error: LeagueMemberRepo.GetCountByLeague) - failed: %w", err)
	}
	return count, nil
}

func (r *leagueMemberRepositoryImpl) Delete(memberID uuid.UUID) error {
	err := r.db.Delete(&models.LeagueMember{}, "id = ?", memberID).Error
	if err != nil {
		return fmt.Errorf("(Error: LeagueMemberRepo.Delete) - failed: %w", err)
	}
	return nil
}

func (r *leagueMemberRepositoryImpl) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.LeagueMember{}).
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("(Error: LeagueMemberRepo.IsUserInLeague) - failed: %w", err)
	}
	return count > 0, nil
}

func (r *leagueMemberRepositoryImpl) GetWithFullRoster(memberID uuid.UUID) (*models.LeagueMember, error) {
	var member models.LeagueMember
	err := r.db.Preload("User").
		Preload("League").
		First(&member, "id = ?", memberID).Error
	if err != nil {
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.GetWithFullRoster) - failed: %w", err)
	}
	return &member, nil
}

func (r *leagueMemberRepositoryImpl) FindByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error) {
	var member models.LeagueMember
	err := r.db.
		Where("user_id = ? AND league_id = ?", userID, leagueID).
		First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.FindByUserAndLeague) - failed: %w", err)
	}
	return &member, nil
}

func (r *leagueMemberRepositoryImpl) FindByInLeagueName(name string, leagueID uuid.UUID) (*models.LeagueMember, error) {
	var member models.LeagueMember
	err := r.db.
		Where("in_league_name = ? AND league_id = ?", name, leagueID).
		First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.FindByInLeagueName) - failed: %w", err)
	}
	return &member, nil
}

func (r *leagueMemberRepositoryImpl) FindByTeamName(name string, leagueID uuid.UUID) (*models.LeagueMember, error) {
	var member models.LeagueMember
	err := r.db.
		Where("team_name = ? AND league_id = ?", name, leagueID).
		First(&member).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("(Error: LeagueMemberRepo.FindByTeamName) - failed: %w", err)
	}
	return &member, nil
}
