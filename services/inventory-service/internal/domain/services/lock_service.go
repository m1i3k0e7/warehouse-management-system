/*
    * LockService provides distributed locking functionality using Redis.
    * It allows acquiring and releasing locks to ensure that only one instance of a service can perform a specific operation at a time.
*/
package services

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type LockService struct {
    redisClient *redis.Client
}

func NewLockService(redisClient *redis.Client) *LockService {
    return &LockService{
        redisClient: redisClient,
    }
}

func (s *LockService) AcquireLock(ctx context.Context, key string, expiration time.Duration) (func(), error) {
    lockKey := fmt.Sprintf("lock:%s", key)
    lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
    
    // Try to acquire the lock using SETNX
    success, err := s.redisClient.SetNX(ctx, lockKey, lockValue, expiration).Result()
    if err != nil {
        return nil, err
    }
    
    if !success {
        return nil, fmt.Errorf("failed to acquire lock for key: %s", key)
    }
    
    // Return an unlock function that releases the lock
    unlock := func() {
        script := `
            if redis.call("get", KEYS[1]) == ARGV[1] then
                return redis.call("del", KEYS[1])
            else
                return 0
            end
        `
        s.redisClient.Eval(ctx, script, []string{lockKey}, lockValue)
    }
    
    return unlock, nil
}