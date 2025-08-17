package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserGuildID extracts the guild ID and user ID from the request context
func GetUserGuildID(c *gin.Context) (guildID, userID int64, err error) {
	guildID, err = strconv.ParseInt(c.Param("guildId"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid guild ID: %w", err)
	}

	userID, err = strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	return guildID, userID, nil
}
