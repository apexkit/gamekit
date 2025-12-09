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

// GetByID 按 game 表主键 id 获取游戏
func (c *GameCache) GetByID(id int64) (*models.Game, error) {
	return c.get(c.cacheKeyByID(id), func() (*models.Game, error) {
		return c.refreshByID(id)
	})
}

// GetByBrandAndVendorGameID 按 game_brand + 厂商原始 game_id 获取游戏
func (c *GameCache) GetByBrandAndVendorGameID(gameBrand, vendorGameId string) (*models.Game, error) {
	return c.get(c.cacheKeyByBrandAndVendorGameID(gameBrand, vendorGameId), func() (*models.Game, error) {
		return c.refreshByBrandAndVendorGameID(gameBrand, vendorGameId)
	})
}

// RefreshByID 刷新单个 Game 缓存（按 game 表主键 id）
func (c *GameCache) RefreshByID(id int64) (*models.Game, error) {
	return c.refreshByID(id)
}

// RefreshByBrandAndVendorGameID 刷新单个 Game 缓存（按 game_brand + 厂商原始 game_id）
func (c *GameCache) RefreshByBrandAndVendorGameID(gameBrand, vendorGameId string) (*models.Game, error) {
	return c.refreshByBrandAndVendorGameID(gameBrand, vendorGameId)
}

func (c *GameCache) setCache(game models.Game) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, key := range []string{
		c.cacheKeyByID(game.ID),
		c.cacheKeyByBrandAndVendorGameID(game.GameBrand, game.GameId),
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

func (c *GameCache) refreshByID(id int64) (*models.Game, error) {
	var game models.Game
	if err := c.db.Where("id = ?", id).First(&game).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (c *GameCache) refreshByBrandAndVendorGameID(gameBrand, vendorGameId string) (*models.Game, error) {
	var game models.Game
	if err := c.db.Where("game_brand = ? AND game_id = ?", gameBrand, vendorGameId).First(&game).Error; err != nil {
		return nil, err
	}
	if err := c.setCache(game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (c *GameCache) cacheKeyByID(id int64) string {
	return fmt.Sprintf("%s:id:%d", gameCacheKeyPrefix, id)
}

func (c *GameCache) cacheKeyByBrandAndVendorGameID(gameBrand, vendorGameId string) string {
	return fmt.Sprintf("%s:brand:%s:%s", gameCacheKeyPrefix, gameBrand, vendorGameId)
}
