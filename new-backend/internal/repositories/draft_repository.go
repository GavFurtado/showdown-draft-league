package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
)

// TODO: prolly needs more safe db transactions

// provides access to the draft data.
type DraftRepository interface {
	// retrieves the draft associated with a league ID.
	GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error)
	// retrieves a draft by its ID.
	GetDraftByID(draftID uuid.UUID) (*models.Draft, error)
	// creates a new draft record.
	CreateDraft(draft *models.Draft) error
	// updates an existing draft record.
	UpdateDraft(draft *models.Draft) (*models.Draft, error)
	// retrieves all drafts with a specific status.
	GetAllDraftsByStatus(status enums.DraftStatus) ([]models.Draft, error)
}

type draftRepositoryImpl struct {
	db *gorm.DB
}

func NewDraftRepository(db *gorm.DB) DraftRepository {
	return &draftRepositoryImpl{db: db}
}

// creates a new draft record.
func (r *draftRepositoryImpl) CreateDraft(draft *models.Draft) error {
	return r.db.Create(draft).Error
}

// retrieves the draft associated with a league ID.
func (r *draftRepositoryImpl) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	draft := &models.Draft{}
	if err := r.db.Where("league_id = ?", leagueID).First(draft).Error; err != nil {
		return nil, err
	}
	return draft, nil
}

// retrieves a draft by its ID.
func (r *draftRepositoryImpl) GetDraftByID(draftID uuid.UUID) (*models.Draft, error) {
	draft := &models.Draft{}
	if err := r.db.Where("id = ?", draftID).First(draft).Error; err != nil {
		return nil, err
	}
	return draft, nil
}

// updates an existing draft record.
func (r *draftRepositoryImpl) UpdateDraft(draft *models.Draft) (*models.Draft, error) {
	err := r.db.Save(draft).Error
	if err != nil {
		return nil, err
	}
	return r.GetDraftByLeagueID(draft.LeagueID)
}

// retrieves all drafts with a specific status.
func (r *draftRepositoryImpl) GetAllDraftsByStatus(status enums.DraftStatus) ([]models.Draft, error) {
	var drafts []models.Draft
	if err := r.db.Where("status = ?", status).Find(&drafts).Error; err != nil {
		return nil, err
	}
	return drafts, nil
}
