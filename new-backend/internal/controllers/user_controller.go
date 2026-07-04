package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) UserController {
	return UserController{
		userService: userService,
	}
}

// GetMyProfile godoc
//
//	@Summary		Get my profile
//	@Description	Your user profile
//	@Tags			Users
//	@Success		200	{object}	models.User
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/profile [get]
func (ctrl *UserController) GetMyProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyProfile) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	user, err := ctrl.userService.GetMyProfileHandler(currentUser.ID)
	if err != nil {
		log.Printf("(Error: GetMyProfile) - Service failed: %v\n", err)
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user profile"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetMyDiscordDetails godoc
//
//	@Summary		Get my Discord details
//	@Description	Discord account info for the navbar
//	@Tags			Users
//	@Success		200	{object}	responses.DiscordUserResponse
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		404	{object}	map[string]interface{}
//	@Router			/api/users/me/discord [get]
func (ctrl *UserController) GetMyDiscordDetails(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyDiscordDetails) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	discordDetails, err := ctrl.userService.GetMyDiscordDetailsHandler(currentUser.ID)
	if err != nil {
		log.Printf("(Error: GetMyDiscordDetails) - Service failed: %v\n", err)
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Discord details"})
		return
	}

	ctx.JSON(http.StatusOK, discordDetails)
}

// UpdateProfile godoc
//
//	@Summary		Update my profile
//	@Description	Update your profile info
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.UserUpdateProfileRequestDTO	true	"Profile fields to update"
//	@Success		200		{object}	models.User
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Router			/api/users/profile [put]
func (ctrl *UserController) UpdateProfile(ctx *gin.Context) {
	// doesn't have admin override (can be done if we just have userID in req instead and modify the service a little bit)
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyProfile) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	var req requests.UserUpdateProfileRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing field(s) in the payload"})
		return
	}

	updatedUser, err := ctrl.userService.UpdateProfileHandler(currentUser.ID, req)
	if err != nil {
		log.Printf("(Error: UpdateProfile) - Service failed: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

// GetMyLeagues godoc
//
//	@Summary		Get my leagues
//	@Description	Leagues you're a member of
//	@Tags			Users
//	@Success		200	{array}		models.League
//	@Failure		401	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/api/users/me/leagues [get]
func (ctrl *UserController) GetMyLeagues(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyLeagues) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	leagues, err := ctrl.userService.GetMyLeaguesHandler(currentUser.ID)
	if err != nil {
		if err.Error() == fmt.Sprintf("user not found: %v", err) { // should be unreachable code
			log.Printf("(Error: GetMyLeagues) - user not found %v\n", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("(Error: GetMyLeagues) - Other Database error occurred %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	ctx.JSON(http.StatusOK, leagues)
}
