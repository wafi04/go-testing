package handler

import (
	"context"

	pb "github.com/wafi04/go-testing/product/grpc"
	"github.com/wafi04/go-testing/product/service"
)

type ProductHandler  struct {
	pb.UnimplementedProductServiceServer
	productService  *service.ProductService
}


func NewProductHandler(service *service.ProductService) *ProductHandler{
	return &ProductHandler{
		productService: service,
	}
}

func (h *ProductHandler)  CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product,error){
	return  h.productService.CreateProduct(ctx, req)
}