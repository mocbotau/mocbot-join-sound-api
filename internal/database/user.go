package database

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"

	"github.com/mocbotau/api-join-sound/internal/models"
)

// CreateOrGetUser creates a new user or returns existing one
func (db *DB) CreateOrGetUser(guildID, userID int64) (*models.User, error) {
	var user models.User

	id, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	err = db.Where("guild_id = ? AND user_id = ?", guildID, userID).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		user = models.User{
			ID:      id,
			GuildID: guildID,
			UserID:  userID,
		}

		if err := db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUserByUserGuildID retrieves a user by their userGuildID
func (db *DB) GetUserByUserGuildID(id string) (*models.User, error) {
	var user models.User
	err := db.Where("id = ?", id).First(&user).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetSoundsByUser retrieves all sounds for a specific user
func (db *DB) GetSoundsByUser(guildID, userID int64) ([]*models.Sound, error) {
	user, err := db.CreateOrGetUser(guildID, userID)
	if err != nil {
		return nil, err
	}

	var sounds []*models.Sound
	err = db.Where("user_guild_id = ?", user.ID).
		Order("created_at DESC").
		Find(&sounds).Error

	if err != nil {
		return nil, err
	}

	return sounds, nil
}
