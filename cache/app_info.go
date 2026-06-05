package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/apexkit/gamekit/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const appCacheKeyPrefix = "cache:app"

type AppInfoCache struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewAppInfoCache(rdb *redis.Client, db *gorm.DB) *AppInfoCache {
	return &AppInfoCache{rdb: rdb, db: db}
}

// GetByID 按 app_info 表主键 id 获取商户应用信息
func (c *AppInfoCache) GetByID(id uint64) (*models.AppInfo, error) {
	return c.get(c.cacheKeyByID(id), func() (*models.AppInfo, error) {
		return c.refreshByID(id)
	})
}

// GetByAppID 按 app_id 获取商户应用信息
func (c *AppInfoCache) GetByAppID(appID string) (*models.AppInfo, error) {
	return c.get(c.cacheKeyByAppID(appID), func() (*models.AppInfo, error) {
		return c.refreshByAppID(appID)
	})
}

// RefreshByID 刷新单个 AppInfo 缓存（按表主键 id）
func (c *AppInfoCache) RefreshByID(id uint64) (*models.AppInfo, error) {
	return c.refreshByID(id)
}

// RefreshByAppID 刷新单个 AppInfo 缓存（按 app_id）
func (c *AppInfoCache) RefreshByAppID(appID string) (*models.AppInfo, error) {
	return c.refreshByAppID(appID)
}

func (c *AppInfoCache) setCache(app models.AppInfo) error {
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	ctx := context.Background()
	keys := []string{
		c.cacheKeyByID(app.Id),
		c.cacheKeyByAppID(app.AppId),
	}
	for _, key := range keys {
		if err := c.rdb.Set(ctx, key, data, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *AppInfoCache) get(cacheKey string, refresh func() (*models.AppInfo, error)) (*models.AppInfo, error) {
	val, err := c.rdb.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return refresh()
	}
	if err != nil {
		return nil, err
	}
	var app models.AppInfo
	if err := json.Unmarshal([]byte(val), &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *AppInfoCache) refreshByID(id uint64) (*models.AppInfo, error) {
	var app models.AppInfo
	if err := c.db.Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *AppInfoCache) refreshByAppID(appID string) (*models.AppInfo, error) {
	var app models.AppInfo
	if err := c.db.Where("app_id = ?", appID).First(&app).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *AppInfoCache) cacheKeyByID(id uint64) string {
	return fmt.Sprintf("%s:id:%d", appCacheKeyPrefix, id)
}

func (c *AppInfoCache) cacheKeyByAppID(appID string) string {
	return fmt.Sprintf("%s:app_id:%s", appCacheKeyPrefix, appID)
}
