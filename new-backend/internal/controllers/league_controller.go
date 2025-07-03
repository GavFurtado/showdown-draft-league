package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
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
	leagueService services.LeagueService
}

func NewLeagueController(leagueService services.LeagueService) *leagueControllerImpl {
	return &leagueControllerImpl{
		leagueService: leagueService,
	}
}

// POST /api/leagues
// creates a new league if the current user has less than 2 Leagues commisioned
func (ctrl *leagueControllerImpl) CreateLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: CreateLeague) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	var req common.LeagueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("(CreateLeague) - Received league creation request: %v", req)

	league, err := ctrl.leagueService.CreateLeague(currentUser.ID, req)
	if err != nil {
		log.Printf("(Error: CreateLeague) - Service failed: %v\n", err)
		// Check for specific service errors to return appropriate HTTP status
		if err.Error() == fmt.Sprintf("max league creation limit reached: %d", 2) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create league"})
		return
	}

	ctx.JSON(http.StatusCreated, league)
}

// GET /api/leagues/:id
// handles fetching a league by ID for an authorized user.
func (ctrl *leagueControllerImpl) GetLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetLeague) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrNoUserInContext.Error()})
		return
	}

	leagueIDStr := ctx.Param("id")

	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		log.Printf("(Error: GetLeague) - Invalid league ID format: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": common.ErrParsingParams.Error()})
		return
	}

	league, err := ctrl.leagueService.GetLeagueByIDForUser(currentUser.ID, leagueID)
	if err != nil {
		log.Printf("(Error: GetLeague) - Service failed: %v\n", err)

		if err.Error() == "not authorized to view this league" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "failed to retrieve league: record not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
			return
		}
		// other errors
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve league"})
		return
	}

	ctx.JSON(http.StatusOK, league)
}
