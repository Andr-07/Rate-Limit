package main

import (
	"context"
	"rate-limiter/config"
	"rate-limiter/pkg/redis"
)

func main() {
	conf := configs.LoadConfig()
	ctx := context.Background()
	redis := redis.NewRedisClient(redis.RedisClientDeps{
		Config: &conf.Redis,
	})
	redis.Ping(ctx)
}
