package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/wafi04/go-testing/common/pkg/logger"
	pb "github.com/wafi04/go-testing/product/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductService struct {
	DB *sql.DB
	log logger.Logger
}

func NewProductService(db *sql.DB)  *ProductService{
	return &ProductService{
		DB: db,
	}
}



func (s *ProductService)  CreateProduct(ctx context.Context,req *pb.CreateProductRequest) (*pb.Product,error){
	productID :=   uuid.New().String()

	query := `
	INSERT INTO products (
		id,
		name,
		sub_title,
		description,
		sku,
		price,
		category_id,
		created_at,
		updated_at
	)
	VALUES
	($1,$2,$3,$4,$5,$6,$7,$8)
	RETURNING (id,name,sub_title,description,sku,price,category_id,created_at,updated_at)
	`

	var product pb.Product
    var createdAt ,updated_at sql.NullTime


    err := s.DB.QueryRowContext(ctx, query,
        productID,
        req.Product.Name,
        req.Product.SubTitle,
        req.Product.Description,
        req.Product.Sku,
        req.Product.Price,
        req.Product.CategoryId,
		
    ).Scan(
        &product.Id,
        &product.Name,
        &product.SubTitle,
        &product.Description,
        &product.Sku,
        &product.Price,
        &product.CategoryId,
		&createdAt,
		&updated_at,
    )


	if err != nil {
		s.log.Log(logger.ErrorLevel,"Failed to insert product  : %v",err)
		return nil, fmt.Errorf("failed to insert : %v", err)
	}

	if createdAt.Valid {
        product.CreatedAt = timestamppb.New(createdAt.Time)
    }
	if updated_at.Valid {
        product.UpdatedAt = timestamppb.New(updated_at.Time)
    }

	return &product,nil
}