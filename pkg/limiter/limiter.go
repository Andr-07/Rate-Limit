package limiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"rate-limiter/pkg/kafka"
	"rate-limiter/pkg/redis"
)

type RateLimiter struct {
	RedisClient *redis.RedisClient
	KafkaClient *kafka.KafkaClient
	Limit       int
}

func New(redisClient *redis.RedisClient, kafkaClient *kafka.KafkaClient, limit int) *RateLimiter {
	return &RateLimiter{
		RedisClient: redisClient,
		KafkaClient: kafkaClient,
		Limit:       limit,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, userID string) bool {
	return r.RedisClient.IsRequestAllowed(ctx, userID, r.Limit)
}

func (r *RateLimiter) Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID := req.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		if !r.Allow(req.Context(), userID) {
			message := fmt.Sprintf(`{"user_id": "%s", "timestamp": "%s", "reason": "Rate limit exceeded"}`, userID, time.Now().Format(time.RFC3339))

			go func() {
				if err := r.KafkaClient.SendMessage(req.Context(), "rate-limiter-events", message); err != nil {
					fmt.Println("Error sending message to Kafka:", err)
				}
			}()

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	}
}
