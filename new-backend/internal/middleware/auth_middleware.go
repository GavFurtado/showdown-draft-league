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

func AuthMiddleware(jwtService *services.JWTService, userRepo *repositories.UserRepository) gin.HandlerFunc {
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
			if err.Error() == "record not found" { // string check is prolly not optimal but shouldn't affect performance much
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
	}
}

// Helper for Controllers to get current user context
func GetUserFromContext(ctx *gin.Context) (*models.User, bool) {
	val, exists := ctx.Get("currentuser")
	if !exists {
		return nil, false
	}

	user, ok := val.(*models.User)
	return user, ok
}
