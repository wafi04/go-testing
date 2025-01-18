package main

import (
	"log"
	"net"

	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/internal/handler"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &handler.AuthHandler{})
	log.Printf("Auth server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
