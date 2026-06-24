package listen

import (
	"os"
	"strings"
)

// IsLocal reports local dev mode (IS_LOCAL=true|1|yes|on).
func IsLocal() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("IS_LOCAL"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
