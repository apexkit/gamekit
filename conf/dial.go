package conf

import (
	"fmt"
	"strings"
)

// DialHint returns troubleshooting hints for common connection errors.
func DialHint(service, endpoint string, err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())

	switch service {
	case "mysql":
		switch {
		case strings.Contains(msg, "connection refused"):
			return fmt.Sprintf("hint: nothing is listening on %s — start MySQL locally or fix the host/port in config.yaml", endpoint)
		case strings.Contains(msg, "access denied"):
			return fmt.Sprintf("hint: wrong username or password for %s — check data.database[].dsn in config.yaml", endpoint)
		case strings.Contains(msg, "unknown database"):
			return fmt.Sprintf("hint: database does not exist on %s — create it or fix the db name in the DSN", endpoint)
		case strings.Contains(msg, "unexpected eof"), strings.Contains(msg, "invalid connection"), strings.Contains(msg, "bad connection"):
			return fmt.Sprintf("hint: MySQL at %s closed the connection — verify the server is running, reachable, and not killing idle connections", endpoint)
		case strings.Contains(msg, "i/o timeout"), strings.Contains(msg, "context deadline exceeded"):
			return fmt.Sprintf("hint: timed out reaching %s — check VPN, firewall, and that the address is correct", endpoint)
		default:
			return fmt.Sprintf("hint: verify MySQL is running and data.database[].dsn points to %s", endpoint)
		}
	case "redis":
		switch {
		case strings.Contains(msg, "connection refused"):
			return fmt.Sprintf("hint: nothing is listening on %s — start Redis locally or fix data.redis in config.yaml", endpoint)
		case strings.Contains(msg, "no auth"), strings.Contains(msg, "wrongpass"):
			return fmt.Sprintf("hint: wrong Redis password for %s — check data.redis.password", endpoint)
		case strings.Contains(msg, "i/o timeout"), strings.Contains(msg, "context deadline exceeded"):
			return fmt.Sprintf("hint: timed out reaching Redis at %s — check host/port, VPN/firewall, and whether data.redis.use_tls matches the server (plain TCP vs TLS)", endpoint)
		default:
			return fmt.Sprintf("hint: verify Redis is running and data.redis points to %s", endpoint)
		}
	default:
		return ""
	}
}
