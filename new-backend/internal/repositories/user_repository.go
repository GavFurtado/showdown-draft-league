package repositories

import (
	"github.com/GavFurtado/showdown-draft-league/new-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByDiscordID(discordID string) (*models.User, error)
	UpdateUser(user *models.User) (*models.User, error)
	// fetches all Leagues that a specific user is a player in.
	GetUserLeagues(userID uuid.UUID) ([]*models.League, error)
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepositoryImpl {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(user *models.User) (*models.User, error) {
	err := r.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}

// retrieves user by internal user id
func (r *userRepositoryImpl) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User

	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

// retrieves a user by their Discord ID
func (r *userRepositoryImpl) GetUserByDiscordID(discordID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("discord_id = ?", discordID).First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateUser(user *models.User) (*models.User, error) {
	err := r.db.Save(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// fetches all Leagues that a specific user is a player in.
func (r *userRepositoryImpl) GetUserLeagues(userID uuid.UUID) ([]*models.League, error) {
	var user models.User

	// Fetch the user, preloading their players and each player's associated league.
	err := r.db.
		Preload("Players").        // Preload the Player records for this user
		Preload("Players.League"). // For each Player, preload its associated League
		Where("id = ?", userID).   // Find the specific user
		First(&user).Error         // Fetch the user

	if err != nil {
		return nil, err
	}

	// Extract the unique Leagues from the preloaded Players slice
	// collect unique leagues
	leagues := make([]*models.League, 0, len(user.Players))
	uniqueLeagues := make(map[uuid.UUID]struct{})

	for _, player := range user.Players {
		// Check if the league has already been added to avoid duplicates
		if _, ok := uniqueLeagues[player.League.ID]; !ok {
			leagues = append(leagues, player.League)
			uniqueLeagues[player.League.ID] = struct{}{}
		}
	}

	return leagues, nil
}
