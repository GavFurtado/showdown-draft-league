package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userRepo   *repositories.UserRepository
	playerRepo *repositories.PlayerRepository
	leagueRepo *repositories.LeagueRepository
}

func NewUserController(userRepo *repositories.UserRepository,
	playerRepo *repositories.PlayerRepository,
	leagueRepo *repositories.LeagueRepository) UserController {

	return UserController{
		userRepo:   userRepo,
		playerRepo: playerRepo,
		leagueRepo: leagueRepo,
	}
}

// gets current user profile (this seems extra cuz it does the same thing as GetUserFromContext but)
func (ctrl *UserController) GetMyProfile(ctx *gin.Context) (*models.User, error) {
	user, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyProfile) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return nil, errors.New("no user in context")
	}

	return user, nil
}

// gets current user's discord details
// main use case is for the profile on navbar
func (ctrl *UserController) GetMyDiscordDetails(ctx *gin.Context) (*common.DiscordUser, error) {
	user, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyDiscordDetails) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return nil, errors.New("no user in context")
	}

	discordDeets := common.DiscordUser{
		ID:       user.ID.String(),
		Username: user.DiscordUsername,
		Avatar:   user.DiscordAvatarURL,
	}

	return &discordDeets, nil
}

func (ctrl *UserController) UpdateProfile(ctx *gin.Context)
