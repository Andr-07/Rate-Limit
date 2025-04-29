package redis

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"rate-limiter/config"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisRequest struct {
	Ip                 string
	UserID             string
	MaxRequestsPerUser int
	MaxRequestsPerIP   int
	TimeWindow         time.Duration
	BlockDuration      time.Duration
}
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

func (r *RedisClient) IsRequestAllowed(ctx context.Context, config RedisRequest) bool {
	now := time.Now().Unix()
	windowStart := now - int64(config.TimeWindow.Seconds())

	userKey := fmt.Sprintf("ratelimit:user:%s", config.UserID)
	ipKey := fmt.Sprintf("ratelimit:ip:%s", config.Ip)
	userBlockKey := fmt.Sprintf("ratelimit:block:user:%s", config.UserID)
	ipBlockKey := fmt.Sprintf("ratelimit:block:ip:%s", config.Ip)

	// Check if user is blocked
	if blockEndStr, err := r.Client.Get(ctx, userBlockKey).Result(); err == nil && blockEndStr != "" {
		if blockEnd, _ := strconv.ParseInt(blockEndStr, 10, 64); now < blockEnd {
			fmt.Println("User is blocked until", blockEnd)
			return false
		}
		_ = r.Client.Del(ctx, userBlockKey).Err()
	}

	// Check if IP is blocked
	if blockEndStr, err := r.Client.Get(ctx, ipBlockKey).Result(); err == nil && blockEndStr != "" {
		if blockEnd, _ := strconv.ParseInt(blockEndStr, 10, 64); now < blockEnd {
			fmt.Println("IP is blocked until", blockEnd)
			return false
		}
		_ = r.Client.Del(ctx, ipBlockKey).Err()
	}

	// Helper function to track request
	trackRequest := func(key string) (int64, error) {
		member := fmt.Sprintf("%d:%d", now, rand.Int63())
		if _, err := r.Client.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: member}).Result(); err != nil {
			return 0, err
		}
		if _, err := r.Client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart)).Result(); err != nil {
			return 0, err
		}
		return r.Client.ZCard(ctx, key).Result()
	}

	userCount, err := trackRequest(userKey)
	if err != nil {
		fmt.Println("Error tracking user requests:", err)
		return false
	}

	ipCount, err := trackRequest(ipKey)
	if err != nil {
		fmt.Println("Error tracking IP requests:", err)
		return false
	}

	// Block user if over limit
	if int(userCount) > config.MaxRequestsPerUser {
		blockEnd := now + int64(config.BlockDuration.Seconds())
		if err := r.Client.Set(ctx, userBlockKey, blockEnd, config.BlockDuration).Err(); err != nil {
			fmt.Println("Error setting user block:", err)
		}
		return false
	}

	// Block IP if over limit
	if int(ipCount) > config.MaxRequestsPerIP {
		blockEnd := now + int64(config.BlockDuration.Seconds())
		if err := r.Client.Set(ctx, ipBlockKey, blockEnd, config.BlockDuration).Err(); err != nil {
			fmt.Println("Error setting IP block:", err)
		}
		return false
	}

	return true
}
