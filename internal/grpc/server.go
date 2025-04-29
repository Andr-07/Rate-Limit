package grpcserver

import (
	"context"

	"github.com/Andr-07/rate-limiter/pkg/limiter"
	pb "github.com/Andr-07/rate-limiter/proto/limiterpb"
)

type RateLimiterGRPCServer struct {
	pb.UnimplementedRateLimiterServiceServer
	Limiter *limiter.RateLimiter
}

func NewServer(l *limiter.RateLimiter) *RateLimiterGRPCServer {
	return &RateLimiterGRPCServer{Limiter: l}
}

func (s *RateLimiterGRPCServer) Allow(ctx context.Context, req *pb.AllowRequest) (*pb.AllowResponse, error) {
	ctx = context.WithValue(ctx, "userID", req.GetUserId())
	ctx = context.WithValue(ctx, "ip", req.GetIp())

	allowed := s.Limiter.Allow(ctx)

	return &pb.AllowResponse{Allowed: allowed}, nil
}
