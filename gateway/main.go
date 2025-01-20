package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/wafi04/go-testing/common/pkg/logger"
	"github.com/wafi04/go-testing/gateway/handlers/authhandler"
	"github.com/wafi04/go-testing/gateway/handlers/categoryhandler"
)

func main() {
    logs :=  logger.NewLogger()

    logs.Info("Staring Server gateway ")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    gateway, err := authhandler.NewGateway(ctx)
    if err != nil {
        logs.Log(logger.ErrorLevel, "Failed to connect Auth Service : %v",err)
    }
    categorygateway,err :=  categoryhandler.NewCategoryGateway(ctx)
    if err != nil {
        logs.Log(logger.ErrorLevel, "Failed to connect Category Service : %v",err)
    }
    
    r := mux.NewRouter()

    // 1. Logging middleware
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logs.Log(logger.InfoLevel,"Incoming request: %s %s", r.Method, r.URL.Path)
            next.ServeHTTP(w, r)
        })
    })

    // 2. CORS middleware HARUS sebelum registrasi route
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Set CORS headers untuk semua response
            w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
            w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
            w.Header().Set("Access-Control-Allow-Credentials", "true")

            // Handle preflight request
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    })
    
    // 3. Register routes
    authhandler.RegisterAuthHandler(r, gateway)
    categoryhandler.RegisterCategoryHandler(r, categorygateway)

    
    // Debug logging untuk routes
    r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
        path, _ := route.GetPathTemplate()
        methods, _ := route.GetMethods()
        logs.Log(logger.InfoLevel,"Registered route: %s (Methods: %v)", path, methods)
        return nil
    })
    
    srv := &http.Server{
        Handler:      r,
        Addr:         "127.0.0.1:4000",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }
    
    logs.Log(logger.InfoLevel,"Gateway server starting on http://localhost:4000")
    log.Print(srv.ListenAndServe())
}