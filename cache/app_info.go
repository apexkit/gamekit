package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/apexkit/gamekit/app"
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

// GetByAppID 按 app_id 获取商户应用信息（仅查询，不存在时不创建）。
func (c *AppInfoCache) GetByAppID(appID string) (*models.AppInfo, error) {
	return c.get(c.cacheKeyByAppID(appID), func() (*models.AppInfo, error) {
		return c.refreshByAppIDExisting(appID)
	})
}

// GetOrCreateByAppID 按 app_id 获取商户应用信息；AppInfo 不存在且对应 group 已存在时自动创建 AppInfo（不创建 group）。
func (c *AppInfoCache) GetOrCreateByAppID(appID string) (*models.AppInfo, error) {
	return c.get(c.cacheKeyByAppID(appID), func() (*models.AppInfo, error) {
		return c.refreshByAppIDOrCreate(appID)
	})
}

// RefreshByID 刷新单个 AppInfo 缓存（按表主键 id）
func (c *AppInfoCache) RefreshByID(id uint64) (*models.AppInfo, error) {
	return c.refreshByID(id)
}

// RefreshByAppID 刷新单个 AppInfo 缓存（按 app_id，仅查询）
func (c *AppInfoCache) RefreshByAppID(appID string) (*models.AppInfo, error) {
	return c.refreshByAppIDExisting(appID)
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

func (c *AppInfoCache) refreshByAppIDExisting(appID string) (*models.AppInfo, error) {
	var appInfo models.AppInfo
	if err := c.db.Where("app_id = ?", appID).First(&appInfo).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

func (c *AppInfoCache) refreshByAppIDOrCreate(appID string) (*models.AppInfo, error) {
	var appInfo models.AppInfo
	err := c.db.Where("app_id = ?", appID).First(&appInfo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		appGroup, err := c.resolveAppGroup(appID)
		if err != nil {
			return nil, err // group 不存在时不创建 AppInfo
		}
		appInfo = models.AppInfo{
			AppId:      appID,
			Name:       appID,
			AppGroupId: &appGroup.Id,
			Rtp:        appGroup.Rtp,
		}
		if err = c.db.Model(&models.AppInfo{}).Create(map[string]any{
			"app_id":       appID,
			"name":         appID,
			"app_group_id": appGroup.Id,
			"rtp":          appGroup.Rtp,
		}).Error; err != nil {
			if err := c.db.Where("app_id = ?", appID).First(&appInfo).Error; err != nil {
				return nil, err
			}
		} else if err := c.db.Where("app_id = ?", appID).First(&appInfo).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if err := c.setCache(appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

func (c *AppInfoCache) resolveAppGroup(appID string) (*models.AppGroupInfo, error) {
	groupID, err := app.AppGroupIDFromAppID(appID)
	if err != nil {
		return nil, err
	}
	var appGroup models.AppGroupInfo
	if err := c.db.Where("group_id = ?", groupID).First(&appGroup).Error; err != nil {
		return nil, err
	}
	return &appGroup, nil
}

func (c *AppInfoCache) cacheKeyByID(id uint64) string {
	return fmt.Sprintf("%s:id:%d", appCacheKeyPrefix, id)
}

func (c *AppInfoCache) cacheKeyByAppID(appID string) string {
	return fmt.Sprintf("%s:app_id:%s", appCacheKeyPrefix, appID)
}
