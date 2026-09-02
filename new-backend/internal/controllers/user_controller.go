package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// TODO: why is this not an interface like the other controllers

type UserController struct {
	logger      *slog.Logger
	userService services.UserService
}

func NewUserController(logger *slog.Logger, userService services.UserService) UserController {
	return UserController{
		logger:      utils.LoggerWithService(logger, "UserController"),
		userService: userService,
	}
}

// GetMyProfile godoc
//
//	@Summary		Get my profile
//	@Description	Your user profile
//	@Tags			Users
//	@Success		200	{object}	models.User
//	@Failure		401	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/profile [get]
func (ctrl *UserController) GetMyProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctrl.logger.Error("no user in context", "method", "GetMyProfile")
		sendError(ctx, http.StatusInternalServerError, "User information not available")
		return
	}

	user, err := ctrl.userService.GetMyProfileHandler(currentUser.ID)
	if err != nil {
		ctrl.logger.Error("Service failed", "error", err, "method", "GetMyProfile")
		if err.Error() == "user not found" {
			sendError(ctx, http.StatusNotFound, "User not found")
			return
		}
		sendError(ctx, http.StatusInternalServerError, "Failed to retrieve user profile")
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
//	@Failure		401	{object}	responses.ErrorResponse
//	@Failure		404	{object}	responses.ErrorResponse
//	@Router			/api/users/me/discord [get]
func (ctrl *UserController) GetMyDiscordDetails(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctrl.logger.Error("no user in context", "method", "GetMyDiscordDetails")
		sendError(ctx, http.StatusInternalServerError, "User information not available")
		return
	}

	discordDetails, err := ctrl.userService.GetMyDiscordDetailsHandler(currentUser.ID)
	if err != nil {
		ctrl.logger.Error("Service failed", "error", err, "method", "GetMyDiscordDetails")
		if err.Error() == "user not found" {
			sendError(ctx, http.StatusNotFound, "User not found")
			return
		}
		sendError(ctx, http.StatusInternalServerError, "Failed to retrieve Discord details")
		return
	}

	ctx.JSON(http.StatusOK, discordDetails)
}

// UpdateProfile godoc
//
//	@Summary		Update my profile
//	@Description	Update the user's profile info
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.UserUpdateProfileRequestDTO	true	"Profile fields to update"
//	@Success		200		{object}	models.User
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		401		{object}	responses.ErrorResponse
//	@Router			/api/users/profile [put]
func (ctrl *UserController) UpdateProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctrl.logger.Error("no user in context", "method", "UpdateProfile")
		sendError(ctx, http.StatusInternalServerError, "User information not available")
		return
	}

	var req requests.UserUpdateProfileRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing field(s) in the payload")
		return
	}

	updatedUser, err := ctrl.userService.UpdateProfileHandler(currentUser.ID, req)
	if err != nil {
		ctrl.logger.Error("Service failed", "error", err, "method", "UpdateProfile")
		sendError(ctx, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}

// GetMyLeagues godoc
//
//	@Summary		Get my leagues
//	@Description	Leagues the user a member of
//	@Tags			Users
//	@Success		200	{array}		models.League
//	@Failure		401	{object}	responses.ErrorResponse
//	@Failure		500	{object}	responses.ErrorResponse
//	@Router			/api/users/me/leagues [get]
func (ctrl *UserController) GetMyLeagues(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctrl.logger.Error("no user in context", "method", "GetMyLeagues")
		sendError(ctx, http.StatusInternalServerError, "User information not available")
		return
	}

	leagues, err := ctrl.userService.GetMyLeaguesHandler(currentUser.ID)
	if err != nil {
		ctrl.logger.Error("Service failed", "error", err, "method", "GetMyLeagues")
		sendError(ctx, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	ctx.JSON(http.StatusOK, leagues)
}
