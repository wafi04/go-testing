package authhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	pb "github.com/wafi04/go-testing/auth/grpc"
	middleware "github.com/wafi04/go-testing/auth/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandler struct {
	authClient pb.AuthServiceClient
}

func RegisterAuthHandler(router *mux.Router, handler *AuthHandler) {
	router.HandleFunc("/register", handler.HandleCreateUser).Methods("POST")
	router.HandleFunc("/login", handler.HandleLogin).Methods("POST")
}

func NewGateway(ctx context.Context) (*AuthHandler, error) {
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
	return &AuthHandler{
		authClient: pb.NewAuthServiceClient(conn),
	}, nil
}

func getClientIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	remoteAddr := r.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		return strings.Split(remoteAddr, ":")[0]
	}

	return remoteAddr
}

func (h *AuthHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
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

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Login
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	loginReq := &pb.LoginRequest{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: userAgent,
		IpAddress:  clientIP,
	}

	resp, err := h.authClient.Login(r.Context(), loginReq)
	if err != nil {
		log.Printf("Login failed: %v", err)

		switch {
		case strings.Contains(err.Error(), "invalid credentials"):
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		case strings.Contains(err.Error(), "account is deactivated"):
			http.Error(w, "Account is deactivated", http.StatusForbidden)
		case strings.Contains(err.Error(), "user not found"):
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", resp.AccessToken))

	response := map[string]interface{}{
		"access_token": resp.AccessToken,
		"user": map[string]interface{}{
			"user_id": resp.UserId,
			"name":    "wafi",
			"email":   loginReq.Email,
		},
		"session": map[string]interface{}{
			"session_id":  resp.SessionInfo.SessionId,
			"device_info": resp.SessionInfo.DeviceInfo,
			"ip_address":  resp.SessionInfo.IpAddress,
			"created_at":  resp.SessionInfo.CreatedAt,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) HandleGetProfile(w *http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create response
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id":           user.UserId,
			"name":              user.Name,
			"email":             user.Email,
			"role":              user.Role,
			"is_email_verified": user.IsEmailVerified,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

}
