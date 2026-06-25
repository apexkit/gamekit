package walle

import (
	"os"
	"strings"

	"github.com/apexkit/gamekit/infra/listen"
)

const EnvGroup = "GROUP"

// SkipConsulRegister reports whether the service should skip Consul registration.
// Default: IS_LOCAL=true and GROUP is set (local debug against Walle remote resources).
func SkipConsulRegister() bool {
	return listen.IsLocal() && strings.TrimSpace(os.Getenv(EnvGroup)) != ""
}
