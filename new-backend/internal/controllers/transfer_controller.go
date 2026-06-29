package controllers

import (
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
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
	transferService services.TransferService
}

func NewTransferController(transferService services.TransferService) TransferController {
	return &transferControllerImpl{
		transferService: transferService,
	}
}

// StartTransferPeriod handles the POST /api/leagues/:leagueId/transfers/start endpoint.
// It initiates the transfer window for a specific league.
func (tc *transferControllerImpl) StartTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	if err := tc.transferService.StartTransferPeriod(leagueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transfer period", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period started successfully"})
}

func (tc *transferControllerImpl) EndTransferPeriod(c *gin.Context) {
	leagueIDStr := c.Param("leagueId")
	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	if err := tc.transferService.EndTransferPeriod(leagueID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end transfer period", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer period ended successfully"})
}

func (tc *transferControllerImpl) DropPokemon(ctx *gin.Context) {
	currentUser, err := tc.getUserFromContext(ctx)
	if err != nil {
		return // response already sent
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	claimID, err := uuid.Parse(ctx.Param("claimId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	if err := tc.transferService.DropPokemon(currentUser, leagueID, claimID); err != nil {
		log.Printf("LOG: (TransferController: DropPokemon) - Service method error: %v\n", err)
		switch err {
		case types.ErrClaimNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case types.ErrInvalidState:
			ctx.JSON(http.StatusConflict, gin.H{"error": "League is not in a transfer window"})
		case types.ErrPokemonAlreadyReleased:
			ctx.JSON(http.StatusConflict, gin.H{"error": "Pokemon has already been released"})
		case types.ErrInsufficientTransferCredits:
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case types.ErrForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Pokemon not in this league"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "pokemon dropped successfully"})
}

func (tc *transferControllerImpl) PickupFreeAgent(ctx *gin.Context) {
	currentUser, err := tc.getUserFromContext(ctx)
	if err != nil {
		return // response already sent
	}

	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	poolEntryID, err := uuid.Parse(ctx.Param("poolEntryId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	if err := tc.transferService.PickupFreeAgent(currentUser, leagueID, poolEntryID); err != nil {
		log.Printf("LOG: (TransferController: PickupFreeAgent) - Service method error: %v\n", err)
		switch err {
		case types.ErrPoolEntryNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrInsufficientTransferCredits:
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case types.ErrInvalidState:
			ctx.JSON(http.StatusConflict, gin.H{"error": "League is not in a transfer window"})
		case types.ErrConflict:
			ctx.JSON(http.StatusConflict, gin.H{"error": "Pokemon is not available to sign"})
		case types.ErrForbidden:
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Pokemon not in this league"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "free agent signed successfully"})
}

// Helpers
func (tc *transferControllerImpl) getUserFromContext(ctx *gin.Context) (*models.User, error) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("LOG: (TransferController: getUserFromContext) - no user in context\n")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return nil, types.ErrNoUserInContext
	}
	return currentUser, nil
}

