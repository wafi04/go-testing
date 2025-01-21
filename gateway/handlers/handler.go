package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wafi04/go-testing/auth/middleware"
	"github.com/wafi04/go-testing/gateway/handlers/authhandler"
	"github.com/wafi04/go-testing/gateway/handlers/categoryhandler"
	"github.com/wafi04/go-testing/gateway/handlers/producthandler"
)
func SetupRoutes(
	authGateway  *authhandler.AuthHandler,
	categoryGateway  *categoryhandler.CategoryHandler,
	productGateway  *producthandler.ProductHandler,
) *mux.Router{
	r :=   mux.NewRouter()

	 r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
            w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
            w.Header().Set("Access-Control-Allow-Credentials", "true")

            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            next.ServeHTTP(w, r)
        })
    })

    // Public routes
    public := r.PathPrefix("").Subrouter()
    public.HandleFunc("/auth/register", authGateway.HandleCreateUser).Methods("POST", "OPTIONS")
    public.HandleFunc("/auth/login", authGateway.HandleLogin).Methods("POST", "OPTIONS")

    // Protected routes
    protected := r.PathPrefix("").Subrouter()
    protected.Use(middleware.AuthMiddleware)

    // Auth protected routes
    protected.HandleFunc("/auth/profile", authGateway.HandleGetProfile).Methods("GET", "OPTIONS")

    // Category protected routes
    protected.HandleFunc("/category", categoryGateway.HandleCreateCategory).Methods("POST", "OPTIONS")
    protected.HandleFunc("/category", categoryGateway.HandleGetCategories).Methods("GET", "OPTIONS")
    protected.HandleFunc("/list-categories", categoryGateway.HandleListCategories).Methods("GET", "OPTIONS")
    protected.HandleFunc("/category/{id}", categoryGateway.HandleUpdateCategory).Methods("PUT", "OPTIONS")
    protected.HandleFunc("/category/{id}", categoryGateway.HandleDeleteCategory).Methods("DELETE", "OPTIONS")

    // Product protected routes
    protected.HandleFunc("/product", productGateway.HandleCreateProduct).Methods("POST", "OPTIONS")
    protected.HandleFunc("/product/{id}", productGateway.HandleGetProduct).Methods("GET", "OPTIONS")
    protected.HandleFunc("/product", productGateway.HandleListProducts).Methods("GET", "OPTIONS")

    return r
}