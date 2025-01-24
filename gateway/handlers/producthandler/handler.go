package producthandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/wafi04/go-testing/auth/middleware"
	apiresponse "github.com/wafi04/go-testing/common/pkg/response"
	"github.com/wafi04/go-testing/gateway/pkg/response"
	pb "github.com/wafi04/go-testing/product/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


type ProductHandler struct {
	productClient pb.ProductServiceClient
}


func RegisterProductHandler(router *mux.Router, handler *ProductHandler) {
    // router.Use(func(next http.Handler) http.Handler {
    //     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    //         w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
    //         w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT,DELETE, OPTIONS") 
    //         w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
    //         w.Header().Set("Access-Control-Allow-Credentials", "true")
            
    //         if r.Method == "OPTIONS" {
    //             w.WriteHeader(http.StatusOK)
    //             return
    //         }
    //         next.ServeHTTP(w, r)
    //     })
    // })
    router.Use(middleware.AuthMiddleware)
    router.HandleFunc("/product", handler.HandleCreateProduct).Methods("POST", "OPTIONS")
    router.HandleFunc("/product/{id}", handler.HandleGetProduct).Methods("GET", "OPTIONS")
    router.HandleFunc("/product", handler.HandleListProducts).Methods("GET", "OPTIONS")

}

func connectWithRetry(target string, service string) (*grpc.ClientConn, error) {
    maxAttempts := 5
    var conn *grpc.ClientConn
    var err error
    
    for i := 0; i < maxAttempts; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        log.Printf("Attempting to connect to %s service (attempt %d/%d)...", service, i+1, maxAttempts)
        
        conn, err = grpc.DialContext(ctx,
            target,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithBlock(),
        )
        
        if err == nil {
            log.Printf("Successfully connected to %s service", service)
            return conn, nil
        }
        
        log.Printf("Failed to connect to %s service: %v. Retrying...", service, err)
        time.Sleep(2 * time.Second)
    }
    
    return nil, fmt.Errorf("failed to connect to %s service after %d attempts: %v", service, maxAttempts, err)
}

func NewProductGateway(ctx context.Context) (*ProductHandler, error) {
    conn, err := connectWithRetry("product_service:50052", "product")
    if err != nil {
        return nil, err
    }
    
    return &ProductHandler{
        productClient: pb.NewProductServiceClient(conn),
    }, nil
}

func (h  *ProductHandler)   HandleCreateProduct(w http.ResponseWriter,  r *http.Request){
	log.Printf("Received create product request: %s %s", r.Method, r.URL.Path)

	var req ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded request: %+v", &req)

	if req.Sku == ""{
		req.Sku = GenerateSku(req.Name)
	}else if !IsSkuValid(req.Sku) {	
		return 
	}

	resp, err := h.productClient.CreateProduct(r.Context(),&pb.CreateProductRequest{
		Product: &pb.Product{
			Name: req.Name,
			Description: req.Description,
			SubTitle: req.SubTitle,
			Price: req.Price,
			Sku: req.Sku,
			CategoryId: req.CategoryId,
		},
	})

	if err != nil {
		log.Printf("Error from auth service: %v", err.Error())
		http.Error(w, fmt.Sprintf("Error creating user: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	log.Printf("Received response from auth service: %+v", resp)

	w.Header().Set("Content-Type", "application/json")

	response := response.Success(resp, "User Created Successfully")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}


func (h *ProductHandler)   HandleGetProduct(w http.ResponseWriter, r *http.Request)  {
	log.Printf("Received get product request: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
    vars := mux.Vars(r)
    id, ok := vars["id"]

	if !ok {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }

	res,err  := h.productClient.GetProduct(r.Context(), &pb.GetProductRequest{
		Id: id,
	})

	if err != nil {
		log.Printf("Failed to get Product: %v", err)
		apiresponse.SendErrorResponseWithDetails(w, http.StatusInternalServerError, "Failed to get products", err.Error())
		return
	}


	resp := response.Success(res,"Get product successfully")

	if err  := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

}



func (h *ProductHandler)   HandleListProducts(w http.ResponseWriter, r *http.Request){
    log.Printf("Received get product request: %s %s", r.Method, r.URL.Path)
    w.Header().Set("Content-Type", "application/json")

    // Get query parameters
    page := r.URL.Query().Get("page")
    limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
    if err != nil || limit <= 0 {
        limit = 10
    }

    // Call gRPC service
    res, err := h.productClient.ListProducts(r.Context(), &pb.ListProductsRequest{
        PageSize: int32(limit),
        PageToken: page,
    })
 
	if err != nil {
		log.Printf("Failed to get Product: %v", err)
		apiresponse.SendErrorResponseWithDetails(w, http.StatusInternalServerError, "Failed to get products", err.Error())
		return
	}

    // Success response
    resp := response.Success(res, "Pagination product successfully")
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }

	
}