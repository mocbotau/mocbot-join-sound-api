package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mocbotau/api-join-sound/internal/models"
)

type DB struct {
	*sql.DB
}

func NewSQLiteDB(dataSource string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbInstance := &DB{db}

	if err := dbInstance.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return dbInstance, nil
}

func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		guild_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		original_name TEXT NOT NULL,
		mime_type TEXT NOT NULL,
		size INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_files_guild_user ON files(guild_id, user_id);
	CREATE INDEX IF NOT EXISTS idx_files_user ON files(user_id);
	`

	_, err := db.Exec(query)
	return err
}

func (db *DB) InsertFile(file *models.FileRecord) error {
	query := `
		INSERT INTO files (id, guild_id, user_id, original_name, mime_type, size, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, file.ID, file.GuildID, file.UserID, file.OriginalName,
		file.MimeType, file.Size, file.CreatedAt)
	return err
}

func (db *DB) GetFileByID(id string) (*models.FileRecord, error) {
	query := `
		SELECT id, guild_id, user_id, original_name, mime_type, size, created_at
		FROM files WHERE id = ?
	`

	var file models.FileRecord
	err := db.QueryRow(query, id).Scan(
		&file.ID, &file.GuildID, &file.UserID, &file.OriginalName,
		&file.MimeType, &file.Size, &file.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (db *DB) GetFilesByUser(guildID, userID string) ([]*models.FileRecord, error) {
	query := `
		SELECT id, guild_id, user_id, original_name, mime_type, size, created_at
		FROM files WHERE guild_id = ? AND user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, guildID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*models.FileRecord
	for rows.Next() {
		var file models.FileRecord
		err := rows.Scan(
			&file.ID, &file.GuildID, &file.UserID, &file.OriginalName,
			&file.MimeType, &file.Size, &file.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	return files, nil
}

func (db *DB) DeleteFile(id string) error {
	query := "DELETE FROM files WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}
