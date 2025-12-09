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

const appGroupGameCacheKeyPrefix = "cache:app_group_game"

type AppGroupGameCache struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewAppGroupGameCache(rdb *redis.Client, db *gorm.DB) *AppGroupGameCache {
	return &AppGroupGameCache{rdb: rdb, db: db}
}

// GetByAppIDAndGameID 按 app_id + game 表主键 id 获取总商户游戏配置
func (c *AppGroupGameCache) GetByAppIDAndGameID(appID string, gameID int64) (*models.AppGroupGame, error) {
	appGroupID, err := app.AppGroupIDFromAppID(appID)
	if err != nil {
		return nil, err
	}
	return c.get(c.cacheKey(appGroupID, gameID), func() (*models.AppGroupGame, error) {
		return c.refreshByAppGroupAndGameID(appGroupID, gameID)
	})
}

// RefreshByAppIDAndGameID 刷新单个 AppGroupGame 缓存
func (c *AppGroupGameCache) RefreshByAppIDAndGameID(appID string, gameID int64) (*models.AppGroupGame, error) {
	appGroupID, err := app.AppGroupIDFromAppID(appID)
	if err != nil {
		return nil, err
	}
	return c.refreshByAppGroupAndGameID(appGroupID, gameID)
}

func (c *AppGroupGameCache) get(cacheKey string, refresh func() (*models.AppGroupGame, error)) (*models.AppGroupGame, error) {
	val, err := c.rdb.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return refresh()
	}
	if err != nil {
		return nil, err
	}
	var agg models.AppGroupGame
	if err := json.Unmarshal([]byte(val), &agg); err != nil {
		return nil, err
	}
	return &agg, nil
}

func (c *AppGroupGameCache) refreshByAppGroupAndGameID(appGroupID string, gameID int64) (*models.AppGroupGame, error) {
	var agg models.AppGroupGame
	if err := c.db.Where("app_group_id = ? AND game_id = ?", appGroupID, gameID).First(&agg).Error; err != nil {
		return nil, err
	}
	data, err := json.Marshal(agg)
	if err != nil {
		return nil, err
	}
	if err := c.rdb.Set(context.Background(), c.cacheKey(appGroupID, gameID), data, 0).Err(); err != nil {
		return nil, err
	}
	return &agg, nil
}

func (c *AppGroupGameCache) cacheKey(appGroupID string, gameID int64) string {
	return fmt.Sprintf("%s:%s:%d", appGroupGameCacheKeyPrefix, appGroupID, gameID)
}
