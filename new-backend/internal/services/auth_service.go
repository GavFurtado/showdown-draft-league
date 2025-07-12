package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/common"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/repositories"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// defines the interface for authentication-related business logic.
type AuthService interface {
	HandleDiscordCallback(ctx context.Context, code string) (*models.User, string, error)
	VerifyToken(token string) (uuid.UUID, error)
}

type authServiceImpl struct {
	userRepo           repositories.UserRepository
	jwtService         *JWTService
	discordOauthConfig *oauth2.Config
}

func (s *authServiceImpl) VerifyToken(token string) (uuid.UUID, error) {
	return s.jwtService.ValidateToken(token)
}

// creates a new instance of AuthService, receiving the pre-configured oauth2.Config.
func NewAuthService(
	userRepo repositories.UserRepository,
	jwtService *JWTService,
	oauthConfig *oauth2.Config,
) AuthService {
	return &authServiceImpl{
		userRepo:           userRepo,
		jwtService:         jwtService,
		discordOauthConfig: oauthConfig,
	}
}

// encapsulates the business logic for Discord OAuth callback.
func (s *authServiceImpl) HandleDiscordCallback(ctx context.Context, code string) (*models.User, string, error) {
	// Exchange the authorization code for an access token using the injected config
	token, err := s.discordOauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code for token: %w", err)
	}

	discordUser, err := getDiscordUserInfo(token.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get Discord user info: %w", err)
	}

	user, err := s.userRepo.GetUserByDiscordID(discordUser.ID)
	if err != nil {
		if err.Error() == "record not found" {
			log.Printf("Creating new user for Discord ID: %s", discordUser.ID)
			newUser := models.User{
				DiscordID:        discordUser.ID,
				DiscordUsername:  fmt.Sprintf("%s#%s", discordUser.Username, discordUser.Discriminator), // Combine if discriminator still exists for old users
				DiscordAvatarURL: getDiscordAvatarURL(discordUser.ID, discordUser.Avatar),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			createdUser, createErr := s.userRepo.CreateUser(&newUser)
			if createErr != nil {
				return nil, "", fmt.Errorf("failed to create new user: %w", createErr)
			}
			user = createdUser
		} else {
			return nil, "", fmt.Errorf("database error during user check: %w", err)
		}
	} else {
		// User exists, update their username/avatar if needed
		updatedUsername := discordUser.Username
		if discordUser.Discriminator != "0" && discordUser.Discriminator != "" {
			updatedUsername = fmt.Sprintf("%s#%s", discordUser.Username, discordUser.Discriminator)
		}

		// actual updating
		if user.DiscordUsername != updatedUsername || user.DiscordAvatarURL != getDiscordAvatarURL(discordUser.ID, discordUser.Avatar) {
			user.DiscordUsername = updatedUsername
			user.DiscordAvatarURL = getDiscordAvatarURL(discordUser.ID, discordUser.Avatar)
			user.UpdatedAt = time.Now()
			if _, updateErr := s.userRepo.UpdateUser(user); updateErr != nil {
				log.Printf("Warning: Failed to update user details for Discord ID %s: %v", discordUser.ID, updateErr)
			}
		}
	}

	// Generate JWT token
	jwtToken, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT for user %s: %w", user.ID, err)
	}

	return user, jwtToken, nil
}

// Helper functions
func getDiscordUserInfo(accessToken string) (*common.DiscordUser, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("(Error: getDiscordUserInfo) - Failed to create Discord user info request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("(Error: getDiscordUserInfo) - Failed to create Discord user info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("(Error: getDiscordUserInfo) - Discord API returned non-OK status: %d, body: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var discordUser common.DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return nil, fmt.Errorf("failed to decode Discord user info: %w", err)
	}
	return &discordUser, nil
}

func getDiscordAvatarURL(userID, avatarHash string) string {
	if avatarHash == "" {
		return ""
	}

	if strings.HasPrefix(avatarHash, "a_") {
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.gif", userID, avatarHash)
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", userID, avatarHash)
}
