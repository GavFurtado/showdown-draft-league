package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockDraftPickRepository struct {
	mock.Mock
}

func (m *MockDraftPickRepository) Create(pick *models.DraftPick) (*models.DraftPick, error) {
	args := m.Called(pick)
	var result *models.DraftPick
	if args.Get(0) != nil {
		result = args.Get(0).(*models.DraftPick)
	}
	return result, args.Error(1)
}

func (m *MockDraftPickRepository) CreateBatch(picks []models.DraftPick) error {
	args := m.Called(picks)
	return args.Error(0)
}

func (m *MockDraftPickRepository) GetByID(id uuid.UUID) (*models.DraftPick, error) {
	args := m.Called(id)
	var result *models.DraftPick
	if args.Get(0) != nil {
		result = args.Get(0).(*models.DraftPick)
	}
	return result, args.Error(1)
}

func (m *MockDraftPickRepository) GetByDraft(draftID uuid.UUID) ([]models.DraftPick, error) {
	args := m.Called(draftID)
	return args.Get(0).([]models.DraftPick), args.Error(1)
}

func (m *MockDraftPickRepository) GetByPlayer(playerID uuid.UUID) ([]models.DraftPick, error) {
	args := m.Called(playerID)
	return args.Get(0).([]models.DraftPick), args.Error(1)
}

func (m *MockDraftPickRepository) GetCountByDraft(draftID uuid.UUID) (int64, error) {
	args := m.Called(draftID)
	return args.Get(0).(int64), args.Error(1)
}
