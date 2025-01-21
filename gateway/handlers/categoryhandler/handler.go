package categoryhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wafi04/go-testing/auth/middleware"
	pb "github.com/wafi04/go-testing/category/grpc"
	"github.com/wafi04/go-testing/gateway/pkg/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CategoryHandler struct {
	categoryClient pb.CategoryServiceClient
}
func RegisterCategoryHandler(router *mux.Router, handler *CategoryHandler) {
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
    router.HandleFunc("/category", handler.HandleCreateCategory).Methods("POST", "OPTIONS")
    router.HandleFunc("/category", handler.HandleGetCategories).Methods("GET", "OPTIONS")
    router.HandleFunc("/list-categories", handler.HandleListCategories).Methods("GET", "OPTIONS")
    router.HandleFunc("/category/{id}", handler.HandleUpdateCategory).Methods("PUT", "OPTIONS") 
    router.HandleFunc("/category/{id}", handler.HandleDeleteCategory).Methods("DELETE", "OPTIONS") 

}

func NewCategoryGateway(ctx context.Context) (*CategoryHandler, error) {
	log.Println("Attempting to connect to auth service...")

	conn, err := grpc.DialContext(ctx,
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to connect to category service: %v", err)
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	log.Println("Successfully connected to auth service")
	return &CategoryHandler{
		categoryClient: pb.NewCategoryServiceClient(conn),
	}, nil
}

func (h *CategoryHandler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received create user request: %s %s", r.Method, r.URL.Path)

	var req pb.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded request: %+v", &req)

	resp, err := h.categoryClient.CreateCategory(r.Context(),&pb.CreateCategoryRequest{
		Name: req.Name,
		Description: req.Description,
		Image: req.Image,
		ParentId: req.ParentId,
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
func (h *CategoryHandler) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
    categories, err := h.categoryClient.ListCategories(r.Context(), &pb.ListCategoriesRequest{})
    if err != nil {
        response.Error(http.StatusBadRequest,"Failed to retrieve categories")
        return
    }

    response.Success(categories, "Categories retrieved successfully")
}

func (h *CategoryHandler) HandleListCategories(w http.ResponseWriter, r *http.Request) {

    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    parentID := r.URL.Query().Get("parent_id")
    includeChildren := r.URL.Query().Get("include_children") == "true"

    if page <= 0 {
        page = 1
    }
    if limit <= 0 {
        limit = 10
    }

    req := &pb.ListCategoriesRequest{
        Page:            int32(page),
        Limit:           int32(limit),
        IncludeChildren: includeChildren,
    }
    
    if parentID != "" {
        req.ParentId = &parentID
    }

    resp, err := h.categoryClient.ListCategories(r.Context(), req)
    if err != nil {
        log.Printf("Error calling ListCategories: %v", err)
        http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    response := response.Success(map[string]interface{}{
        "categories": resp.Categories,
        "total":     resp.Total,
        "page":      page,
        "limit":     limit,
    }, "Categories retrieved successfully")
    
    if err = json.NewEncoder(w).Encode(response); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }
}


func (h *CategoryHandler) HandleUpdateCategory(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    vars := mux.Vars(r)
    id, ok := vars["id"]
    if !ok {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }

    var request struct {
        Name        *string `json:"name,omitempty"`
        Description *string `json:"description,omitempty"`
        Image       *string `json:"image,omitempty"`
        ParentID    *string `json:"parent_id,omitempty"`
        Depth       *int32  `json:"depth,omitempty"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Create gRPC request
    updateReq := &pb.UpdateCategoryRequest{
        Id:          id,
        Name:        request.Name,
        Description: request.Description,
        Image:       request.Image,
        ParentId:    request.ParentID,
        Depth:       request.Depth,
    }

    ctx := r.Context()
    category, err := h.categoryClient.UpdateCategory(ctx, updateReq)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "not found"):
            http.Error(w, err.Error(), http.StatusNotFound)
        case strings.Contains(err.Error(), "invalid"):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    response :=  response.Success(category,"Update Category Succcess")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }
}

func (h *CategoryHandler) HandleDeleteCategory(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    vars := mux.Vars(r)
    id, ok := vars["id"]
    if !ok {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }

  

    // Create gRPC request
    updateReq := &pb.DeleteCategoryRequest{
        Id:          id,
        DeleteChildren: false,
    }

    ctx := r.Context()
    category, err := h.categoryClient.DeleteCategory(ctx, updateReq)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "not found"):
            http.Error(w, err.Error(), http.StatusNotFound)
        case strings.Contains(err.Error(), "invalid"):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    response :=  response.Success(category,"Delete Category Succcess")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }
}