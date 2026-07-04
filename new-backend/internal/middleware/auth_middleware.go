package middleware

import (
	"net/http"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/dtos/responses"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/types"
	"github.com/gin-gonic/gin"
	uuid "github.com/google/uuid"
)

type AuthMiddlewareDependencies struct {
	UserRepo    repositories.UserRepository
	JWTService  *services.JWTService
	RBACService services.RBACService
}
type LeagueRBACDependencies struct {
	UserRepo    repositories.UserRepository
	RBACService services.RBACService
}

func AuthMiddleware(
	deps AuthMiddlewareDependencies,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("token")
		if err != nil {
			GetLogger(ctx).Warn("missing or invalid token cookie", "error", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, responses.NewErrorResponse(http.StatusUnauthorized, "Missing or invalid token", ctx.Request.URL.Path))
			return
		}

		// validate token
		userID, err := deps.JWTService.ValidateToken(token)
		if err != nil {
			GetLogger(ctx).Warn("invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token", ctx.Request.URL.Path))
			return
		}

		user, err := deps.UserRepo.GetUserByID(userID)
		if err != nil {
			if err.Error() == "record not found" {
				GetLogger(ctx).Warn("user ID not found in DB", "user_id", userID, "error", err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, responses.NewErrorResponse(http.StatusUnauthorized, "User not found", ctx.Request.URL.Path))
				return
			}
			// Other errors (DB)
			GetLogger(ctx).Error("database error fetching user", "user_id", userID, "error", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, responses.NewErrorResponse(http.StatusInternalServerError, "Internal Server Error during authentication", ctx.Request.URL.Path))
			return
		}

		SetUserOnLogger(ctx, userID.String(), string(user.Role))

		ctx.Set("currentUser", user)
		ctx.Set("currentUserID", userID)
		ctx.Set("role", user.Role)

		ctx.Next()
	}
}

// LeagueRBACMiddleware checks for league-specific permissions.
func LeagueRBACMiddleware(
	deps LeagueRBACDependencies,
	requiredPermission rbac.Permission,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUser, exists := GetUserFromContext(ctx)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusNotFound, responses.NewErrorResponse(http.StatusNotFound, "User not found in context", ctx.Request.URL.Path))
			return
		}

		// bypass checks if user is an admin
		if currentUser.Role == "admin" {
			GetLogger(ctx).Info("rbac bypass for admin user", "username", currentUser.DiscordUsername, "user_id", currentUser.ID)
			ctx.Next()
			return
		}

		leagueIDStr := ctx.Param("leagueId")
		leagueID, err := uuid.Parse(leagueIDStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, responses.NewErrorResponse(http.StatusBadRequest, "Invalid league ID format", ctx.Request.URL.Path))
			return
		}

		// Check if the user has the required permission for the league
		player, ok, err := deps.RBACService.CanAccess(currentUser.ID, leagueID, requiredPermission)
		if err != nil {
			if err == types.ErrInternalService {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, responses.NewErrorResponse(http.StatusInternalServerError, "Internal Server Error", ctx.Request.URL.Path))
				return
			}
			// some record not found error (atleast it should be)
			GetLogger(ctx).Error("rbac access check failed", "error", err)
			ctx.AbortWithStatusJSON(http.StatusNotFound, responses.NewErrorResponse(http.StatusNotFound, "Record Not Found", ctx.Request.URL.Path))
			return
		}
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, responses.NewErrorResponse(http.StatusForbidden, "Forbidden: Insufficient permissions for this league", ctx.Request.URL.Path))
			return
		}

		// Set playerID and player role and player in the context for downstream handlers
		if player != nil {
			ctx.Set("playerID", player.ID)
			ctx.Set("playerRole", player.Role)
			ctx.Set("player", player)
		}

		ctx.Next()
	}
}

// Helper for Controllers to get current user context
func GetUserFromContext(ctx *gin.Context) (*models.User, bool) {
	val, exists := ctx.Get("currentUser")
	if !exists {
		return nil, false
	}

	user, ok := val.(*models.User)
	return user, ok
}
