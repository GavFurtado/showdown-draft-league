package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockDraftRepository struct {
	mock.Mock
}

func (m *MockDraftRepository) CreateDraft(draft *models.Draft) (*models.Draft, error) {
	args := m.Called(draft)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftRepository) GetDraftByID(id uuid.UUID) (*models.Draft, error) {
	args := m.Called(id)
	var result *models.Draft
	if args.Get(0) != nil {
		result = args.Get(0).(*models.Draft)
	}
	return result, args.Error(1)
}

func (m *MockDraftRepository) GetDraftByLeagueID(leagueID uuid.UUID) (*models.Draft, error) {
	args := m.Called(leagueID)
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

func (m *MockDraftRepository) DeleteDraft(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
