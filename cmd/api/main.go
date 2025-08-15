package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/mocbotau/api-join-sound/internal/database"
	"github.com/mocbotau/api-join-sound/internal/handlers"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/sounds.db"
	}

	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	if err := os.MkdirAll("./files", 0755); err != nil {
		log.Fatal("Failed to create files directory:", err)
	}

	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	handler := handlers.NewHandler(db)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", handler.Ping)
		v1.POST("/upload", handler.UploadFiles)
		v1.GET("/file/:fileId", handler.GetFile)
		v1.GET("/files/:guildId/:userId", handler.GetUserFiles)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
