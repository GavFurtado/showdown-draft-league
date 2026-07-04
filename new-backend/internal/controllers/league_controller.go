package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeagueController interface {
	// creates a new league if the user has less than 2 leagues commissioned
	CreateLeague(ctx *gin.Context)
	// handles fetching a league by ID for an authorized user.
	GetLeague(ctx *gin.Context)
}

type leagueControllerImpl struct {
	logger        *slog.Logger
	leagueService services.LeagueService
}

func NewLeagueController(logger *slog.Logger, leagueService services.LeagueService) LeagueController {
	return &leagueControllerImpl{
		logger:        utils.LoggerWithService(logger, "LeagueController"),
		leagueService: leagueService,
	}
}

// CreateLeague godoc
//
//	@Summary		Create a league
//	@Description	Create a new league
//	@Tags			Leagues
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.LeagueCreateRequestDTO	true	"League config"
//	@Success		200		{object}	models.League
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Router			/api/leagues [post]
func (ctrl *leagueControllerImpl) CreateLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctrl.logger.Error("no user in context", "method", "CreateLeague")
		sendError(ctx, http.StatusUnauthorized, types.ErrNoUserInContext.Error())
		return
	}

	var req requests.LeagueCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendError(ctx, http.StatusBadRequest, "bad request")
		return
	}

	ctrl.logger.Info("Received league creation request", "request", req)

	league, err := ctrl.leagueService.CreateLeague(currentUser.ID, &req)
	if err != nil {
		ctrl.logger.Error("Service failed", "error", err, "method", "CreateLeague")
		switch {
		case errors.Is(err, types.ErrMaxLeagueCreationLimitReached):
			sendError(ctx, http.StatusBadRequest, types.ErrMaxLeagueCreationLimitReached.Error())
		case errors.Is(err, types.ErrExceedsMaxAllowableGroupCount):
			sendError(ctx, http.StatusBadRequest, types.ErrExceedsMaxAllowableGroupCount.Error())
		case errors.Is(err, types.ErrInvalidLeagueConfiguration):
			sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, "Failed to create league")
		}
		return
	}

	ctx.JSON(http.StatusOK, league)
}

// GetLeague godoc
//
//	@Summary		Get a league
//	@Description	League details and settings
//	@Tags			Leagues
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	models.League
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		403			{object}	map[string]interface{}
//	@Failure		404			{object}	map[string]interface{}
//	@Router			/api/leagues/{leagueId} [get]
func (ctrl *leagueControllerImpl) GetLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctrl.logger.Error("no user in context", "method", "GetLeague")
		sendError(ctx, http.StatusInternalServerError, types.ErrNoUserInContext.Error())
		return
	}

	leagueIDStr := ctx.Param("leagueId")

	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		ctrl.logger.Error("Invalid league ID format", "error", err)
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	league, err := ctrl.leagueService.GetLeagueByIDForUser(currentUser.ID, leagueID)
	if err != nil {
		ctrl.logger.Error("Service failed", "error", err, "method", "GetLeague")

		if err.Error() == "not authorized to view this league" {
			sendError(ctx, http.StatusForbidden, err.Error())
			return
		}
		if err.Error() == "failed to retrieve league: record not found" {
			sendError(ctx, http.StatusNotFound, "League not found")
			return
		}
		sendError(ctx, http.StatusInternalServerError, "Failed to retrieve league")
		return
	}

	ctx.JSON(http.StatusOK, league)
}

