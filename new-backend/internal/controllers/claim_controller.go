package controllers

import (
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
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
	claimService services.ClaimService
}

func NewClaimController(claimService services.ClaimService) ClaimController {
	return &claimControllerImpl{
		claimService: claimService,
	}
}

func (c *claimControllerImpl) GetByID(ctx *gin.Context) {
	claimID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	claim, err := c.claimService.GetByID(claimID)
	if err != nil {
		log.Printf("LOG: (ClaimController: GetByID) - Service method error: %v\n", err)
		switch err {
		case types.ErrClaimNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrClaimNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, claim)
}

func (c *claimControllerImpl) GetActiveByPlayer(ctx *gin.Context) {
	playerID, err := uuid.Parse(ctx.Param("playerId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	claims, err := c.claimService.GetActiveByPlayer(playerID)
	if err != nil {
		log.Printf("LOG: (ClaimController: GetActiveByPlayer) - Service method error: %v\n", err)
		switch {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, claims)
}

func (c *claimControllerImpl) GetActiveByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	claims, err := c.claimService.GetActiveByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (ClaimController: GetActiveByLeague) - Service method error: %v\n", err)
		switch {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, claims)
}

func (c *claimControllerImpl) GetReleasedByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	claims, err := c.claimService.GetReleasedByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (ClaimController: GetReleasedByLeague) - Service method error: %v\n", err)
		switch {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, claims)
}
