package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mocbotau/api-join-sound/internal/database"
	"github.com/mocbotau/api-join-sound/internal/models"
	"github.com/mocbotau/api-join-sound/internal/utils"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (h *Handler) UploadFiles(c *gin.Context) {
	c.Request.ParseMultipartForm(50 << 20)

	var uploadReq models.UploadRequest
	if err := c.ShouldBind(&uploadReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form data is improperly formatted"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}

	var uploadedFiles []models.UploadResponse
	for _, file := range files {
		mimeType, err := utils.ValidateFileUpload(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fileID := uuid.New().String()
		internalFilename := utils.GenerateInternalFilename(fileID, mimeType)

		if err := c.SaveUploadedFile(file, "./files/"+internalFilename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		fileRecord := &models.FileRecord{
			ID:           fileID,
			GuildID:      uploadReq.GuildID,
			UserID:       uploadReq.UserID,
			OriginalName: file.Filename,
			MimeType:     mimeType,
			Size:         file.Size,
			CreatedAt:    time.Now(),
		}

		if err := h.db.InsertFile(fileRecord); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file record"})
			return
		}

		uploadedFiles = append(uploadedFiles, models.UploadResponse{
			ID:           fileID,
			OriginalName: file.Filename,
			Size:         file.Size,
			MimeType:     mimeType,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%d files uploaded successfully!", len(uploadedFiles)),
		"files":   uploadedFiles,
	})
}

func (h *Handler) GetFile(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	record, err := h.db.GetFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	internalFilename := utils.GenerateInternalFilename(record.ID, record.MimeType)

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", record.OriginalName))
	c.File(fmt.Sprintf("./files/%s", internalFilename))
}

func (h *Handler) GetUserFiles(c *gin.Context) {
	guildID := c.Param("guildId")
	userID := c.Param("userId")

	if guildID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guild ID and User ID are required"})
		return
	}

	files, err := h.db.GetFilesByUser(guildID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	var response []gin.H
	for _, file := range files {
		response = append(response, gin.H{
			"id":            file.ID,
			"original_name": file.OriginalName,
			"size":          file.Size,
			"mime_type":     file.MimeType,
			"uploaded_at":   file.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"files": response,
	})
}
