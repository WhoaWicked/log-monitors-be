package cache

import (
	"context"
	"log"
	"log-monitors/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect to redis failed: %v\n", err)
	}
	return client
}
