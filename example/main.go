package main

import (
	"context"
	"fmt"
	"github.com/Andr-07/rate-limiter/config"
	"github.com/Andr-07/rate-limiter/pkg/kafka"
	"github.com/Andr-07/rate-limiter/pkg/limiter"
	"github.com/Andr-07/rate-limiter/pkg/redis"
	"net/http"
	"time"
)

func main() {
	conf := configs.LoadConfig()
	ctx := context.Background()

	redis := redis.NewRedisClient(redis.RedisClientDeps{Config: &conf.Redis})
	redis.Ping(ctx)

	kafka, err := kafka.NewKafkaClient(kafka.KafkaClientDeps{Config: &conf.Kafka})
	if err != nil {
		fmt.Println("Failed to create Kafka client:", err)
		return
	}
	defer kafka.Close()

	// Example
	rl := limiter.New(redis, kafka, &limiter.RateLimiterConfig{
		MaxRequestsPerUser: 10,
		MaxRequestsPerIP:   10,
		TimeWindow:         60 * time.Second,
		BlockDuration:      5 * time.Minute,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.Handle("/api", rl.Middleware(handler))
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
