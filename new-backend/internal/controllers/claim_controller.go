package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClaimController interface {
	GetByID(ctx *gin.Context)
	GetActiveByPlayer(ctx *gin.Context)
	GetActiveByLeague(ctx *gin.Context)
	GetReleasedByLeague(ctx *gin.Context)
}

type claimControllerImpl struct {
	logger       *slog.Logger
	claimService services.ClaimService
}

func NewClaimController(logger *slog.Logger, claimService services.ClaimService) ClaimController {
	return &claimControllerImpl{
		logger:       utils.LoggerWithService(logger, "ClaimController"),
		claimService: claimService,
	}
}

// GetByID godoc
//
//	@Summary		Get a claim
//	@Description	Claim details
//	@Tags			Claims
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			id			path		string	true	"Claim ID"
//	@Success		200			{object}	models.Claim
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		500			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/claims/{id} [get]
func (c *claimControllerImpl) GetByID(ctx *gin.Context) {
	claimID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	claim, err := c.claimService.GetByID(claimID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetByID")
		switch err {
		case types.ErrClaimNotFound:
			sendError(ctx, http.StatusNotFound, types.ErrClaimNotFound.Error())
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, claim)
}

// GetActiveByPlayer godoc
//
//	@Summary		Get claims by player
//	@Description	Active claims for a specific player
//	@Tags			Claims
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			playerId	path		string	true	"Player ID"
//	@Success		200			{array}		models.Claim
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		500			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/claims/player/{playerId} [get]
func (c *claimControllerImpl) GetActiveByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	claims, err := c.claimService.GetActiveByPlayer(playerID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetActiveByPlayer")
		switch {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, claims)
}

// GetActiveByLeague godoc
//
//	@Summary		Get active claims
//	@Description	All active claims in a league
//	@Tags			Claims
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.Claim
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		500			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/claims [get]
func (c *claimControllerImpl) GetActiveByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	claims, err := c.claimService.GetActiveByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetActiveByLeague")
		switch {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, claims)
}

// GetReleasedByLeague godoc
//
//	@Summary		Get released Pokemon
//	@Description	All released Pokemon in a league
//	@Tags			Claims
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{array}		models.Claim
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		500			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/claims/released [get]
func (c *claimControllerImpl) GetReleasedByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	claims, err := c.claimService.GetReleasedByLeague(leagueID)
	if err != nil {
		c.logger.Error("Service method error", "error", err, "method", "GetReleasedByLeague")
		switch {
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, claims)
}
