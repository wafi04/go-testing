package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/wafi04/go-testing/common/pkg/logger"
	pb "github.com/wafi04/go-testing/product/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductService  struct {
	db *sqlx.DB  
	log logger.Logger
}

func  NewProductService(db *sqlx.DB)  *ProductService{
	return &ProductService{
		db: db,
	}
}	


func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
    now := time.Now()
    productID := uuid.New().String()
    query := `
    INSERT INTO products  
    (id, name, sub_title, description, sku, price, category_id, created_at, updated_at)
    VALUES 
    ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING id, name, sub_title, description, sku, price, category_id, created_at, updated_at
    `

    var product pb.Product
    var createdAt, updatedAt time.Time

    err := s.db.QueryRowContext(ctx, query,
        productID,
        req.Product.Name,
        req.Product.SubTitle,
        req.Product.Description,
        req.Product.Sku,
        req.Product.Price,
        req.Product.CategoryId,
        now,
        now,
    ).Scan(
        &product.Id,
        &product.Name,
        &product.SubTitle,
        &product.Description,
        &product.Sku,
        &product.Price,
        &product.CategoryId,
        &createdAt,
        &updatedAt,
    )

    if err != nil {
        s.log.Log(logger.ErrorLevel, "Failed to insert product: %v", err)
        return nil, fmt.Errorf("failed to insert Product: %v", err)
    }

    product.CreatedAt = timestamppb.New(createdAt)
    product.UpdatedAt = timestamppb.New(updatedAt)

    return &product, nil
}



