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

const gameCacheKeyPrefix = "cache:game"

type GameCache struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewGameCache(rdb *redis.Client, db *gorm.DB) *GameCache {
	return &GameCache{rdb: rdb, db: db}
}

// GetByPlatformGameID 按平台游戏 ID（game 表主键 id）获取游戏
func (c *GameCache) GetByPlatformGameID(platformGameID int64) (*models.Game, error) {
	return c.get(c.cacheKeyByPlatformGameID(platformGameID), func() (*models.Game, error) {
		return c.refreshByPlatformGameID(platformGameID)
	})
}

// GetByBrandAndGameID 按 game_brand + 原厂 game_id 获取游戏
func (c *GameCache) GetByBrandAndGameID(gameBrand, gameId string) (*models.Game, error) {
	return c.get(c.cacheKeyByBrandAndGameID(gameBrand, gameId), func() (*models.Game, error) {
		return c.refreshByBrandAndGameID(gameBrand, gameId)
	})
}

// RefreshByPlatformGameID 刷新单个 Game 缓存（按平台游戏 ID）
func (c *GameCache) RefreshByPlatformGameID(platformGameID int64) (*models.Game, error) {
	return c.refreshByPlatformGameID(platformGameID)
}

// RefreshByBrandAndGameID 刷新单个 Game 缓存（按 game_brand + 原厂 game_id）
func (c *GameCache) RefreshByBrandAndGameID(gameBrand, gameId string) (*models.Game, error) {
	return c.refreshByBrandAndGameID(gameBrand, gameId)
}

func (c *GameCache) setCache(game models.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, key := range []string{
		c.cacheKeyByPlatformGameID(game.PlatformGameID),
		c.cacheKeyByBrandAndGameID(game.GameBrand, game.GameId),
	} {
		if err := c.rdb.Set(ctx, key, data, 0).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *GameCache) get(cacheKey string, refresh func() (*models.Game, error)) (*models.Game, error) {
	val, err := c.rdb.Get(context.Background(), cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return refresh()
	}
	if err != nil {
		return nil, err
	}
	var game models.Game
	if err := json.Unmarshal([]byte(val), &game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (c *GameCache) refreshByPlatformGameID(platformGameID int64) (*models.Game, error) {
	var game models.Game
	if err := c.db.Where("id = ?", platformGameID).First(&game).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (c *GameCache) refreshByBrandAndGameID(gameBrand, gameId string) (*models.Game, error) {
	var game models.Game
	if err := c.db.Where("game_brand = ? AND game_id = ?", gameBrand, gameId).First(&game).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (c *GameCache) cacheKeyByPlatformGameID(platformGameID int64) string {
	return fmt.Sprintf("%s:id:%d", gameCacheKeyPrefix, platformGameID)
}

func (c *GameCache) cacheKeyByBrandAndGameID(gameBrand, gameId string) string {
	return fmt.Sprintf("%s:brand:%s:%s", gameCacheKeyPrefix, gameBrand, gameId)
}
