package helpers

import (
	"log"
	"net/http"
)
func CorsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("CORS Middleware - Method: %s, Path: %s", r.Method, r.URL.Path)
        
        // Set CORS headers for all responses
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        // Handle preflight
        if r.Method == "OPTIONS" {
            log.Println("Handling OPTIONS request")
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}