func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	// Main product query
	productQuery := `
		SELECT 
			id,
			name,
			sub_title,
			description,
			price,
			sku,
			category_id,
			created_at,
			updated_at
		FROM products
		WHERE id = $1
	`

	product := &pb.Product{}
	var subTitle sql.NullString
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, productQuery, req.Id).Scan(
		&product.Id,
		&product.Name,
		&subTitle,
		&product.Description,
		&product.Price,
		&product.Sku,
		&product.CategoryId,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Log(logger.ErrorLevel, "Product not found: %v", err)
			return nil, fmt.Errorf("product not found")
		}
		s.log.Log(logger.ErrorLevel, "Failed to get product: %v", err)
		return nil, fmt.Errorf("failed to get product")
	}

	// Handle nullable subtitle
	if subTitle.Valid {
		product.SubTitle = &subTitle.String
	}

	// Convert timestamps
	product.CreatedAt = timestamppb.New(createdAt)
	product.UpdatedAt = timestamppb.New(updatedAt)

	variantsQuery := `
		SELECT 
			v.id,
			v.color,
			v.sku,
			i.id,
			i.size,
			i.stock,
			i.reserved_stock,
			i.available_stock,
			img.id,
			img.url,
			img.is_main
		FROM product_variants v
		LEFT JOIN inventory i ON v.id = i.variant_id
		LEFT JOIN product_images img ON v.id = img.variant_id
		WHERE v.product_id = $1
		ORDER BY v.id, i.size, img.is_main DESC
	`

	rows, err := s.db.QueryContext(ctx, variantsQuery, req.Id)
	if err != nil {
		s.log.Log(logger.ErrorLevel, "Failed to get variants: %v", err)
		return nil, fmt.Errorf("failed to get product variants")
	}
	defer rows.Close()

	variantMap := make(map[string]*pb.ProductVariant)
	
	for rows.Next() {
		var variantId, variantColor, variantSku string
		var inventoryId, size sql.NullString
		var stock, reservedStock, availableStock sql.NullInt32
		var imageId, imageUrl sql.NullString
		var isMain sql.NullBool

		err := rows.Scan(
			&variantId,
			&variantColor,
			&variantSku,
			&inventoryId,
			&size,
			&stock,
			&reservedStock,
			&availableStock,
			&imageId,
			&imageUrl,
			&isMain,
		)

		if err != nil {
			s.log.Log(logger.ErrorLevel, "Failed to scan variant row: %v", err)
			continue
		}

		// Get or create variant
		variant, exists := variantMap[variantId]
		if !exists {
			variant = &pb.ProductVariant{
				Id:        variantId,
				Color:     variantColor,
				Sku:       variantSku,
				ProductId: req.Id,
				Images:    make([]*pb.ProductImage, 0),
				Inventory: make([]*pb.Inventory, 0),
			}
			variantMap[variantId] = variant
		}

		// Add inventory if exists
		if inventoryId.Valid && size.Valid {
			inventory := &pb.Inventory{
				Id:             inventoryId.String,
				VariantId:     variantId,
				Size:          size.String,
				Stock:         stock.Int32,
				ReservedStock: reservedStock.Int32,
				AvailableStock: availableStock.Int32,
			}
			variant.Inventory = append(variant.Inventory, inventory)
		}

		if imageId.Valid && imageUrl.Valid {
			imageExists := false
			for _, img := range variant.Images {
				if img.Id == imageId.String {
					imageExists = true
					break
				}
			}
			
			if !imageExists {
				image := &pb.ProductImage{
					Id:        imageId.String,
					Url:       imageUrl.String,
					VariantId: variantId,
					IsMain:    isMain.Bool,
				}
				variant.Images = append(variant.Images, image)
			}
		}
	}

	product.Variants = make([]*pb.ProductVariant, 0, len(variantMap))
	for _, variant := range variantMap {
		product.Variants = append(product.Variants, variant)
	}

	return product, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
    if req.PageSize == 0 {
        req.PageSize = 10
    }

    baseQuery := `
        SELECT 
            p.id,
            p.name,
            p.sub_title,
            p.description,
            p.price,
            p.sku,
            p.category_id,
            p.created_at,
            p.updated_at,
            (
                SELECT COALESCE(JSON_AGG(v.*), '[]'::json)
                FROM product_variants v
                WHERE v.product_id = p.id
            ) as variants
        FROM 
            products p
        WHERE 1=1
        ORDER BY p.created_at DESC
        LIMIT :page_size
        OFFSET (:page_size * COALESCE(NULLIF(:page_token, ''), '0')::integer)
    `

    params := map[string]interface{}{
        "page_size":  req.PageSize,
        "page_token": req.PageToken,
    }

    rows, err := s.db.NamedQueryContext(ctx, baseQuery, params)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to query products: %v", err)
    }
    defer rows.Close()

    var products []*pb.Product
    for rows.Next() {
        var product struct {
            ID          string          `db:"id"`
            Name        string          `db:"name"`
            SubTitle    sql.NullString  `db:"sub_title"`
            Description string          `db:"description"`
            Price       float64         `db:"price"`
            SKU         string          `db:"sku"`
            CategoryID  string          `db:"category_id"`
            CreatedAt   time.Time       `db:"created_at"`
            UpdatedAt   time.Time       `db:"updated_at"`
            Variants    json.RawMessage `db:"variants"`
        }

        if err := rows.StructScan(&product); err != nil {
            return nil, status.Errorf(codes.Internal, "failed to scan product: %v", err)
        }

        pbProduct := &pb.Product{
            Id:          product.ID,
            Name:        product.Name,
            Description: product.Description,
            Price:      product.Price,
            Sku:        product.SKU,
            CategoryId: product.CategoryID,
            CreatedAt:  timestamppb.New(product.CreatedAt),
            UpdatedAt:  timestamppb.New(product.UpdatedAt),
        }
        if product.SubTitle.Valid {
            pbProduct.SubTitle = &product.SubTitle.String
        }

        var variants []*pb.ProductVariant
        if err := json.Unmarshal(product.Variants, &variants); err != nil {
            return nil, status.Errorf(codes.Internal, "failed to parse variants: %v", err)
        }
        pbProduct.Variants = variants

        products = append(products, pbProduct)
    }

    if err = rows.Err(); err != nil {
        return nil, status.Errorf(codes.Internal, "error iterating products: %v", err)
    }

    nextPageToken := ""
    if len(products) == int(req.PageSize) {
        currentPage := 0
        if req.PageToken != "" {
            currentPage, _ = strconv.Atoi(req.PageToken)
        }
        nextPageToken = strconv.Itoa(currentPage + 1)
    }

    return &pb.ListProductsResponse{
        Products:      products,
        NextPageToken: nextPageToken,
    }, nil
}