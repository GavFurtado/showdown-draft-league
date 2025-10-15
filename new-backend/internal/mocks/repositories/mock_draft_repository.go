package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models/enums"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockDraftRepository struct {
	mock.Mock
}

func (m *MockDraftRepository) CreateDraft(draft *models.Draft) error {
	args := m.Called(draft)
	return args.Error(0)
}

func (m *MockDraftRepository) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	args := m.Called(leagueID)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftRepository) GetDraftByID(draftID uuid.UUID) (*models.Draft, error) {
	args := m.Called(draftID)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftRepository) UpdateDraft(draft *models.Draft) (*models.Draft, error) {
	args := m.Called(draft)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftRepository) GetAllDraftsByStatus(status enums.DraftStatus) ([]models.Draft, error) {
	args := m.Called(status)
	var result []models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).([]models.Draft)
	}
	return result, args.Error(1)
}
