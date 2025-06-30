package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
)

// provides access to the draft data.
type DraftRepository interface {
	GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error)
	CreateDraft(draft *models.Draft) error
	UpdateDraft(draft *models.Draft) error
}

type draftRepository struct {
	db *gorm.DB
}

// creates a new DraftRepository.
func NewDraftRepository(db *gorm.DB) DraftRepository {
	return &draftRepository{db: db}
}

// retrieves the draft associated with a league ID.
func (r *draftRepository) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	draft := &models.Draft{}
	if err := r.db.Where("league_id = ?", leagueID).First(draft).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Or a custom "not found" error
		}
		return nil, err
	}
	return draft, nil
}

// creates a new draft record.
func (r *draftRepository) CreateDraft(draft *models.Draft) error {
	return r.db.Create(draft).Error
}

// UpdateDraft updates an existing draft record.
func (r *draftRepository) UpdateDraft(draft *models.Draft) error {
	return r.db.Save(draft).Error
}
