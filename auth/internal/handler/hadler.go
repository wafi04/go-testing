package handler

import (
	"context"

	pb "github.com/wafi04/go-testing/auth-go/grpc"
	"github.com/wafi04/go-testing/common/pkg/logger"
)

type AuthHanndler struct {
	pb.UnimplementedAuthServiceServer
	logger logger.Logger
}

func (s *AuthHanndler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	s.logger.Log(logger.InfoLevel, "Received CreateUser request for user: %v", req)

	return &pb.CreateUserResponse{
		UserId: "12345",
		Name:   "wafiuddin",
	}, nil
}
