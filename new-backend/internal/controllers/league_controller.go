package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeagueController struct {
	leagueRepo         *repositories.LeagueRepository
	userRepo           *repositories.UserRepository
	playerRepo         *repositories.PlayerRepository
	leaguePokemonRepo  *repositories.LeaguePokemonRepository
	draftedPokemonRepo *repositories.DraftedPokemonRepository
	gameRepo           *repositories.GameRepository
}

type leagueRequest struct {
	Name                  string     `json:"name" binding:"required"`
	RulesetID             *uuid.UUID `json:"ruleset_id"`
	MaxPokemonPerPlayer   uint       `json:"max_pokemon_per_player" binding:"gte=1"`
	StartingDraftPoints   uint       `json:"starting_draft_points" binding:"gte=20"`
	AllowWeeklyFreeAgents bool       `json:"allow_free_agents"`
	StartDate             time.Time  `json:"start_date" binding:"required,datetime=02/01/2006"`
	EndDate               *time.Time `json:"end_date" binding:"omitempty,datetime=02/01/2006"`
}

func NewLeagueController(leagueRepo *repositories.LeagueRepository,
	userRepo *repositories.UserRepository,
	playerRepo *repositories.PlayerRepository,
	leaguePokemonRepo *repositories.LeaguePokemonRepository,
	draftedPokemonRepo *repositories.DraftedPokemonRepository,
	gameRepo *repositories.GameRepository) LeagueController {

	return LeagueController{
		userRepo:           userRepo,
		playerRepo:         playerRepo,
		leaguePokemonRepo:  leaguePokemonRepo,
		draftedPokemonRepo: draftedPokemonRepo,
		gameRepo:           gameRepo,
	}
}
func (ctrl *LeagueController) CreateLeague(ctx *gin.Context) {
	const maxLeaguesCommisionable = 2
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: CreateLeague) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return
	}

	count, err := ctrl.leagueRepo.GetLeaguesCountByCommissioner(currentUser.ID)
	if err != nil {
		log.Printf("(Error: CreateLeague) - Could not get commisioner league count: %w\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if count >= maxLeaguesCommisionable {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Max League Creation Limit Reached: %d", maxLeaguesCommisionable)})
		return
	}

	var req leagueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("(CreateLeague) - Received league creation request: %v", req)

	league := &models.League{
		Name:                  req.Name,
		CommissionerUserID:    currentUser.ID,
		RulesetID:             req.RulesetID,
		MaxPokemonPerPlayer:   req.MaxPokemonPerPlayer,
		StartingDraftPoints:   req.StartingDraftPoints,
		AllowWeeklyFreeAgents: req.AllowWeeklyFreeAgents,
		StartDate:             req.StartDate,
		EndDate:               req.EndDate,
	}

	_, err = ctrl.leagueRepo.CreateLeague(league)
	if err != nil {
		log.Printf("(Error: CreateLeague) - transaction failed: %w\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
}

// endpoint: /api/league/:id
func (ctrl *LeagueController) GetLeague(ctx *gin.Context) (*models.League, error) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		log.Printf("(Error: GetLeague) - no user in context\n")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User information not available"})
		return nil, errors.New("no user in context")
	}

	leagueIDStr := ctx.Param("id") // matches the :id in your route definition paramater

	leagueID, err := uuid.Parse(leagueIDStr)
	if err != nil {
		log.Printf("(Error: GetLeagueByIDHandler) - Invalid league ID format: %w\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID format"})
		return nil, err
	}

	isPlayer, err := ctrl.leagueRepo.IsUserPlayerInLeague(currentUser.ID, leagueID)
	if err != nil {
		log.Printf("(Error: GetLeague) - user in league check failed: %w\n", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return nil, err
	}

	if !isPlayer {
		err = errors.New("You are not authorized")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}

	return ctrl.leagueRepo.GetLeagueByID(leagueID)
}
