package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/mocbotau/api-join-sound/internal/models"
)

// UpdateUserSetting updates user settings.
func (db *DB) UpdateUserSetting(userID string, req *models.UpdateSettingsRequest) (*models.Setting, error) {
	setting, err := db.getCreateSettings(userID)
	if err != nil {
		return nil, err
	}

	if req.ActiveSoundID != nil && *req.ActiveSoundID != "" {
		setting.ActiveSoundID = req.ActiveSoundID
	}

	if req.Mode != nil && *req.Mode != "" {
		setting.Mode = *req.Mode
	}

	if err := db.Save(setting).Error; err != nil {
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	return setting, nil
}

// GetOrCreateUserSetting retrieves or creates user settings if they don't already exist.
func (db *DB) GetOrCreateUserSetting(guildID, userID int64) (*models.Setting, error) {
	user, err := db.CreateOrGetUser(guildID, userID)
	if err != nil {
		return nil, err
	}

	setting, err := db.getCreateSettings(user.ID)
	if err != nil {
		return nil, err
	}

	err = db.
		Where("user_guild_id = ?", user.ID).
		First(setting).Error
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (db *DB) getCreateSettings(userGuildID string) (*models.Setting, error) {
	var setting *models.Setting

	err := db.Where("user_guild_id = ?", userGuildID).First(&setting).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newSetting := &models.Setting{
			UserGuildID: userGuildID,
		}

		if err := db.Create(newSetting).Error; err != nil {
			return nil, fmt.Errorf("failed to create setting: %w", err)
		}

		setting = newSetting
	} else if err != nil {
		return nil, err
	}

	return setting, nil
}
