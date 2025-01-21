package apiresponse

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
    Status  int    `json:"status"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
    Status  int         `json:"status"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    resp := ErrorResponse{
        Status:  statusCode,
        Message: message,
    }
    
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        log.Printf("Error encoding error response: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func SendErrorResponseWithDetails(w http.ResponseWriter, statusCode int, message, details string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    resp := ErrorResponse{
        Status:  statusCode,
        Message: message,
        Details: details,
    }
    
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        log.Printf("Error encoding error response: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func SendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    resp := SuccessResponse{
        Status:  statusCode,
        Message: message,
        Data:    data,
    }
    
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        log.Printf("Error encoding success response: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}