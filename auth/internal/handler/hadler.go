package handler

import (
	"context"
	"log"
	"time"

	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/internal/service"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	UserService *service.UserService
}

func (s *AuthHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Printf("Received CreateUser request for user: %v", req)

	user, err := s.UserService.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserResponse{
		UserId:    user.UserId,
		Name:      user.Email,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: time.Now().Unix(),
	}, nil
}
func (s *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Received Login request for user: %v", req)

	user, err := s.UserService.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return user, nil
}
func (s *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse,error) {

	user, err := s.UserService.GetUser(ctx, req.UserId)
	if err != nil {
		return &pb.GetUserResponse{}, err
	}

	return user, nil
}
