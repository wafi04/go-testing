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
func (s *AuthHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	log.Printf("Received verify email request for user: %v", req)

	user, err := s.UserService.VerifyEmail(ctx, req)
	if err != nil {
		log.Fatalf("Failed  to verifiy email  :%v ",err)
		return nil, err
	}

	return &pb.VerifyEmailResponse{
		Success: true,
		UserId: user.UserId,
	},nil
}
func (s *AuthHandler) ResendVerification(ctx context.Context, req *pb.ResendVerificationRequest) (*pb.ResendVerificationResponse, error) {
	log.Printf("Received verify email request for user: %v", req)

	user, err := s.UserService.ResendVerification(ctx, req)
	if err != nil {
		log.Fatalf("Failed  to verifiy email  :%v ",err)
		return nil, err
	}

	return user,nil
}
func (s *AuthHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse,error) {

	user, err := s.UserService.GetUser(ctx, req)
	if err != nil {
		return &pb.GetUserResponse{}, err
	}

	return user, nil
}
func (s *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse,error) {

	user, err := s.UserService.Logout(ctx, req)
	if err != nil {
		return &pb.LogoutResponse{}, err
	}

	return user, nil
}
func (s *AuthHandler) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest) (*pb.RevokeSessionResponse,error) {

	user, err := s.UserService.RevokeSession(ctx, req)
	if err != nil {
		return nil, err
	}
	return user, nil
}


func  (s *AuthHandler)   RefreshToken(ctx context.Context,req *pb.RefreshTokenRequest)(*pb.RefreshTokenResponse,error){
	log.Printf("Received verify email request for user: %v", req)

	refresh, err := s.UserService.RefreshToken(ctx, req)
	if err != nil {
		log.Fatalf("Failed  to refresh token  :%v ",err)
		return nil, err
	}

	return refresh,nil
}

func(s *AuthHandler)  ListSessions(ctx context.Context,req *pb.ListSessionsRequest)(*pb.ListSessionsResponse,error){
	log.Printf("Received verify email request for user: %v", req)

	ListSessions, err := s.UserService.ListSessions(ctx, req)
	if err != nil {
		log.Fatalf("Failed  to ListSessions token  :%v ",err)
		return nil, err
	}

	return ListSessions,nil
}
