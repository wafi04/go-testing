package authhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/middleware"
	"github.com/wafi04/go-testing/common/pkg/logger"
	"github.com/wafi04/go-testing/gateway/helpers"
	"github.com/wafi04/go-testing/gateway/pkg/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandler struct {
	authClient pb.AuthServiceClient
	logger logger.Logger
}

func connectWithRetry(target string, service string) (*grpc.ClientConn, error) {
    maxAttempts := 5
    var conn *grpc.ClientConn
    var err error
    
    for i := 0; i < maxAttempts; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        log.Printf("Attempting to connect to %s service (attempt %d/%d)...", service, i+1, maxAttempts)
        
        conn, err = grpc.DialContext(ctx,
            target,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithBlock(),
        ) 
    
        if err == nil {
            log.Printf("Successfully connected to %s service", service)
            return conn, nil
        }
        
        log.Printf("Failed to connect to %s service: %v. Retrying...", service, err)
        time.Sleep(2 * time.Second)
    }
    
    return nil, fmt.Errorf("failed to connect to %s service after %d attempts: %v", service, maxAttempts, err)
}

func NewGateway(ctx context.Context) (*AuthHandler, error) {
    conn, err := connectWithRetry("auth_service:50051", "auth")
    if err != nil {
        return nil, err
    }
    
    return &AuthHandler{
        authClient: pb.NewAuthServiceClient(conn),
    }, nil
}



func (h *AuthHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received create user request: %s %s", r.Method, r.URL.Path)

	var req struct {
		Name string `json:"name"`
		Email  string  `json:"email"`
		Password string  `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded request: %+v", &req)
	clientIP := helpers.GetClientIP(r)
	userAgent := r.UserAgent()

	regis :=  &pb.CreateUserRequest{
		Name: req.Name,
		Email: req.Email,
		Password: req.Password,
		Role: "",
		IpAddress: clientIP,
		DeviceInfo: userAgent,
	}

	resp, err := h.authClient.CreateUser(r.Context(), regis)

	if err != nil {
		log.Printf("Error from auth service: %v", err.Error())
		http.Error(w, fmt.Sprintf("Error creating user: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	log.Printf("Received response from auth service: %+v", resp)

	w.Header().Set("Content-Type", "application/json")

	response := response.Success(resp, "User Created Successfully")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
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

	clientIP := helpers.GetClientIP(r)
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

	response := response.Success(resp, "Login Successfully")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {

    userid, err := middleware.GetUserFromContext(r.Context())
	 if err != nil {
		response.Error(http.StatusUnauthorized, "Unauthorized")
        return
    }

	users ,err :=   h.authClient.GetUser(r.Context(), &pb.GetUserRequest{
		UserId: userid.UserId,
	})

    if err != nil {
		response.Error(http.StatusUnauthorized, "Unauthorized")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    response := response.Success(users, "Profile received successfully")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler)   HandleLogout(w http.ResponseWriter,  r *http.Request){
	user , err :=   middleware.GetUserFromContext(r.Context())
	token :=  r.URL.Query().Get("token")

	if err != nil {
		response.Error(http.StatusUnauthorized, "Unauthorized")
        return
	}

	logout,err :=  h.authClient.Logout(r.Context(), &pb.LogoutRequest{
		AccessToken: token,
		UserId: user.UserId,
	})

	if err != nil {
		response.Error(http.StatusUnauthorized, "Unauthorized")
        return
	}

	if err := json.NewEncoder(w).Encode(logout); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}