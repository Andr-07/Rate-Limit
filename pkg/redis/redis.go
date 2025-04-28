package redis

import (
	"context"
	"fmt"
	"log"
	"rate-limiter/config"

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
		Addr:    deps.Config.Addr,
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


