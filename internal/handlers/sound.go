package handlers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"

	"github.com/mocbotau/api-join-sound/internal/database"
	"github.com/mocbotau/api-join-sound/internal/models"
	"github.com/mocbotau/api-join-sound/internal/utils"
)

type fileUploader struct {
	currentSoundCount int
	db                *database.DB
	failedFiles       []*models.FileError
	soundsFilePath    string
	successFiles      []*models.UploadResponse
	user              *models.User
}

// GetSound retrieves a sound by its global ID.
func (h *Handler) GetSound(c *gin.Context) {
	soundID := c.Param("soundId")
	if soundID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sound ID is required"})
		return
	}

	sound, err := h.db.GetSoundByID(soundID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sound not found"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", sound.OriginalName))
	c.Header("Content-Type", sound.MimeType)
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(fmt.Sprintf("%s/%s", h.soundsFilePath, sound.InternalFilename))
}

// DeleteSound deletes a sound given its global ID.
func (h *Handler) DeleteSound(c *gin.Context) {
	soundID := c.Param("soundId")
	if soundID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sound ID is required"})
		return
	}

	deletedSound, newSound, err := h.db.DeleteSound(soundID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sound not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sound"})
		return
	}

	err = os.Remove(fmt.Sprintf("%s/%s", h.soundsFilePath, deletedSound.InternalFilename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sound file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Sound deleted successfully",
		"deleted_sound": deletedSound,
		"new_sound":     newSound,
	})
}

// GetUserSounds retrieves all sounds for a given user in a given guild.
func (h *Handler) GetUserSounds(c *gin.Context) {
	guildID, userID, err := utils.GetUserGuildID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sounds, err := h.db.GetSoundsByUser(guildID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sounds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sounds": sounds,
	})
}

// UploadUserSounds handles the upload of sounds for a given user in a given guild.
func (h *Handler) UploadUserSounds(c *gin.Context) {
	guildID, userID, err := utils.GetUserGuildID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	if len(files) > utils.MaxFilesPerUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Can't upload more than %d files at once", utils.MaxFilesPerUser)})
		return
	}

	user, err := h.db.CreateOrGetUser(guildID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	currentSounds, err := h.db.GetSoundsByUser(guildID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch current sounds"})
		return
	}

	fileUploader := &fileUploader{
		currentSoundCount: len(currentSounds),
		db:                h.db,
		failedFiles:       make([]*models.FileError, 0),
		soundsFilePath:    h.soundsFilePath,
		successFiles:      make([]*models.UploadResponse, 0),
		user:              user,
	}

	response := fileUploader.processBulkUpload(c, files)

	if response.SuccessCount > 0 {
		if response.FailureCount > 0 {
			c.JSON(http.StatusMultiStatus, response)
		} else {
			c.JSON(http.StatusOK, response)
		}
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

func (fu *fileUploader) processBulkUpload(c *gin.Context, files []*multipart.FileHeader) models.BulkUploadResponse {
	for i, file := range files {
		if uploadResponse, err := fu.uploadSingleFile(c, file); err != nil {
			fu.failedFiles = append(fu.failedFiles, &models.FileError{
				Filename: file.Filename,
				Error:    err.Error(),
				Index:    i,
			})
		} else {
			fu.successFiles = append(fu.successFiles, &uploadResponse)
			fu.currentSoundCount++
		}
	}

	return utils.BuildBulkUploadResponse(len(files), fu.successFiles, fu.failedFiles)
}

func (fu *fileUploader) uploadSingleFile(c *gin.Context, file *multipart.FileHeader) (models.UploadResponse, error) {
	if fu.currentSoundCount >= utils.MaxFilesPerUser {
		return models.UploadResponse{}, fmt.Errorf("maximum file limit of %d reached per user", utils.MaxFilesPerUser)
	}

	mimeType, err := utils.ValidateFileUpload(file)
	if err != nil {
		return models.UploadResponse{}, fmt.Errorf("file validation failed: %w", err)
	}

	fileID, err := gonanoid.New()
	if err != nil {
		return models.UploadResponse{}, fmt.Errorf("failed to generate file ID: %w", err)
	}

	internalFilename := utils.GenerateInternalFilename(fileID, mimeType)

	filePath := fmt.Sprintf("%s/%s", fu.soundsFilePath, internalFilename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return models.UploadResponse{}, fmt.Errorf("failed to save file: %w", err)
	}

	sound, err := fu.db.CreateSound(fu.user.ID, file.Filename, internalFilename, mimeType)
	if err != nil {
		removeErr := os.Remove(filePath)
		if removeErr != nil {
			return models.UploadResponse{}, fmt.Errorf("failed to remove uploaded file: %w", removeErr)
		}

		return models.UploadResponse{}, fmt.Errorf("failed to store file record: %w", err)
	}

	return models.UploadResponse{
		ID:           sound.ID,
		OriginalName: file.Filename,
		Size:         file.Size,
		MimeType:     mimeType,
	}, nil
}
