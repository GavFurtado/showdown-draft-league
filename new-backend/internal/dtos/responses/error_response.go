package responses

import (
	"net/http"
	"time"
)

type ErrorResponse struct {
	Timestamp string `json:"Timestamp"`
	Status    int    `json:"Status"`
	Error     string `json:"Error"`
	Message   string `json:"Message"`
	Path      string `json:"Path"`
}

func NewErrorResponse(status int, message string, path string) ErrorResponse {
	return ErrorResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    status,
		Error:     http.StatusText(status),
		Message:   message,
		Path:      path,
	}
}
