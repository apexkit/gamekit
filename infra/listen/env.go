package listen

import (
	"os"
	"strings"
)

const EnvLocal = "IS_LOCAL"

// IsLocal reports local dev mode when IS_LOCAL is true, 1, yes, or on.
// Used for listen addresses, Walle external endpoints, dev logging, CORS, etc.
func IsLocal() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvLocal))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
