package conf

import (
	"github.com/go-kratos/kratos/v2/log"
)

var global *Bootstrap

// GetConf returns the process-wide Bootstrap set by SetConf.
func GetConf() *Bootstrap {
	return global
}

// SetConf stores Bootstrap and logs startup connection targets.
func SetConf(c *Bootstrap, logger log.Logger) {
	global = c
	logStartupTargets(c, logger)
}

func logStartupTargets(bc *Bootstrap, logger log.Logger) {
	helper := log.NewHelper(logger)
	if bc == nil || bc.Data == nil {
		return
	}
	for _, db := range bc.Data.Database {
		endpoint := RedactDSN(db.GetDsn())
		if db.GetType() == "mysql" {
			endpoint = MySQLEndpoint(db.GetDsn())
		}
		helper.Infof("startup target: database name=%s type=%s endpoint=%s", db.GetName(), db.GetType(), endpoint)
	}
	if r := bc.Data.Redis; r != nil {
		helper.Infof("startup target: redis endpoint=%s:%d db=%d tls=%v", r.GetHost(), r.GetPort(), r.GetDatabase(), r.GetUseTls())
	}
}
