package utils

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type DistributedLock struct {
	rs *redsync.Redsync //分布式锁
}

func NewDistributedLock(redisClient *redis.Client) *DistributedLock {

	// 创建redsync的客户端连接池
	pool := goredis.NewPool(redisClient)

	// 创建redsync实例
	rs := redsync.New(pool)

	return &DistributedLock{rs: rs}
}

// =====================================================================================
// 分布式锁的操作
func (sc *DistributedLock) Lock(key string) (*redsync.Mutex, error) {
	// 创建一个互斥锁
	mutex := sc.rs.NewMutex(key)
	if err := mutex.Lock(); err != nil {
		//panic(err)
	}
	return mutex, nil
}

func (sc *DistributedLock) Unlock(mutex *redsync.Mutex) error {
	// 释放锁
	if _, err := mutex.Unlock(); err != nil {
		//panic(err)
	}
	return nil
}

// =====================================================================================
