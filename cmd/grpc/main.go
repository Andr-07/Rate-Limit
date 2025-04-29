package main

import (
	"fmt"
	"log"
	"net"

	configs "github.com/Andr-07/rate-limiter/config"
	grpcserver "github.com/Andr-07/rate-limiter/internal/grpc"
	"github.com/Andr-07/rate-limiter/pkg/kafka"
	"github.com/Andr-07/rate-limiter/pkg/limiter"
	"github.com/Andr-07/rate-limiter/pkg/redis"
	pb "github.com/Andr-07/rate-limiter/proto/limiterpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// инициализация redis и kafka
	redisClient := redis.NewRedisClient(redis.RedisClientDeps{
		Config: &configs.RedisConfig{
			Addr:     "localhost:6379",
			Password: "yourpassword",
		},
	})
	kafkaClient, _ := kafka.NewKafkaClient(kafka.KafkaClientDeps{
		Config: &configs.KafkaConfig{
			Addr: "localhost:9092",
		}},
	)

	rl := limiter.New(redisClient, kafkaClient, &limiter.RateLimiterConfig{
		MaxRequestsPerUser: 10,
		MaxRequestsPerIP:   20,
		TimeWindow:         60,
		BlockDuration:      30,
	})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRateLimiterServiceServer(s, grpcserver.NewServer(rl))

	reflection.Register(s)
	fmt.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
