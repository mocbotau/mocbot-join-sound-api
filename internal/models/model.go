package models

import (
	"mime/multipart"
	"time"
)

// User represents a user in a specific guild.
type User struct {
	ID      string `json:"id" gorm:"type:text;primaryKey;not null"`
	GuildID int64  `json:"guild_id" gorm:"not null;index"`
	UserID  int64  `json:"user_id" gorm:"not null;index"`

	Sounds   []Sound  `json:"-" gorm:"foreignKey:UserGuildID"`
	Settings *Setting `json:"-" gorm:"foreignKey:UserGuildID"`
}

// Sound represents an uploaded sound file.
type Sound struct {
	ID               string    `json:"id" gorm:"type:text;primaryKey;not null"`
	UserGuildID      string    `json:"user_guild_id" gorm:"type:text;not null;index"`
	OriginalName     string    `json:"original_name" gorm:"type:text;not null"`
	InternalFilename string    `json:"-" gorm:"type:text;not null"` // we don't want to expose this to the user
	MimeType         string    `json:"mime_type" gorm:"type:text;not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`

	User     User      `json:"-" gorm:"foreignKey:UserGuildID;references:ID"`
	Settings []Setting `json:"-" gorm:"foreignKey:ActiveSoundID"`
}

// Setting represents user settings for sound playback.
type Setting struct {
	UserGuildID   string  `json:"user_guild_id" gorm:"type:text;not null;primaryKey"`
	ActiveSoundID *string `json:"active_sound_id" gorm:"type:text;index"`
	Mode          string  `json:"mode" gorm:"type:text;not null;default:'single';check:mode IN ('single', 'random')"`

	User        User  `json:"-" gorm:"foreignKey:UserGuildID;references:ID"`
	ActiveSound Sound `json:"-" gorm:"foreignKey:ActiveSoundID;references:ID"`
}

// UploadRequest represents a request to upload files.
type UploadRequest struct {
	Files []*multipart.FileHeader `form:"files"`
}

// UploadResponse represents a response after uploading files.
type UploadResponse struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
}

// BulkUploadResponse represents a response after bulk uploading files.
type BulkUploadResponse struct {
	Status          string            `json:"status"` // "success", "partial", "failure"
	TotalFiles      int               `json:"total_files"`
	SuccessCount    int               `json:"success_count"`
	FailureCount    int               `json:"failure_count"`
	SuccessfulFiles []*UploadResponse `json:"successful_files"`
	FailedFiles     []*FileError      `json:"failed_files"`
	Message         string            `json:"message"`
}

// FileError represents an error that occurred during file upload.
type FileError struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
	Index    int    `json:"index"`
}

// UpdateSettingsRequest represents a request to update user settings.
type UpdateSettingsRequest struct {
	ActiveSoundID *string `json:"active_sound_id"`
	Mode          *string `json:"mode"`
}
