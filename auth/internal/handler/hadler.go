package handler

import (
	"context"
	"log"

	pb "github.com/wafi04/go-testing/auth/grpc"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
}

func (s *AuthHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Printf("Received CreateUser request for user: %v", req)

	return &pb.CreateUserResponse{
		UserId: "12345",
		Name:   "wafiuddin",
	}, nil
}
