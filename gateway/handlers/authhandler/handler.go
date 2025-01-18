package authhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	pb "github.com/wafi04/auth-go/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHanndler struct {
	authClient pb.AuthServiceClient
}

func RegisterAuthHandler(router *mux.Router, handler *AuthHanndler) {
	router.HandleFunc("/users", handler.HandleCreateUser).Methods("POST")
}

func NewGateway(ctx context.Context) (*AuthHanndler, error) {
	log.Println("Attempting to connect to auth service...")

	conn, err := grpc.DialContext(ctx,
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to connect to auth service: %v", err)
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	log.Println("Successfully connected to auth service")
	return &AuthHanndler{
		authClient: pb.NewAuthServiceClient(conn),
	}, nil
}

func (h *AuthHanndler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received create user request: %s %s", r.Method, r.URL.Path)

	var req pb.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded request: %+v", &req)

	resp, err := h.authClient.CreateUser(r.Context(), &pb.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     "Admin",
	})

	if err != nil {
		log.Printf("Error from auth service: %v", err.Error())
		http.Error(w, fmt.Sprintf("Error creating user: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	log.Printf("Received response from auth service: %+v", resp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": resp.UserId,
		"name":    resp.Name,
		"message": "hello world",
	})
}
