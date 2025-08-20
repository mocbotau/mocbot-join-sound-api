package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/mocbotau/api-join-sound/internal/database"
	"github.com/mocbotau/api-join-sound/internal/handlers"
	"github.com/mocbotau/api-join-sound/internal/middleware"
	"github.com/mocbotau/api-join-sound/internal/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading the .env file: %v... continuing", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/main.db"
	}

	soundsFilePath := os.Getenv("SOUNDS_PATH")
	if soundsFilePath == "" {
		soundsFilePath = "./data/sounds"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	if auth0Domain := os.Getenv("AUTH0_DOMAIN"); auth0Domain == "" {
		log.Fatalf("AUTH0_DOMAIN is required")
	}

	if auth0Audience := os.Getenv("AUTH0_AUDIENCE"); auth0Audience == "" {
		log.Fatalf("AUTH0_AUDIENCE is required")
	}

	// this folder should be created by the container/kubernetes by default
	if err := os.MkdirAll(soundsFilePath, 0o750); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	defer func() {
		_ = sqlDB.Close()
	}()

	handler := handlers.NewHandler(db, soundsFilePath)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, utils.MaxPayloadSize)
		c.Next()
	})

	v1Public := r.Group("/api/v1")
	{
		v1Public.GET("/ping", handler.Ping)

		v1Public.GET("/sound/:soundId", handler.GetSound)
		v1Public.GET("/sounds/:guildId/:userId", handler.GetUserSounds)
		v1Public.GET("/settings/:guildId/:userId", handler.GetUserSettings)
	}

	v1Private := r.Group("/api/v1/", middleware.EnsureValidToken())
	{
		v1Private.DELETE("/sound/:soundId", middleware.EnsureResourceOwnership(db), handler.DeleteSound)

		v1Private.POST("/sounds/:guildId/:userId", middleware.EnsureUserAuthorization(), handler.UploadUserSounds)
		v1Private.PATCH("/settings/:guildId/:userId", middleware.EnsureUserAuthorization(), handler.UpdateUserSettings)
	}

	log.Printf("Server starting on port %s", port)

	if err := r.Run(":" + port); err != nil {
		log.Panic("Failed to start server:", err)
	}
}
