package controllers

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
)

// handles player-related HTTP requests.
type PlayerController struct {
	playerService services.PlayerService
}

func NewPlayerController(playerService services.PlayerService) *PlayerController {
	return &PlayerController{
		playerService: playerService,
	}
}
