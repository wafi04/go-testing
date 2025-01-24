package authhandler

import (
	"encoding/json"
	"net/http"
	"sync"

	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/middleware"
	"github.com/wafi04/go-testing/common/pkg/logger"
	"github.com/wafi04/go-testing/gateway/pkg/response"
)

var tokenStore = &TokenStore{
    tokens: make(map[string]string),
}

type TokenStore struct {
	tokens map[string]string	
    mu     sync.RWMutex
}

func (ts *TokenStore) StoreToken(userId string, token string) {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    ts.tokens[userId] = token
}
func (ts *TokenStore) GetToken(userId string) (string, bool) {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    token, exists := ts.tokens[userId]
    return token, exists
}

func (ts *TokenStore) ValidateToken(userId string, token string) bool {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    storedToken, exists := ts.tokens[userId]
    return exists && storedToken == token
}

func (ts *TokenStore) ClearToken(userId string) {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    delete(ts.tokens, userId)
}


func (h *AuthHandler)  HandleVerifyEmail(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var code struct {
		Code string  `json:"code"`
	}
	users,err  :=   middleware.GetUserFromContext(r.Context())
	if err != nil {
		 http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
        return
	}
	 storedToken, exists := tokenStore.GetToken(users.UserId)
    if !exists   {
        http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
        return
    }
	if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
		h.logger.Log(logger.ErrorLevel, "Send : %v"  ,storedToken)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Log(logger.InfoLevel, "Send : %v"  ,storedToken)

	user,err := h.authClient.VerifyEmail(r.Context(), &pb.VerifyEmailRequest{
		VerificationToken: storedToken,
		VerifyCode: code.Code,
	})
	if err != nil {
		h.logger.Log(logger.ErrorLevel, "Failed to validate token : %v",err)
		return
	}

	tokenStore.ClearToken(user.UserId)
	res := response.Success(user,"Success to Verification Email")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Log(logger.ErrorLevel,"Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

var Type  struct {
	Type   string  `json:"type"`
}

func (h *AuthHandler)  HandleResendVerification(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	user,err :=  middleware.GetUserFromContext(r.Context())

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
	}
	token,err :=  middleware.GenerateToken(&pb.UserInfo{
		UserId: user.UserId,
		Role : user.Role,
		Name: user.Name,
		Email: user.Email,
	})

    tokenStore.StoreToken(user.UserId, token)
	if err != nil {
        http.Error(w, "Invalid Generate Token", http.StatusBadRequest)
        return
    }

	_,err =  middleware.ValidateToken(token)
	if err != nil {
		 http.Error(w, "Token is invalid", http.StatusBadRequest)
        return
	} 

	verif,err := h.authClient.ResendVerification(r.Context(), &pb.ResendVerificationRequest{
		UserId: user.UserId,
		Type: "EMAIL_VERIFICATION",
		Token: token,
	})

	if err != nil {
		h.logger.Log(logger.ErrorLevel, "Failed to validate token : %v",err)
	}
	res := response.Success(verif,"Success to Verification Email")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Log(logger.ErrorLevel,"Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}