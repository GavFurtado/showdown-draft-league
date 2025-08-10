package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/services"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *services.JWTService, userRepo repositories.UserRepository, leagueService *services.LeagueService, rbacService *services.RBACService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// validate token
		userID, err := jwtService.ValidateToken(token)
		if err != nil {
			log.Printf("(Error: AuthMiddleware): Invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		user, err := userRepo.GetUserByID(userID)
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

		// Layer 2: RBAC check for league-specific routes
		// Extract league ID from request context or URL parameters if necessary
		leagueID := ctx.Param("leagueID") // Assuming league ID is in the URL

		// Bypass RBAC if user is an admin
		if user.Role == "admin" {
			ctx.Next()
			return
		}

		// Implement league-specific RBAC check here
		requiredPermission := "can_view_league" // Example permission
		if !rbacService.CanAccess(user, requiredPermission) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
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

	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// validate token
		userID, err := jwtService.ValidateToken(token)
		if err != nil {
			log.Printf("(Error: AuthMiddleware): Invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		user, err := userRepo.GetUserByID(userID)
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

		// Layer 2: RBAC check for league-specific routes
		// Extract league ID from request context or URL parameters if necessary
		// For now, assume league ID is available or not needed for all protected routes
		// Example: leagueID := getLeagueIDFromRequest(r) // Implement this helper

		// Bypass RBAC if user is an admin
		if user.Role == "admin" {
			ctx.Next()
			return
		}

		// TODO: Implement league-specific RBAC check here
		// This will involve checking user.Role against required permissions for the specific league/route
		// For example:
		// if !rbacService.CanAccess(user, leagueID, requiredPermission) {
		// 	http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
		// 	return
		// }

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
