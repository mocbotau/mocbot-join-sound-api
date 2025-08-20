package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mocbotau/api-join-sound/internal/database"
)

// Handler is the HTTP handler for the API.
type Handler struct {
	db             *database.DB
	soundsFilePath string
}

// NewHandler creates a new Handler instance.
func NewHandler(db *database.DB, soundsFilePath string) *Handler {
	return &Handler{db: db, soundsFilePath: soundsFilePath}
}

// Ping responds with a pong message.
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
