/*
    * Cache Service provides caching capabilities for inventory data using Redis, allowing for quick access to frequently used data.
    * It supports setting, getting, and deleting cache entries, as well as specific methods for managing shelf statuses.
*/
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
    "WMS/services/inventory-service/internal/domain/entities"
)

type CacheService struct {
    client *redis.Client
}

func NewCacheService(client *redis.Client) *CacheService {
    return &CacheService{client: client}
}

func (s *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return s.client.Set(ctx, key, data, expiration).Err()
}

func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := s.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
    return s.client.Del(ctx, key).Err()
}

func (s *CacheService) GetShelfStatus(ctx context.Context, shelfID string) (*entities.ShelfStatus, error) {
    key := fmt.Sprintf("shelf_status:%s", shelfID)
    var status entities.ShelfStatus
    
    if err := s.Get(ctx, key, &status); err != nil {
        return nil, err
    }
    
    return &status, nil
}

func (s *CacheService) SetShelfStatus(ctx context.Context, status *entities.ShelfStatus) error {
    key := fmt.Sprintf("shelf_status:%s", status.ShelfID)
    return s.Set(ctx, key, status, time.Minute*10)
}