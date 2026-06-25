package mysql

import (
	"fmt"

	gkconf "github.com/apexkit/gamekit/conf"
	common_db "github.com/apexkit/gamekit/infra/db"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// Options holds mysql-related config from Bootstrap.data.
type Options struct {
	Data *gkconf.Data
}

// Databases holds read/write gorm handles opened from config.
type Databases struct {
	RDB *gorm.DB
	WDB *gorm.DB
}

// NewOptions builds mysql store options from shared config.
func NewOptions(data *gkconf.Data) *Options {
	return &Options{Data: data}
}

// NewDatabases opens all entries in data.database and returns read/write handles.
func NewDatabases(opt *Options, logger log.Logger) (*Databases, func(), error) {
	if opt == nil || opt.Data == nil {
		return nil, nil, fmt.Errorf("mysql: data config is nil")
	}

	mgr := common_db.NewDBManager()
	helper := log.NewHelper(logger)

	for _, v := range opt.Data.Database {
		endpoint := gkconf.RedactDSN(v.Dsn)
		if v.Type == "mysql" {
			endpoint = gkconf.MySQLEndpoint(v.Dsn)
		}
		if err := mgr.Init(v.Name, v.Type, v.Dsn); err != nil {
			mgr.Close()
			hint := gkconf.DialHint(v.Type, endpoint, err)
			if hint != "" {
				return nil, nil, fmt.Errorf("database %q (%s) at %s: %w; %s", v.Name, v.Type, endpoint, err, hint)
			}
			return nil, nil, fmt.Errorf("database %q (%s) at %s: %w", v.Name, v.Type, endpoint, err)
		}
		helper.Infof("database %q connected (%s, endpoint=%s)", v.Name, v.Type, endpoint)
	}

	wdb := mgr.GetDefaultGorm()
	if wdb == nil {
		mgr.Close()
		return nil, nil, fmt.Errorf("mysql: default database is nil — check data.database in config")
	}

	dbs := &Databases{
		RDB: mgr.GetReadGorm(),
		WDB: wdb,
	}
	cleanup := func() {
		helper.Info("closing database connections")
		mgr.Close()
	}
	return dbs, cleanup, nil
}
