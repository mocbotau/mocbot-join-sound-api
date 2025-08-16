package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocbotau/api-join-sound/internal/models"
	"github.com/mocbotau/api-join-sound/internal/utils"
)

// GetUserSettings returns user settings with active sound details
func (h *Handler) GetUserSettings(c *gin.Context) {
	guildID, userID, err := utils.GetUserGuildID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.db.GetOrCreateUserSetting(guildID, userID)
	if err != nil {
		// should always automatically create settings if they don't exist
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"setting": setting,
	})
}

// UpdateUserSettings updates user settings
func (h *Handler) UpdateUserSettings(c *gin.Context) {
	guildID, userID, err := utils.GetUserGuildID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ActiveSoundID == nil && req.Mode == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// should always create user if doesn't exist
	user, err := h.db.CreateOrGetUser(guildID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	var sound *models.Sound

	if req.ActiveSoundID != nil {
		sound, err = h.db.GetSoundByID(*req.ActiveSoundID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Sound not found"})
			return
		}

		if sound.UserGuildID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Sound does not belong to this user"})
			return
		}
	}

	setting, err := h.db.UpdateUserSetting(user.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"setting": setting,
	})
}
