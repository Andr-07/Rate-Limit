package limiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ictx "github.com/Andr-07/rate-limiter/internal/context"
	"github.com/Andr-07/rate-limiter/internal/utils"
	"github.com/Andr-07/rate-limiter/pkg/kafka"
	"github.com/Andr-07/rate-limiter/pkg/redis"
)

type RateLimiterConfig struct {
	// MaxRequestsPerUser defines how many requests a single user can make within the given TimeWindow.
	MaxRequestsPerUser int

	// TimeWindow is the duration in which MaxRequestsPerUser and MaxRequestsPerIP are counted.
	TimeWindow time.Duration

	// MaxRequestsPerIP sets the request limit for a single IP address within the TimeWindow.
	MaxRequestsPerIP int

	// BlockDuration is the time a user or IP will be blocked after exceeding the rate limit.
	BlockDuration time.Duration

	// EnableKafkaLog enables sending rate-limit events to Kafka when set to true.
	EnableKafkaLog bool
}

type RateLimiter struct {
	RedisClient *redis.RedisClient
	KafkaClient *kafka.KafkaClient
	Config      *RateLimiterConfig
}

func New(redisClient *redis.RedisClient, kafkaClient *kafka.KafkaClient, config *RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		RedisClient: redisClient,
		KafkaClient: kafkaClient,
		Config:      config,
	}
}

func (r *RateLimiter) Allow(ctx context.Context) bool {
	userID, _ := ctx.Value(ictx.ContextUserKey).(string)
	ip, _ := ctx.Value(ictx.ContextIPKey).(string)

	return r.RedisClient.IsRequestAllowed(ctx, redis.RedisRequest{
		Ip:                 ip,
		UserID:             userID,
		MaxRequestsPerUser: r.Config.MaxRequestsPerUser,
		MaxRequestsPerIP:   r.Config.MaxRequestsPerIP,
		TimeWindow:         r.Config.TimeWindow,
		BlockDuration:      r.Config.BlockDuration,
	})
}

func (r *RateLimiter) Middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := req.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		ip, err := utils.GetClientIP(req)
		if err != nil || ip == "" {
			http.Error(w, "IP address is required", http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, ictx.ContextUserKey, userID)
		ctx = context.WithValue(ctx, ictx.ContextIPKey, ip)
		req = req.WithContext(ctx)

		if !r.Allow(req.Context()) {
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
