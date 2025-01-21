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
	})
}

func (s *UserService) Login(ctx context.Context, login *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.UserRepository.Login(ctx, login)
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*pb.GetUserResponse, error) {
	return s.UserRepository.GetUser(ctx, userID)
}
