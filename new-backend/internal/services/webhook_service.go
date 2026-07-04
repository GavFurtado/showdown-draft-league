package services

import (
	"log/slog"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
)

// handles sending notifications to external webhooks.
type WebhookService interface {
	SendWebhookMessage(webhookURL string, message string) error
}

type webhookService struct {
	logger *slog.Logger
}

func NewWebhookService(logger *slog.Logger) WebhookService {
	return &webhookService{
		logger: utils.LoggerWithService(logger, "WebhookService"),
	}
}

// sends a message to the specified webhook URL.
// Currently, this is a placeholder and only logs the attempt.
// TODO: Implement actual HTTP POST request to the webhookURL.
func (s *webhookService) SendWebhookMessage(webhookURL string, message string) error {
	if webhookURL == "" {
		// No webhook configured, just return without error :(
		return nil
	}
	s.logger.Info("webhook placeholder - attempting to send message", "webhook_url", webhookURL, "message", message)

	return nil
}
