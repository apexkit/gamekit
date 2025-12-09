package db

import (
	_ "time/tzdata" // 导入时区数据

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
		return err
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
