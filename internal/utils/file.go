package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/h2non/filetype"
)

const MAX_UPLOAD_SIZE = 10 * 1024 * 1024 // 10 MB

func ValidateFileUpload(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > MAX_UPLOAD_SIZE {
		return "", fmt.Errorf("file too large: %s (max %d bytes)", fileHeader.Filename, MAX_UPLOAD_SIZE)
	}

	if fileHeader.Size == 0 {
		return "", fmt.Errorf("empty file: %s", fileHeader.Filename)
	}

	filename := strings.TrimSpace(fileHeader.Filename)
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}
	filename = filepath.Base(filename)

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("cannot open uploaded file: %v", err)
	}
	defer file.Close()

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

	allowedTypes := map[string]bool{
		"audio/mpeg": true, // MP3
		"audio/wav":  true, // WAV
	}

	if !allowedTypes[kind.MIME.Value] {
		return "", fmt.Errorf("unsupported file type: %s (detected: %s)", filepath.Ext(filename), kind.MIME.Value)
	}

	return kind.MIME.Value, nil
}

func GenerateInternalFilename(id, mimeType string) string {
	var ext string
	switch mimeType {
	case "audio/mpeg":
		ext = ".mp3"
	case "audio/wav":
		ext = ".wav"
	}

	return fmt.Sprintf("%s%s", id, ext)
}

func GenerateRandomInternalFilename(mimeType string) string {
	id := uuid.New()
	return GenerateInternalFilename(id.String(), mimeType)
}
