package mock_repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockLeagueMemberRepository struct {
	mock.Mock
}

func (m *MockLeagueMemberRepository) Create(member *models.LeagueMember) (*models.LeagueMember, error) {
	args := m.Called(member)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(1)
}

func (m *MockLeagueMemberRepository) GetByID(id uuid.UUID) (*models.LeagueMember, error) {
	args := m.Called(id)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(1)
}

func (m *MockLeagueMemberRepository) GetByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error) {
	args := m.Called(userID, leagueID)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(2)
}

func (m *MockLeagueMemberRepository) GetByLeague(leagueID uuid.UUID) ([]models.LeagueMember, error) {
	args := m.Called(leagueID)
	return args.Get(0).([]models.LeagueMember), args.Error(1)
}

func (m *MockLeagueMemberRepository) GetByLeagueAndGroup(leagueID uuid.UUID, groupNumber int) ([]models.LeagueMember, error) {
	args := m.Called(leagueID, groupNumber)
	return args.Get(0).([]models.LeagueMember), args.Error(1)
}

func (m *MockLeagueMemberRepository) GetByUser(userID uuid.UUID) ([]models.LeagueMember, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.LeagueMember), args.Error(1)
}

func (m *MockLeagueMemberRepository) Update(member *models.LeagueMember) (*models.LeagueMember, error) {
	args := m.Called(member)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(1)
}

func (m *MockLeagueMemberRepository) UpdateDraftPoints(memberID uuid.UUID, points int) error {
	args := m.Called(memberID, points)
	return args.Error(0)
}

func (m *MockLeagueMemberRepository) UpdateRecord(memberID uuid.UUID, wins, losses int) error {
	args := m.Called(memberID, wins, losses)
	return args.Error(0)
}

func (m *MockLeagueMemberRepository) UpdateDraftPosition(memberID uuid.UUID, position int) error {
	args := m.Called(memberID, position)
	return args.Error(0)
}

func (m *MockLeagueMemberRepository) UpdateRole(memberID uuid.UUID, role rbac.MemberRole) error {
	args := m.Called(memberID, role)
	return args.Error(0)
}

func (m *MockLeagueMemberRepository) GetCountByLeague(leagueID uuid.UUID) (int64, error) {
	args := m.Called(leagueID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLeagueMemberRepository) Delete(memberID uuid.UUID) error {
	args := m.Called(memberID)
	return args.Error(0)
}

func (m *MockLeagueMemberRepository) IsUserInLeague(userID, leagueID uuid.UUID) (bool, error) {
	args := m.Called(userID, leagueID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLeagueMemberRepository) GetWithFullRoster(memberID uuid.UUID) (*models.LeagueMember, error) {
	args := m.Called(memberID)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(1)
}

func (m *MockLeagueMemberRepository) FindByUserAndLeague(userID, leagueID uuid.UUID) (*models.LeagueMember, error) {
	args := m.Called(userID, leagueID)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(2)
}

func (m *MockLeagueMemberRepository) FindByInLeagueName(name string, leagueID uuid.UUID) (*models.LeagueMember, error) {
	args := m.Called(name, leagueID)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(2)
}

func (m *MockLeagueMemberRepository) FindByTeamName(name string, leagueID uuid.UUID) (*models.LeagueMember, error) {
	args := m.Called(name, leagueID)
	var result *models.LeagueMember
	if args.Get(0) != nil {
		result = args.Get(0).(*models.LeagueMember)
	}
	return result, args.Error(2)
}
