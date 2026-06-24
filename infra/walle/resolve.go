package walle

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/apexkit/gamekit/infra/listen"
	"go.uber.org/zap/zapcore"
)

const (
	ConnectionInternal = 1
	ConnectionExternal = 2
)

// MySQLDSN builds a MySQL DSN from Walle mysql_config.
// IS_LOCAL=true uses external_host; otherwise internal_host.
func MySQLDSN(cfg *MySQLConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("mysql_config is nil")
	}
	host := pickHost(useExternalNetwork(), cfg.InternalHost, cfg.ExternalHost)
	if host == "" {
		return "", fmt.Errorf("mysql host is empty")
	}
	if cfg.Account == "" {
		return "", fmt.Errorf("mysql account is empty")
	}
	port := cfg.Port
	if port == 0 {
		port = 3306
	}
	dbName := strings.TrimSpace(cfg.Database)
	if dbName == "" {
		dbName = "game"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Account, cfg.Password, host, port, dbName), nil
}

// RedisEndpoint returns host:port, password, and logical DB index.
func RedisEndpoint(cfg *RedisConfig) (addr string, password string, db int, useTLS bool, err error) {
	if cfg == nil {
		return "", "", 0, false, fmt.Errorf("redis_config is nil")
	}
	useExternal := useExternalNetwork()
	endpoint := pickEndpoint(useExternal, cfg.InternalEndpoint, cfg.ExternalEndpoint)
	if endpoint == "" {
		return "", "", 0, false, fmt.Errorf("redis endpoint is empty")
	}
	addr = strings.TrimPrefix(strings.TrimPrefix(endpoint, "redis://"), "rediss://")
	if idx := strings.Index(addr, "/"); idx >= 0 {
		addr = addr[:idx]
	}
	password = cfg.Auth
	if strings.EqualFold(cfg.AuthMode, "acl") && cfg.RedisUser != "" {
		// go-redis ACL: username is set via redis.Options.Username when needed.
		_ = cfg.RedisUser
	}
	db = cfg.DB
	// TLS is only implied by an explicit rediss:// scheme. External endpoints are
	// still plain TCP unless Walle configures them that way.
	useTLS = strings.HasPrefix(strings.ToLower(endpoint), "rediss://")
	return addr, password, db, useTLS, nil
}

// RedisUsername returns ACL username when auth_mode is acl.
func RedisUsername(cfg *RedisConfig) string {
	if cfg == nil {
		return ""
	}
	if strings.EqualFold(cfg.AuthMode, "acl") {
		return cfg.RedisUser
	}
	return ""
}

// ConsulEndpoint returns address and scheme for hashicorp/consul client config.
func ConsulEndpoint(cfg *ConsulConfig) (address string, scheme string, err error) {
	if cfg == nil {
		return "", "", fmt.Errorf("consul_config is nil")
	}
	raw := pickEndpoint(useExternalNetwork(), cfg.InternalEndpoint, cfg.ExternalEndpoint)
	if raw == "" {
		return "", "", fmt.Errorf("consul endpoint is empty")
	}
	return parseServiceURL(raw)
}

func parseServiceURL(raw string) (address string, scheme string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("endpoint is empty")
	}
	if !strings.Contains(raw, "://") {
		return raw, "http", nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("parse endpoint %q: %w", raw, err)
	}
	scheme = u.Scheme
	if scheme == "" {
		scheme = "http"
	}
	address = u.Host
	if address == "" {
		address = strings.TrimPrefix(strings.TrimPrefix(u.Path, "/"), "/")
	}
	if address == "" {
		return "", "", fmt.Errorf("invalid endpoint %q", raw)
	}
	return address, scheme, nil
}

// LogLevelZap maps Walle log_level to zapcore.Level.
func LogLevelZap(level string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zapcore.DebugLevel
	case "info", "":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func useExternalNetwork() bool {
	return listen.IsLocal()
}

func pickHost(useExternal bool, internalHost, externalHost string) string {
	if useExternal {
		return strings.TrimSpace(externalHost)
	}
	return strings.TrimSpace(internalHost)
}

func pickEndpoint(useExternal bool, internalEndpoint, externalEndpoint string) string {
	if useExternal {
		return strings.TrimSpace(externalEndpoint)
	}
	return strings.TrimSpace(internalEndpoint)
}

// SplitRedisAddr splits host:port for legacy conf.Data_Redis style configs.
func SplitRedisAddr(addr string) (host string, port int32, err error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", 0, fmt.Errorf("redis addr is empty")
	}
	if strings.Contains(addr, ":") {
		host, portStr, ok := strings.Cut(addr, ":")
		if !ok {
			return "", 0, fmt.Errorf("invalid redis addr %q", addr)
		}
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("invalid redis port in %q: %w", addr, err)
		}
		return host, int32(p), nil
	}
	return addr, 6379, nil
}
