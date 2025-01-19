package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	pb "github.com/wafi04/go-testing/auth/grpc"
)

var jwtSecretKey = []byte("jsjxakabxjaigisyqyg189")

type JWTClaims struct {
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Role            string `json:"role"`
	IsActive        bool   `json:"is_active"`
	IsEmailVerified bool   `json:"is_email_verified"`
	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GenerateToken(user *pb.UserInfo) (string, error) {
	isEmailVerified := false
	if user.IsEmailVerified {
		isEmailVerified = true
	}

	claims := JWTClaims{
		UserID:          user.UserId,
		Email:           user.Email,
		Name:            user.Name,
		Role:            user.Role,
		IsEmailVerified: isEmailVerified,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "wafiuddin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

type contextKey string

const UserContextKey contextKey = "user"

func GetUserFromContext(ctx context.Context) (*pb.UserInfo, error) {
	user, ok := ctx.Value(UserContextKey).(*pb.UserInfo)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		user := &pb.UserInfo{
			UserId:          claims.UserID,
			Email:           claims.Email,
			Name:            claims.Name,
			Role:            claims.Role,
			IsEmailVerified: claims.IsEmailVerified,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
