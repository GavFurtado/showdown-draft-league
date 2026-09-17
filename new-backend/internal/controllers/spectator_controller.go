package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SpectatorController interface {
	GetByLeague(ctx *gin.Context)
	JoinLeague(ctx *gin.Context)
	LeaveLeague(ctx *gin.Context)
	UpdateOwnProfile(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	UpdateRole(ctx *gin.Context)
	DeleteSpectator(ctx *gin.Context)
}

type spectatorControllerImpl struct {
	logger           *slog.Logger
	spectatorService services.SpectatorService
}

func NewSpectatorController(logger *slog.Logger, spectatorService services.SpectatorService) SpectatorController {
	return &spectatorControllerImpl{
		logger:           utils.LoggerWithService(logger, "SpectatorController"),
		spectatorService: spectatorService,
	}
}

// GetByLeague godoc
//
//	@Summary		Get league spectators
//	@Description	All spectators of a league. Requires read:spectator.
//	@Tags			Spectators
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.Spectator
//	@Failure		400			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators [get]
func (c *spectatorControllerImpl) GetByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	spectators, err := c.spectatorService.GetByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByLeague")
		sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		return
	}

	ctx.JSON(http.StatusOK, spectators)
}

// JoinLeague godoc
//
//	@Summary		Join a league as a spectator
//	@Description	Add yourself to a league as a spectator (a non-participating member).
//	@Tags			Spectators
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			request		body		object	true	"Specify inLeagueName"
//	@Success		200			{object}	models.Spectator
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators/join [post]
func (c *spectatorControllerImpl) JoinLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		InLeagueName *string `json:"InLeagueName" validate:"omitempty,min=3,max=20"`
	}
	if err = ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	spectator, err := c.spectatorService.JoinLeague(currentUser.ID, leagueID, req.InLeagueName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "JoinLeague")
		c.handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, spectator)
}

// LeaveLeague godoc
//
//	@Summary		Leave a league as a spectator
//	@Description	Remove yourself as a spectator from a league.
//	@Tags			Spectators
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators/me [delete]
func (c *spectatorControllerImpl) LeaveLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err = c.spectatorService.LeaveLeague(currentUser.ID, leagueID); err != nil {
		c.logger.Error("Service method error", "error", err, "method", "LeaveLeague")
		c.handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Left the league as a spectator"})
}

// UpdateOwnProfile godoc
//
//	@Summary		Update your own spectator profile
//	@Description	Update your own in-league name. Requires update:own_spectator.
//	@Tags			Spectators
//	@Accept			json
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			request		body		object	true	"Specify inLeagueName"
//	@Success		200			{object}	models.Spectator
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators/me/profile [put]
func (c *spectatorControllerImpl) UpdateOwnProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		InLeagueName *string `json:"InLeagueName" validate:"omitempty,min=3,max=20"`
	}
	if err = ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	spectator, err := c.spectatorService.UpdateOwnProfile(currentUser.ID, leagueID, req.InLeagueName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateOwnProfile")
		c.handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, spectator)
}

// UpdateProfile godoc
//
//	@Summary		Update a spectator's profile (staff)
//	@Description	Update a spectator's in-league name. Requires update:spectator.
//	@Tags			Spectators
//	@Accept			json
//	@Produce		json
//	@Param			leagueId		path		string	true	"League ID"
//	@Param			spectatorId	path		string	true	"Spectator ID"
//	@Param			request			body		object	true	"Specify inLeagueName"
//	@Success		200				{object}	models.Spectator
//	@Failure		400				{object}	responses.ErrorResponse
//	@Failure		404				{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators/{spectatorId}/profile [put]
func (c *spectatorControllerImpl) UpdateProfile(ctx *gin.Context) {
	spectatorID, err := uuid.Parse(ctx.Param("spectatorId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		InLeagueName *string `json:"InLeagueName" validate:"omitempty,min=3,max=20"`
	}
	if err = ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Invalid request payload")
		return
	}

	spectator, err := c.spectatorService.UpdateProfile(spectatorID, req.InLeagueName)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateProfile")
		c.handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, spectator)
}

// UpdateRole godoc
//
//	@Summary		Update a spectator's role
//	@Description	Change a spectator's role to another spectator role. Requires update:role.
//	@Tags			Spectators
//	@Accept			json
//	@Produce		json
//	@Param			leagueId		path		string			true	"League ID"
//	@Param			spectatorId	path		string			true	"Spectator ID"
//	@Param			request			body		object			true	"Role"
//	@Success		200				{object}	models.Spectator
//	@Failure		400				{object}	responses.ErrorResponse
//	@Failure		404				{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators/{spectatorId}/role [put]
func (c *spectatorControllerImpl) UpdateRole(ctx *gin.Context) {
	spectatorID, err := uuid.Parse(ctx.Param("spectatorId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	var req struct {
		Role string `json:"Role" validate:"required"`
	}
	if err = ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "Missing or invalid Role")
		return
	}

	newRole, ok := rbac.ParseMemberRole(req.Role)
	if !ok {
		sendError(ctx, http.StatusBadRequest, "Unrecognized role")
		return
	}

	spectator, err := c.spectatorService.UpdateRole(spectatorID, newRole)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "UpdateRole")
		c.handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, spectator)
}

// DeleteSpectator godoc
//
//	@Summary		Delete a spectator (staff)
//	@Description	Remove a spectator from a league. Requires delete:spectator.
//	@Tags			Spectators
//	@Produce		json
//	@Param			leagueId		path		string	true	"League ID"
//	@Param			spectatorId	path		string	true	"Spectator ID"
//	@Success		200				{object}	map[string]interface{}
//	@Failure		400				{object}	responses.ErrorResponse
//	@Failure		404				{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/spectators/{spectatorId} [delete]
func (c *spectatorControllerImpl) DeleteSpectator(ctx *gin.Context) {
	spectatorID, err := uuid.Parse(ctx.Param("spectatorId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err := c.spectatorService.DeleteSpectator(spectatorID); err != nil {
		c.logger.Error("Service method error", "error", err, "method", "DeleteSpectator")
		c.handleServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Spectator removed from the league"})
}

func (c *spectatorControllerImpl) handleServiceError(ctx *gin.Context, err error) {
	switch err {
	case types.ErrSpectatorNotFound, types.ErrLeagueNotFound:
		sendError(ctx, http.StatusNotFound, err.Error())
	case types.ErrUserAlreadyInLeague, types.ErrUserAlreadySpectator, types.ErrCannotDemoteLeagueOwner:
		sendError(ctx, http.StatusConflict, err.Error())
	case types.ErrInvalidInput:
		sendError(ctx, http.StatusBadRequest, err.Error())
	default:
		sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
	}
}
