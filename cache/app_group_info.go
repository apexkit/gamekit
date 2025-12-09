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

const appGroupInfoCacheKeyPrefix = "cache:app_group"

type AppGroupInfoCache struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewAppGroupInfoCache(rdb *redis.Client, db *gorm.DB) *AppGroupInfoCache {
	return &AppGroupInfoCache{rdb: rdb, db: db}
}

// GetByID 按 app_group_info 表主键 id 获取总商户信息
func (c *AppGroupInfoCache) GetByAppID(appID string) (*models.AppGroupInfo, error) {
	appGroupID, err := app.AppGroupIDFromAppID(appID)
	if err != nil {
		return nil, err
	}
	return c.GetByGroupID(appGroupID)
}

// GetByGroupID 按 group_id 获取总商户信息
func (c *AppGroupInfoCache) GetByGroupID(groupID string) (*models.AppGroupInfo, error) {
	return c.get(c.cacheKeyByGroupID(groupID), func() (*models.AppGroupInfo, error) {
		return c.refreshByGroupID(groupID)
	})
}

// GetByAccessKey 按 access_key 获取总商户信息
func (c *AppGroupInfoCache) GetByAccessKey(accessKey string) (*models.AppGroupInfo, error) {
	return c.get(c.cacheKeyByAccessKey(accessKey), func() (*models.AppGroupInfo, error) {
		return c.refreshByAccessKey(accessKey)
	})
}

// RefreshByID 刷新单个 AppGroupInfo 缓存（按表主键 id）
func (c *AppGroupInfoCache) RefreshByID(id uint64) (*models.AppGroupInfo, error) {
	return c.refreshByID(id)
}

// RefreshByGroupID 刷新单个 AppGroupInfo 缓存（按 group_id）
func (c *AppGroupInfoCache) RefreshByGroupID(groupID string) (*models.AppGroupInfo, error) {
	return c.refreshByGroupID(groupID)
}

// RefreshByAccessKey 刷新单个 AppGroupInfo 缓存（按 access_key）
func (c *AppGroupInfoCache) RefreshByAccessKey(accessKey string) (*models.AppGroupInfo, error) {
	return c.refreshByAccessKey(accessKey)
}

func (c *AppGroupInfoCache) setCache(agi models.AppGroupInfo) error {
	data, err := json.Marshal(agi)
	if err != nil {
		return err
	}
	ctx := context.Background()
	keys := []string{
		c.cacheKeyByID(agi.Id),
		c.cacheKeyByGroupID(agi.GroupId),
	}
	if agi.AccessKey != "" {
		keys = append(keys, c.cacheKeyByAccessKey(agi.AccessKey))
	}
	for _, key := range keys {
		if err := c.rdb.Set(ctx, key, data, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *AppGroupInfoCache) get(cacheKey string, refresh func() (*models.AppGroupInfo, error)) (*models.AppGroupInfo, error) {
	val, err := c.rdb.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return refresh()
	}
	if err != nil {
		return nil, err
	}
	var agi models.AppGroupInfo
	if err := json.Unmarshal([]byte(val), &agi); err != nil {
		return nil, err
	}
	return &agi, nil
}

func (c *AppGroupInfoCache) refreshByID(id uint64) (*models.AppGroupInfo, error) {
	var agi models.AppGroupInfo
	if err := c.db.Where("id = ?", id).First(&agi).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(agi); err != nil {
		return nil, err
	}
	return &agi, nil
}

func (c *AppGroupInfoCache) refreshByGroupID(groupID string) (*models.AppGroupInfo, error) {
	var agi models.AppGroupInfo
	if err := c.db.Where("group_id = ?", groupID).First(&agi).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(agi); err != nil {
		return nil, err
	}
	return &agi, nil
}

func (c *AppGroupInfoCache) refreshByAccessKey(accessKey string) (*models.AppGroupInfo, error) {
	var agi models.AppGroupInfo
	if err := c.db.Where("access_key = ?", accessKey).First(&agi).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(agi); err != nil {
		return nil, err
	}
	return &agi, nil
}

func (c *AppGroupInfoCache) cacheKeyByID(id uint64) string {
	return fmt.Sprintf("%s:id:%d", appGroupInfoCacheKeyPrefix, id)
}

func (c *AppGroupInfoCache) cacheKeyByGroupID(groupID string) string {
	return fmt.Sprintf("%s:group_id:%s", appGroupInfoCacheKeyPrefix, groupID)
}

func (c *AppGroupInfoCache) cacheKeyByAccessKey(accessKey string) string {
	return fmt.Sprintf("%s:access_key:%s", appGroupInfoCacheKeyPrefix, accessKey)
}
