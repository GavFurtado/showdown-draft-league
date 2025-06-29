package controllers

import (
	"log"
	"net/http"

	"github.com.GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	user, err := ctrl.userService.GetMyProfile(currentUser.ID)
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

	discordDetails, err := ctrl.userService.GetMyDiscordDetails(currentUser.ID)
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
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil { // unreachable code (should be)
		log.Printf("(Error: UpdateProfile) - bad user ID format\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	var req common.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser, err := ctrl.userService.UpdateProfile(userID, req)
	if err != nil {
		log.Printf("(Error: UpdateProfile) - Service failed: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

