package utils

import (
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"github.com/h2non/filetype"

	"github.com/mocbotau/api-join-sound/internal/models"
)

// ValidateFileUpload checks if the uploaded file meets the required criteria
func ValidateFileUpload(fileHeader *multipart.FileHeader) (string, error) {
	filename, err := validateFileMetadata(fileHeader)
	if err != nil {
		return "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("cannot open uploaded file: %v", err)
	}

	defer file.Close()

	kind, err := detectFileType(file, filename)
	if err != nil {
		return "", err
	}

	if err := checkAudioDuration(file, kind); err != nil {
		return "", err
	}

	return kind, nil
}

// validateFileMetadata validates basic file metadata, and sanities the file path
func validateFileMetadata(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > MAX_UPLOAD_SIZE {
		return "", fmt.Errorf("file too large: %s (max %d bytes)", fileHeader.Filename, MAX_UPLOAD_SIZE)
	}

	if len(fileHeader.Filename) > MAX_FILENAME_LEN {
		return "", fmt.Errorf("filename too long: %s (max %d characters)", fileHeader.Filename, MAX_FILENAME_LEN)
	}

	if fileHeader.Size == 0 {
		return "", fmt.Errorf("empty file: %s", fileHeader.Filename)
	}

	filename := strings.TrimSpace(fileHeader.Filename)
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	// sanitize filename
	return filepath.Base(filename), nil
}

// detectFileType determines the MIME type of the uploaded file by reading the actual data
func detectFileType(file multipart.File, filename string) (string, error) {
	buffer := make([]byte, 8192)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("cannot read file content: %v", err)
	}

	buffer = buffer[:n]

	kind, err := filetype.Match(buffer)
	if err != nil {
		return "", fmt.Errorf("cannot determine file type: %v", err)
	}

	if kind == filetype.Unknown {
		return "", fmt.Errorf("unknown or unsupported file type")
	}

	if !slices.Contains(slices.Collect(maps.Keys(ALLOWED_TYPES)), kind.MIME.Value) {
		return "", fmt.Errorf("unsupported file type: %s (detected: %s)", filepath.Ext(filename), kind.MIME.Value)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if expected, ok := ALLOWED_TYPES[kind.MIME.Value]; !ok || expected != ext {
		return "", fmt.Errorf("file type mismatch: %s (expected: %s, detected: %s)", ext, expected, kind.MIME.Value)
	}

	return kind.MIME.Value, nil
}

// checkAudioDuration will restrict the audio duration to a maximum limit
func checkAudioDuration(file multipart.File, kind string) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("cannot rewind file for duration check: %v", err)
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	switch kind {
	case "audio/mpeg":
		streamer, format, err = mp3.Decode(file)
	case "audio/wav":
		streamer, format, err = wav.Decode(file)
	default:
		return nil
	}

	if err != nil {
		return fmt.Errorf("cannot decode audio file: %v", err)
	}

	defer streamer.Close()

	duration := time.Duration(float64(streamer.Len())/float64(format.SampleRate)) * time.Second
	if duration > MAX_AUDIO_DURATION {
		return fmt.Errorf("audio too long: %v (max %vs)", duration, MAX_AUDIO_DURATION)
	}

	return nil
}

// GenerateInternalFilename creates a unique internal filename based on the provided ID and MIME type
func GenerateInternalFilename(id, mimeType string) string {
	return fmt.Sprintf("%s%s", id, ALLOWED_TYPES[mimeType])
}

// BuildBulkUploadResponse creates a structured response for bulk upload operations
func BuildBulkUploadResponse(totalFiles int, successfulFiles []*models.UploadResponse, failedFiles []*models.FileError) models.BulkUploadResponse {
	successCount := len(successfulFiles)
	failureCount := len(failedFiles)

	var status string
	var message string

	if successCount == totalFiles {
		status = "success"
		message = fmt.Sprintf("All %d files uploaded successfully!", totalFiles)
	} else if successCount > 0 {
		status = "partial"
		message = fmt.Sprintf("%d of %d files uploaded successfully", successCount, totalFiles)
	} else {
		status = "failure"
		message = "No files were uploaded successfully"
	}

	return models.BulkUploadResponse{
		Status:          status,
		TotalFiles:      totalFiles,
		SuccessCount:    successCount,
		FailureCount:    failureCount,
		SuccessfulFiles: successfulFiles,
		FailedFiles:     failedFiles,
		Message:         message,
	}
}
