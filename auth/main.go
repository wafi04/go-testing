package main

import (
	"context"
	"log"
	"net"

	pb "github.com/wafi04/auth-go/grpc"
	"google.golang.org/grpc"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	log.Printf("Received CreateUser request for user: %v", req)

	return &pb.CreateUserResponse{
		UserId: "12345",
		Name:   "wafiuddin",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &authServer{})
	log.Printf("Auth server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
