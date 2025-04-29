package redis

import (
	"context"
	"fmt"
	"log"
	"rate-limiter/config"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	Client *redis.Client
}

type RedisClientDeps struct {
	Config *configs.RedisConfig
}

func NewRedisClient(deps RedisClientDeps) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     deps.Config.Addr,
		Password: deps.Config.Password,
	})

	return &RedisClient{
		Client: client,
	}
}

func (r *RedisClient) Ping(ctx context.Context) {
	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connection to Redis:", err)
	}
	fmt.Println("Redis is connected")
}

func (r *RedisClient) IsRequestAllowed(ctx context.Context, userID string, limit int) bool {
	now := time.Now().Unix()
	windowStart := now - 60
	key := fmt.Sprintf("ratelimit:%s", userID)

	_, err := r.Client.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: now}).Result()
	if err != nil {
		fmt.Println("Error adding to Redis:", err)
		return false
	}

	_, err = r.Client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart)).Result()
	if err != nil {
		fmt.Println("Error removing old entries:", err)
		return false
	}

	count, err := r.Client.ZCard(ctx, key).Result()
	if err != nil {
		fmt.Println("Error counting requests:", err)
		return false
	}

	return int(count) <= limit
}
