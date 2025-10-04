package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
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

// gets current user profile.
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

// gets current user's discord details
// main use case is for the profile on navbar
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

// updates a user's profile
// currently (29/06/25) only does Showdown Username cuz that's the only thing that should be updatable
func (ctrl *UserController) UpdateProfile(ctx *gin.Context) {
	// doesn't have admin override (can be done if we just have userID in req instead and modify the service a little bit)
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetMyProfile) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	var req common.UserUpdateProfileRequest
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

	ctx.JSON(http.StatusOK, leagues) // lets hope i didn't screw up the json tags
}
