package conf

import (
	"fmt"
	"strings"
)

// RedactDSN masks credentials in a database DSN for safe logging.
func RedactDSN(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return ""
	}

	if strings.Contains(dsn, "://") {
		scheme, rest, ok := strings.Cut(dsn, "://")
		if !ok {
			return dsn
		}
		if at := strings.LastIndex(rest, "@"); at >= 0 {
			userInfo := rest[:at]
			hostPart := rest[at+1:]
			if colon := strings.Index(userInfo, ":"); colon >= 0 {
				userInfo = userInfo[:colon+1] + "***"
			}
			return scheme + "://" + userInfo + "@" + hostPart
		}
		return dsn
	}

	if at := strings.Index(dsn, "@"); at > 0 {
		userInfo := dsn[:at]
		rest := dsn[at:]
		if colon := strings.Index(userInfo, ":"); colon >= 0 {
			userInfo = userInfo[:colon+1] + "***"
		}
		return userInfo + rest
	}

	return dsn
}

// MySQLEndpoint extracts host/db from a MySQL DSN for logging and errors.
func MySQLEndpoint(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "unknown"
	}

	rest := dsn
	if at := strings.Index(dsn, "@"); at >= 0 {
		rest = dsn[at+1:]
	}

	rest = strings.TrimPrefix(rest, "tcp(")
	rest = strings.TrimPrefix(rest, "unix(")
	if end := strings.Index(rest, ")"); end >= 0 {
		host := rest[:end]
		rest = rest[end+1:]
		db := strings.TrimPrefix(rest, "/")
		if q := strings.Index(db, "?"); q >= 0 {
			db = db[:q]
		}
		if db == "" {
			return host
		}
		return fmt.Sprintf("%s/%s", host, db)
	}

	return RedactDSN(dsn)
}
