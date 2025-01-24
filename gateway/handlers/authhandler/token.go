package authhandler

import (
	"encoding/json"
	"net/http"

	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/middleware"
	"github.com/wafi04/go-testing/common/pkg/logger"
)


func (s *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
    s.logger.Log(logger.InfoLevel, "Handle Refresh Token incoming")

    session := r.URL.Query().Get("session")
    refreshToken := r.URL.Query().Get("token")

    if session == "" || refreshToken == "" {
        http.Error(w, "Invalid session or token", http.StatusBadRequest)
        return
    }

    _, err := middleware.GetUserFromContext(r.Context())
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    refreshResp, err := s.authClient.RefreshToken(r.Context(), &pb.RefreshTokenRequest{
    	RefreshToken: refreshToken,
        SessionId:    session,
    })
    if err != nil {
        http.Error(w, "Token refresh failed", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "access_token":  refreshResp.AccessToken,
        "refresh_token": refreshResp.RefreshToken,
        "expires_at":    refreshResp.ExpiresAt,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}