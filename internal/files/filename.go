package files
import (
	"path/filepath"
	"regexp"
	"strings"
	"uuid"
)

var (
	invalidCharsRegex = regexp.MustCompile(`[\x00-\x1f\\/:*?"<>|]`)
	trimRegex = regexp.MustCompile(`^[.\s]+|[.\s]+$`)
)

func SanitizeFilename(input string) string {
	base := filepath.Base(input)

	safe := invalidCharsRegex.ReplaceAllString(base, "_")

	safe = trimRegex.ReplaceAllString(safe, "")

	upperName := strings.ToUpper(strings.Split(safe, ".")[0])
	switch upperName {
	case "CON", "PRN", "AUX", "NUL", 
		 "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		 "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		safe = "_" + safe
	}

	if safe == "" || safe == "." || safe == ".." {
		safe = "file_" + uuid.New().String()
	}

	return safe
}