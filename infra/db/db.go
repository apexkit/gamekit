package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	_ "time/tzdata" // 导入时区数据

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/apexkit/gamekit/infra/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

type DBManager struct {
	dbMap    map[string]*gorm.DB      //关系型数据库的操作
	redisMap map[string]*redis.Client //redis数据库的操作
}

func NewDBManager() *DBManager {
	return &DBManager{
		dbMap:    make(map[string]*gorm.DB),
		redisMap: make(map[string]*redis.Client),
	}
}

func (dbManager *DBManager) Init(name string, dbType string, dsn string) error {
	logger := zapgorm2.New(zap.L())
	logger.SetAsDefault()

	var dialector gorm.Dialector = nil

	switch dbType {
	case "mysql":
		mysqldriver.SetLogger(mysqlDriverLogger{})
		dialector = mysql.Open(dsn)
	case "pgsql":
		dialector = postgres.Open(dsn)
	case "redis":
		client, err := utils.InitRedisByDNS(dsn)
		if err != nil {
			return err
		}
		dbManager.redisMap[name] = client
		return nil
	default:
		panic("unsuport sql drive")
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger})

	if err != nil {
		return fmt.Errorf("open gorm (%s): %w", redactDSN(dsn), err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql handle for %q: %w", name, err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping %s database %q at %s: %w", dbType, name, endpointFromDSN(dbType, dsn), err)
	}

	dbManager.dbMap[name] = db
	return nil
}

func (dbManager *DBManager) GetGorm(name string) *gorm.DB {
	if _, ok := dbManager.dbMap[name]; ok {
		return dbManager.dbMap[name]
	}
	return nil
}

func (dbManager *DBManager) GetDefaultGorm() *gorm.DB {
	return dbManager.GetGorm("default")
}

// GetReadGorm returns read_only when configured, otherwise default.
func (dbManager *DBManager) GetReadGorm() *gorm.DB {
	if db := dbManager.GetGorm("read_only"); db != nil {
		return db
	}
	return dbManager.GetGorm("default")
}

func (dbManager *DBManager) GetRedisClient(name string) *redis.Client {
	if _, ok := dbManager.redisMap[name]; ok {
		return dbManager.redisMap[name]
	}
	return nil
}

func (dbManager *DBManager) GetDefaultRedis() *redis.Client {
	return dbManager.GetRedisClient("default")
}

func (dbManager *DBManager) Close() {
	for _, db := range dbManager.dbMap {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
}

type mysqlDriverLogger struct{}

func (mysqlDriverLogger) Print(v ...interface{}) {
	zap.L().Warn(fmt.Sprint(v...), zap.String("component", "mysql/driver"))
}

func redactDSN(dsn string) string {
	dsn = strings.TrimSpace(dsn)
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

func endpointFromDSN(dbType, dsn string) string {
	if dbType != "mysql" {
		return redactDSN(dsn)
	}
	rest := dsn
	if at := strings.Index(dsn, "@"); at >= 0 {
		rest = dsn[at+1:]
	}
	rest = strings.TrimPrefix(rest, "tcp(")
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
		return host + "/" + db
	}
	return redactDSN(dsn)
}
