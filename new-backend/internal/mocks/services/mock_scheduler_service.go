package mock_services

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	u "github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/stretchr/testify/mock"
)

type MockSchedulerService struct {
	mock.Mock
}

func (m *MockSchedulerService) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSchedulerService) RegisterTask(task *u.ScheduledTask) {
	m.Called(task)
}

func (m *MockSchedulerService) DeregisterTask(taskID string) {
	m.Called(taskID)
}

func (m *MockSchedulerService) Stop() {
	m.Called()
}

func (m *MockSchedulerService) SetDraftService(draftService services.DraftService) {
	m.Called(draftService)
}
