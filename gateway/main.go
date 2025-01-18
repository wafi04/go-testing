package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/wafi04/go-testing/gateway/handlers/authhandler"
)

func main() {
	log.Println("Starting gateway service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gateway, err := authhandler.NewGateway(ctx)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	authhandler.RegisterAuthHandler(r, gateway)

	// Tambahkan logging untuk debugging
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("Registered route: %s (Methods: %v)", path, methods)
		return nil
	})

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Gateway server starting on http://localhost:8000")
	log.Fatal(srv.ListenAndServe())
}
