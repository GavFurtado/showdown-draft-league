package mock_services

import (
	"github.com/stretchr/testify/mock"
)

type MockWebhookService struct {
	mock.Mock
}

func (m *MockWebhookService) SendWebhookMessage(webhookURL, message string) error {
	args := m.Called(webhookURL, message)
	return args.Error(0)
}
