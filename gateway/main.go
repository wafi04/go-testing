package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/wafi04/go-testing/common/pkg/logger"
	"github.com/wafi04/go-testing/gateway/handlers"
	"github.com/wafi04/go-testing/gateway/handlers/authhandler"
	"github.com/wafi04/go-testing/gateway/handlers/categoryhandler"
	"github.com/wafi04/go-testing/gateway/handlers/producthandler"
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
    productGateway,err :=  producthandler.NewCategoryGateway(ctx)
    if err != nil {
        logs.Log(logger.ErrorLevel, "Failed to connect product Service : %v",err)
    }
    
    r := mux.NewRouter()

    // 1. Logging middleware
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logs.Log(logger.InfoLevel,"Incoming request: %s %s", r.Method, r.URL.Path)
            next.ServeHTTP(w, r)
        })
    })

    r = handlers.SetupRoutes(gateway, categorygateway, productGateway)
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