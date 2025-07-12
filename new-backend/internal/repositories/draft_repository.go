package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
)

// TODO: prolly needs more safe db transactions

// provides access to the draft data.
type DraftRepository interface {
	// retrieves the draft associated with a league ID.
	GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error)
	// creates a new draft record.
	CreateDraft(draft *models.Draft) error
	// updates an existing draft record.
	UpdateDraft(draft *models.Draft) error
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
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return draft, nil
}

// updates an existing draft record.
func (r *draftRepositoryImpl) UpdateDraft(draft *models.Draft) error {
	return r.db.Save(draft).Error
}
