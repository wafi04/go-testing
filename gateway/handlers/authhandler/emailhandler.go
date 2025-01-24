package authhandler

import (
	"encoding/json"
	"net/http"

	pb "github.com/wafi04/go-testing/auth/grpc"
	"github.com/wafi04/go-testing/auth/middleware"
	"github.com/wafi04/go-testing/common/pkg/logger"
	"github.com/wafi04/go-testing/gateway/pkg/response"
)

func (h *AuthHandler)  HandleVerifyEmail(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var code struct {
		Code string  `json:"code"`
		Token  string  `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&code); err != nil {
		h.logger.Log(logger.ErrorLevel, "Send : %v"  ,code.Token)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}


	h.logger.Log(logger.InfoLevel, "Send : %v"  ,code.Token)
	user,err := h.authClient.VerifyEmail(r.Context(), &pb.VerifyEmailRequest{
		VerificationToken: code.Token,
		VerifyCode: code.Code,
	})
	if err != nil {
		h.logger.Log(logger.ErrorLevel, "Failed to validate token : %v",err)
		return
	}
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
	token :=  r.URL.Query().Get("token")
	user, err :=  middleware.GetUserFromContext(r.Context())

	if err != nil {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
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