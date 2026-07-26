package controllers

import (
	"log/slog"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransferController interface {
	StartTransferPeriod(ctx *gin.Context)
	EndTransferPeriod(ctx *gin.Context)
	DropPokemon(ctx *gin.Context)
	PickupFreeAgent(ctx *gin.Context)
}

type transferControllerImpl struct {
	logger          *slog.Logger
	transferService services.TransferService
}

func NewTransferController(logger *slog.Logger, transferService services.TransferService) TransferController {
	return &transferControllerImpl{
		logger:          utils.LoggerWithService(logger, "TransferController"),
		transferService: transferService,
	}
}

// StartTransferPeriod godoc
//
//	@Summary		Start transfer period
//	@Description	Open the transfer window for a league
//	@Tags			Transfers
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/transfers/start [post]
func (tc *transferControllerImpl) StartTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, "Invalid league ID")
		return
	}

	if err := tc.transferService.StartTransferPeriod(leagueID); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to start transfer period")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period started successfully"})
}

// EndTransferPeriod godoc
//
//	@Summary		End transfer period
//	@Description	Close the transfer window for a league
//	@Tags			Transfers
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/transfers/end [post]
func (tc *transferControllerImpl) EndTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		sendError(c, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err := tc.transferService.EndTransferPeriod(leagueID); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to end transfer period")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period ended successfully"})
}

// DropPokemon godoc
//
//	@Summary		Drop a Pokemon
//	@Description	Release a Pokemon from your roster
//	@Tags			Transfers
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			claimId		path		string	true	"Claim ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		401			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/transfers/drop/{claimId} [post]
func (tc *transferControllerImpl) DropPokemon(ctx *gin.Context) {
	currentUser, err := tc.getUserFromContext(ctx)
	if err != nil {
		return
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	claimID, err := uuid.Parse(ctx.Param("claimId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err := tc.transferService.DropPokemon(currentUser, leagueID, claimID); err != nil {
		tc.logger.Error("Service method error", "error", err, "method", "DropPokemon")
		switch err {
		case types.ErrClaimNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrUnauthorized:
			sendError(ctx, http.StatusUnauthorized, err.Error())
		case types.ErrInvalidState:
			sendError(ctx, http.StatusConflict, "League is not in a transfer window")
		case types.ErrPokemonAlreadyReleased:
			sendError(ctx, http.StatusConflict, "Pokemon has already been released")
		case types.ErrInsufficientTransferCredits:
			sendError(ctx, http.StatusForbidden, err.Error())
		case types.ErrForbidden:
			sendError(ctx, http.StatusForbidden, "Pokemon not in this league")
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "pokemon dropped successfully"})
}

// PickupFreeAgent godoc
//
//	@Summary		Pick up a free agent
//	@Description	Sign an unclaimed Pokemon
//	@Tags			Transfers
//	@Produce		json
//	@Param			leagueId	path		string	true	"League ID"
//	@Param			poolEntryId	path		string	true	"Pool Entry ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	responses.ErrorResponse
//	@Failure		403			{object}	responses.ErrorResponse
//	@Failure		404			{object}	responses.ErrorResponse
//	@Failure		409			{object}	responses.ErrorResponse
//	@Router			/api/leagues/{leagueId}/transfers/pickup/{poolEntryId} [post]
func (tc *transferControllerImpl) PickupFreeAgent(ctx *gin.Context) {
	currentUser, err := tc.getUserFromContext(ctx)
	if err != nil {
		return
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	poolEntryID, err := uuid.Parse(ctx.Param("poolEntryId"))
	if err != nil {
		sendError(ctx, http.StatusBadRequest, types.ErrParsingParams.Error())
		return
	}

	if err := tc.transferService.PickupFreeAgent(currentUser, leagueID, poolEntryID); err != nil {
		tc.logger.Error("Service method error", "error", err, "method", "PickupFreeAgent")
		switch err {
		case types.ErrPoolEntryNotFound:
			sendError(ctx, http.StatusNotFound, err.Error())
		case types.ErrInsufficientTransferCredits:
			sendError(ctx, http.StatusForbidden, err.Error())
		case types.ErrInvalidState:
			sendError(ctx, http.StatusConflict, "League is not in a transfer window")
		case types.ErrConflict:
			sendError(ctx, http.StatusConflict, "Pokemon is not available to sign")
		case types.ErrForbidden:
			sendError(ctx, http.StatusForbidden, "Pokemon not in this league")
		default:
			sendError(ctx, http.StatusInternalServerError, types.ErrInternalService.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "free agent signed successfully"})
}

// Helpers
func (tc *transferControllerImpl) getUserFromContext(ctx *gin.Context) (*models.User, error) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		tc.logger.Error("no user in context", "method", "getUserFromContext")
		sendError(ctx, http.StatusBadRequest, types.ErrNoUserInContext.Error())
		return nil, types.ErrNoUserInContext
	}
	return currentUser, nil
}
