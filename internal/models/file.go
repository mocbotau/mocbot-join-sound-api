package models

import "time"

type FileRecord struct {
	ID           string    `json:"id" db:"id"`
	GuildID      string    `json:"guild_id" db:"guild_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	OriginalName string    `json:"original_name" db:"original_name"`
	MimeType     string    `json:"mime_type" db:"mime_type"`
	Size         int64     `json:"size" db:"size"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type UploadResponse struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
}

type UploadRequest struct {
	GuildID string `form:"guild_id" binding:"required"`
	UserID  string `form:"user_id" binding:"required"`
}
