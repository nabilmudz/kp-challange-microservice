package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"order-service/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg *config.Config) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		MaxRetries:      5,
		MinRetryBackoff: 200 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Second,
	})

	var err error
	for i := 1; i <= 5; i++ {
		err = rdb.Ping(context.Background()).Err()
		if err == nil {
			break
		}
		log.Printf("[Redis] Ping failed, retry %d/5 in %ds", i, i)
		time.Sleep(time.Duration(i) * time.Second)
	}
	if err != nil {
		log.Fatalf("[Redis] Failed to connect after retries: %v", err)
	}

	log.Println("[Redis] Connected")
	return &RedisClient{client: rdb}
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
