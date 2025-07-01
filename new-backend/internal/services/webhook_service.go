package services

import (
	"fmt"
)

// handles sending notifications to external webhooks.
type WebhookService interface {
	SendWebhookMessage(webhookURL string, message string) error
}

type webhookService struct {
	// not sure what to put here yet
}

func NewWebhookService() WebhookService {
	return &webhookService{}
}

// sends a message to the specified webhook URL.
// Currently, this is a placeholder and only logs the attempt.
// TODO: Implement actual HTTP POST request to the webhookURL.
func (s *webhookService) SendWebhookMessage(webhookURL string, message string) error {
	if webhookURL == "" {
		// No webhook configured, just return without error :(
		return nil
	}
	fmt.Printf("WEBHOOK PLACEHOLDER: Attempting to send message to %s with content: %s\n", webhookURL, message)

	return nil
}
