package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/rbac"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
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
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			log.Printf("(Middleware: AuthMiddleware) Missing token\n")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// validate token
		userID, err := deps.JWTService.ValidateToken(token)
		if err != nil {
			log.Printf("(Error: AuthMiddleware): Invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		user, err := deps.UserRepo.GetUserByID(userID)
		if err != nil {
			if err.Error() == "record not found" {
				log.Printf("(Error: AuthMiddleware): User ID %s not found in DB: %v", userID, err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}
			// Other errors (DB)
			log.Printf("(Error: AuthMiddleware): Database error fetching user %s: %v", userID, err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error during authentication"})
			return
		}

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
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
			return
		}

		// bypass checks if user is an admin
		if currentUser.Role == "admin" {
			log.Printf("LOG: [BYPASS: RBAC Middleware]: Skipped check for admin user %s (%s)\n", currentUser.DiscordUsername, currentUser.ID)
			ctx.Next()
		}

		leagueIDStr := ctx.Param("leagueId")
		leagueID, err := uuid.Parse(leagueIDStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID format"})
			return
		}

		// Check if the user has the required permission for the league
		// requires the right permission and have a valid player in the league
		if ok, err := deps.RBACService.CanAccess(currentUser.ID, leagueID, requiredPermission); !ok {
			if err != nil {
				if err == common.ErrInternalService {
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
					return
				}
				// some record not found error (atleast it should be)
				log.Printf("(Error: LeagueRBACMiddleware) - %s", err.Error())
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record Not Found"})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions for this league"})
			return
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
