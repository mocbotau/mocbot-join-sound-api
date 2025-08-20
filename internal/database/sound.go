package database

import (
	"errors"
	"fmt"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"

	"github.com/mocbotau/api-join-sound/internal/models"
)

// CreateSound creates a new sound record.
func (db *DB) CreateSound(userGuildID, originalName, internalFilename, mimeType string) (*models.Sound, error) {
	id, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	sound := models.Sound{
		ID:               id,
		UserGuildID:      userGuildID,
		OriginalName:     originalName,
		InternalFilename: internalFilename,
		MimeType:         mimeType,
		CreatedAt:        time.Now().UTC(),
	}

	if err := db.Create(&sound).Error; err != nil {
		return nil, fmt.Errorf("failed to create sound: %w", err)
	}

	return &sound, nil
}

// GetSoundByID retrieves a sound by ID with user relationship.
func (db *DB) GetSoundByID(id string) (*models.Sound, error) {
	var sound models.Sound

	err := db.Where("id = ?", id).First(&sound).Error
	if err != nil {
		return nil, err
	}

	return &sound, nil
}

// DeleteSound deletes a sound, updates the active sound if necessary, and returns the new active sound, if any.
func (db *DB) DeleteSound(id string) (deletedSound, newSound *models.Sound, err error) {
	tx := db.Begin()

	defer tx.Rollback()

	if err := tx.Where("id = ?", id).First(&deletedSound).Error; err != nil {
		return nil, nil, err
	}

	var setting models.Setting

	err = tx.Where("active_sound_id = ?", deletedSound.ID).First(&setting).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	// if found, pick a replacement or clear it
	if err == nil {
		err = tx.Where("user_guild_id = ? AND id <> ?", deletedSound.UserGuildID, deletedSound.ID).
			Order("created_at desc").
			First(&newSound).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}

		if err == nil {
			setting.ActiveSoundID = &newSound.ID
		} else {
			setting.ActiveSoundID = nil // no other sound available
		}

		if err := tx.Save(&setting).Error; err != nil {
			return nil, nil, err
		}
	}

	if err := tx.Delete(&deletedSound).Error; err != nil {
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return deletedSound, newSound, nil
}
