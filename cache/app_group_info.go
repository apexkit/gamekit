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
	return c.GetByAppGroupID(appGroupID)
}

// GetByAppGroupID 按 group_id 获取总商户信息
func (c *AppGroupInfoCache) GetByAppGroupID(appGroupID string) (*models.AppGroupInfo, error) {
	return c.get(c.cacheKeyByAppGroupID(appGroupID), func() (*models.AppGroupInfo, error) {
		return c.refreshByAppGroupID(appGroupID)
	})
}

// RefreshByID 刷新单个 AppGroupInfo 缓存（按表主键 id）
func (c *AppGroupInfoCache) RefreshByID(id uint64) (*models.AppGroupInfo, error) {
	return c.refreshByID(id)
}

// RefreshByAppGroupID 刷新单个 AppGroupInfo 缓存（按 group_id）
func (c *AppGroupInfoCache) RefreshByAppGroupID(appGroupID string) (*models.AppGroupInfo, error) {
	return c.refreshByAppGroupID(appGroupID)
}

// EvictByGroupID removes cached app group entries for group_id (and id when provided).
func (c *AppGroupInfoCache) EvictByGroupID(groupID string, id uint64) error {
	ctx := context.Background()
	keys := []string{c.cacheKeyByAppGroupID(groupID)}
	if id > 0 {
		keys = append(keys, c.cacheKeyByID(id))
	}
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *AppGroupInfoCache) setCache(agi models.AppGroupInfo) error {
	data, err := json.Marshal(agi)
	if err != nil {
		return err
	}
	ctx := context.Background()
	keys := []string{
		c.cacheKeyByID(agi.Id),
		c.cacheKeyByAppGroupID(agi.GroupId),
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

func (c *AppGroupInfoCache) refreshByAppGroupID(appGroupID string) (*models.AppGroupInfo, error) {
	var agi models.AppGroupInfo
	if err := c.db.Where("group_id = ?", appGroupID).First(&agi).Error; err != nil {
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

func (c *AppGroupInfoCache) cacheKeyByAppGroupID(appGroupID string) string {
	return fmt.Sprintf("%s:group_id:%s", appGroupInfoCacheKeyPrefix, appGroupID)
}
