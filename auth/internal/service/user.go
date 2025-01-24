package service

import (
	"context"

	"github.com/wafi04/common/pkg/logger"
	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserRepository *user.UserRepository
	log            logger.Logger
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (pb.CreateUserResponse, error) {

	hashPw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		s.log.Log(logger.ErrorLevel, "Failes Password : %v", err)
	}
	return s.UserRepository.CreateUser(ctx, &pb.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashPw),
		Role:     "",
		IpAddress: req.IpAddress,
		DeviceInfo: req.DeviceInfo,
	})
}

func (s *UserService) Login(ctx context.Context, login *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.UserRepository.Login(ctx, login)
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return s.UserRepository.GetUser(ctx, req)
}

func (s *UserService) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest)(*pb.VerifyEmailResponse,error){
	return s.UserRepository.VerifyEmail(ctx, req)
}
func (s *UserService) ResendVerification(ctx context.Context, req *pb.ResendVerificationRequest)(*pb.ResendVerificationResponse,error){
	return s.UserRepository.ResendVerification(ctx, req)
}

func (s *UserService) Logout(ctx context.Context, req *pb.LogoutRequest)(*pb.LogoutResponse,error){
	return s.UserRepository.Logout(ctx, req)
}
func (s *UserService) RevokeSession(ctx context.Context, req *pb.RevokeSessionRequest)(*pb.RevokeSessionResponse,error){
	return s.UserRepository.RevokeSession(ctx, req)
}
func (s *UserService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest)(*pb.RefreshTokenResponse,error){
	return s.UserRepository.RefreshToken(ctx, req)
}
func (s *UserService) ListSessions(ctx context.Context, req *pb.ListSessionsRequest)(*pb.ListSessionsResponse,error){
	return s.UserRepository.ListSessions(ctx, req)
}
