package controllers

import (
	"log"
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/requests"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/middleware"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeagueMemberController interface {
	GetByID(ctx *gin.Context)
	GetByUserAndLeague(ctx *gin.Context)
	GetByLeague(ctx *gin.Context)
	GetByUser(ctx *gin.Context)
	GetWithFullRoster(ctx *gin.Context)
	JoinLeague(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)
	UpdateDraftPoints(ctx *gin.Context)
	UpdateRecord(ctx *gin.Context)
	UpdateDraftPosition(ctx *gin.Context)
}

type leagueMemberControllerImpl struct {
	leagueMemberService services.LeagueMemberService
}

func NewLeagueMemberController(leagueMemberService services.LeagueMemberService) LeagueMemberController {
	return &leagueMemberControllerImpl{
		leagueMemberService: leagueMemberService,
	}
}

func (c *leagueMemberControllerImpl) GetByID(ctx *gin.Context) {
	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	member, err := c.leagueMemberService.GetByID(memberID)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: GetByID) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPlayerNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) GetByUserAndLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	userIDStr := ctx.DefaultQuery("userId", "")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid userId query"})
		return
	}

	member, err := c.leagueMemberService.GetByUserAndLeague(userID, leagueID)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: GetByUserAndLeague) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPlayerNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) GetByLeague(ctx *gin.Context) {
	leagueID, err := uuid.Parse(ctx.Param("leagueId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	members, err := c.leagueMemberService.GetByLeague(leagueID)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: GetByLeague) - Service method error: %v\n", err)
		switch err {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, members)
}

func (c *leagueMemberControllerImpl) GetByUser(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	members, err := c.leagueMemberService.GetByUser(userID)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: GetByUser) - Service method error: %v\n", err)
		switch err {
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, members)
}

func (c *leagueMemberControllerImpl) GetWithFullRoster(ctx *gin.Context) {
	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	member, err := c.leagueMemberService.GetWithFullRoster(memberID)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: GetWithFullRoster) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": types.ErrPlayerNotFound.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": types.ErrInternalService.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) JoinLeague(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	var req requests.PlayerCreateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields in payload"})
		return
	}

	if currentUser.Role != "admin" && req.UserID != currentUser.ID {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Cannot perform this transaction"})
		return
	}

	if req.UserID == uuid.Nil || req.LeagueID == uuid.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bad or Malformed Request"})
		return
	}

	member, err := c.leagueMemberService.Create(currentUser, &req)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: JoinLeague) - Service method error: %v\n", err)
		switch err {
		case types.ErrUserNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrUserAlreadyInLeague, types.ErrInLeagueNameTaken, types.ErrTeamNameTaken:
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case types.ErrInternalService, types.ErrFailedToCreatePlayer:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case types.ErrInvalidState:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) UpdateProfile(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	var req requests.UpdatePlayerInfoRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	if req.InLeagueName == nil && req.TeamName == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	member, err := c.leagueMemberService.UpdateProfile(currentUser, memberID, req.InLeagueName, req.TeamName)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: UpdateProfile) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrInLeagueNameTaken, types.ErrTeamNameTaken:
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case types.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case types.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) UpdateDraftPoints(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	var req struct {
		DraftPoints *int `json:"DraftPoints" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid DraftPoints"})
		return
	}

	member, err := c.leagueMemberService.UpdateDraftPoints(currentUser, memberID, req.DraftPoints)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: UpdateDraftPoints) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case types.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) UpdateRecord(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	var req struct {
		Wins   *int `json:"Wins" binding:"required"`
		Losses *int `json:"Losses" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid Wins/Losses"})
		return
	}

	member, err := c.leagueMemberService.UpdateRecord(currentUser, memberID, *req.Wins, *req.Losses)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: UpdateRecord) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case types.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (c *leagueMemberControllerImpl) UpdateDraftPosition(ctx *gin.Context) {
	currentUser, exists := middleware.GetUserFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrNoUserInContext.Error()})
		return
	}

	memberID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": types.ErrParsingParams.Error()})
		return
	}

	var req struct {
		DraftPosition *int `json:"DraftPosition" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid DraftPosition"})
		return
	}

	member, err := c.leagueMemberService.UpdateDraftPosition(currentUser, memberID, *req.DraftPosition)
	if err != nil {
		log.Printf("LOG: (LeagueMemberController: UpdateDraftPosition) - Service method error: %v\n", err)
		switch err {
		case types.ErrPlayerNotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case types.ErrInternalService:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		case types.ErrUnauthorized:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, member)
}
