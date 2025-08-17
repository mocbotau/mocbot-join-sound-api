package utils

const MAX_AUDIO_DURATION = 5

const MAX_PAYLOAD_SIZE = 50 << 20        // 50 MB
const MAX_UPLOAD_SIZE = 10 * 1024 * 1024 // 10 MB

const MAX_FILES_PER_USER = 5
const MAX_FILENAME_LEN = 255

var ALLOWED_TYPES = map[string]string{
	"audio/mpeg": ".mp3",
	"audio/wav":  ".wav",
}

var ALLOWED_MODES = []string{"single", "random"}
