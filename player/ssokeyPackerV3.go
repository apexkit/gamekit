package player

import (
	"context"
	"errors"
	"fmt"
	"time"

	minitoken "github.com/apexkit/gamekit/player/mini_token"
	"github.com/redis/go-redis/v9"
)

const (
	RedisKeySSOKeyV3Payload = "ssoKeyV3:%s"
	SSOKeyV3Expire          = 24 * time.Hour

	SSOKeyV3FieldKicked = "kicked"
	SSOKeyV3KickedNo     = "0"
	SSOKeyV3KickedYes    = "1"
)

var ErrSSOKeyKicked = errors.New("kicked")

func EncodedSSOKeyV3(rdb *redis.Client, params *minitoken.TokenPayload) (string, error) {
	ssoKey := generate32CharString()

	ctx := context.Background()
	key := fmt.Sprintf(RedisKeySSOKeyV3Payload, ssoKey)
	err := rdb.HSet(ctx, key, []string{
		"appId", params.AppId,
		"playerId", params.PlayerId,
		"gameBrand", params.GameBrand,
		"gameId", params.GameId,
		SSOKeyV3FieldKicked, SSOKeyV3KickedNo,
	}).Err()
	if err != nil {
		return "", err
	}
	if err = rdb.Expire(ctx, key, SSOKeyV3Expire).Err(); err != nil {
		return "", err
	}

	return ssoKey, nil
}

func DecodedSSOKeyV3(rdb *redis.Client, encodedSSOKey string) (*minitoken.TokenPayload, error) {
	ctx := context.Background()
	key := fmt.Sprintf(RedisKeySSOKeyV3Payload, encodedSSOKey)

	values, err := rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, errors.New("ssokey not found")
	}
	if values[SSOKeyV3FieldKicked] == SSOKeyV3KickedYes {
		return nil, ErrSSOKeyKicked
	}
	return &minitoken.TokenPayload{
		AppId:     values["appId"],
		PlayerId:  values["playerId"],
		GameBrand: values["gameBrand"],
		GameId:    values["gameId"],
	}, nil
}

// MarkSSOKeyV3Kicked 将旧 ssoKey 标记为已被顶号
func MarkSSOKeyV3Kicked(rdb *redis.Client, ssoKey string) error {
	if ssoKey == "" {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf(RedisKeySSOKeyV3Payload, ssoKey)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	return rdb.HSet(ctx, key, SSOKeyV3FieldKicked, SSOKeyV3KickedYes).Err()
}

// 判断一个游戏有没有多地登陆
func IsMultiLogin(tokenInfo *minitoken.TokenPayload, playerInfo *PlayerInfo) bool {
	if playerInfo.Brand != tokenInfo.GameBrand {
		return true
	}

	if playerInfo.GameID != tokenInfo.GameId {
		return true
	}
	return false
}
