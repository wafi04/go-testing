package handler

import (
	"context"

	"github.com/wafi04/go-testing/common/pkg/logger"
	pb "github.com/wafi04/go-testing/product/grpc"
	"github.com/wafi04/go-testing/product/service"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	productService *service.ProductService
	log  logger.Logger
}


func NewProductHandler(service  *service.ProductService)  *ProductHandler{
	return  &ProductHandler{
		productService: service,
	}
}

func (h *ProductHandler)  CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product,error){
	h.log.Log(logger.InfoLevel, "incoming request ")
	return h.productService.CreateProduct(ctx, req)
}

func (h *ProductHandler)  GetProduct(ctx context.Context,req *pb.GetProductRequest)  (*pb.Product,error){
	h.log.Log(logger.InfoLevel, "incoming request ")
	return h.productService.GetProduct(ctx, req)
}
func (h *ProductHandler)  ListProducts(ctx context.Context,req *pb.ListProductsRequest)  (*pb.ListProductsResponse,error){
	h.log.Log(logger.InfoLevel, "incoming request ")
	return h.productService.ListProducts(ctx, req)
}