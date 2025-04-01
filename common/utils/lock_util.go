package utils

import (
	"context"
	"github.com/redis/go-redis/v9"
	"synapse/common/log"
	"time"
)

type Lock struct {
	Key            string
	Token          string
	Client         redis.UniversalClient
	TimeoutSeconds int
}

func NewRedisLock(key, token string, client redis.UniversalClient, timeoutSeconds ...int) *Lock {
	timeout := 30 // 默认值
	if len(timeoutSeconds) > 0 {
		timeout = timeoutSeconds[0]
	}
	return &Lock{
		Key:            key,
		Token:          token,
		Client:         client,
		TimeoutSeconds: timeout,
	}
}

func (lock *Lock) TryLock(ctx context.Context, timeoutSecond ...int) (bool, error) {
	success, err := lock.Client.SetNX(ctx, lock.Key, lock.Token, time.Duration(lock.TimeoutSeconds)*time.Second).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (lock *Lock) Unlock(ctx context.Context) {
	err := lock.Client.Del(ctx, lock.Key).Err()
	if err != nil {
		log.Log.Errorf("failed to unlock redis lock. lockKey:%s, error: %v", lock.Key, err)
	}
}
