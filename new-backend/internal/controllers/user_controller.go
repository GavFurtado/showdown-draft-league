package controllers

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"gorm.io/gorm"
)

type UserController struct {
	userRepo   *repositories.UserRepository
	playerRepo *repositories.PlayerRepository
	leagueRepo *repositories.LeagueRepository
}
