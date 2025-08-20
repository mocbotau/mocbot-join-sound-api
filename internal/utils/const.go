package utils

import "time"

// MaxAudioDuration is the maximum duration for audio files that we permit users to upload.
const MaxAudioDuration = 5 * time.Second

const (
	// MaxPayloadSize is the maximum size for request payloads.
	MaxPayloadSize = 50 << 20 // 50 MB
	// MaxUploadSize is the maximum size for individual file uploads.
	MaxUploadSize = 10 * 1024 * 1024 // 10 MB
)

const (
	// MaxFilesPerUser is the maximum number of files a user can upload.
	MaxFilesPerUser = 5
	// MaxFilenameLen is the maximum length of uploaded file names.
	MaxFilenameLen = 255
)

// AllowedTypes is a map of allowed audio file types and their corresponding extensions.
var AllowedTypes = map[string]string{
	"audio/mpeg":  ".mp3",
	"audio/x-wav": ".wav",
}

// AllowedModes is a list of allowed playback modes.
var AllowedModes = []string{"single", "random"}
