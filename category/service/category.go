package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	pb "github.com/wafi04/go-testing/category/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CategoryService struct {
    DB     *sql.DB
}

func NewCategoryService(db *sql.DB) *CategoryService {
    return &CategoryService{
        DB:     db,
    }
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
    tx, err := s.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer tx.Rollback()

    // Generate unique ID
    categoryID := uuid.New().String()
    
    var depth int32 = 0
    
    if req.ParentId != nil {
        var parentDepth int32
        err := tx.QueryRowContext(ctx, "SELECT depth FROM categories WHERE id = $1", req.ParentId).Scan(&parentDepth)
        if err != nil {
            if err == sql.ErrNoRows {
                return nil, fmt.Errorf("parent category not found")
            }
            return nil, fmt.Errorf("failed to get parent category: %v", err)
        }
        depth = parentDepth + 1
    }

    query := `
        INSERT INTO categories (
            id,
            name,
            description,
            image,
            parent_id,
            depth,
            created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6,  CURRENT_TIMESTAMP
        )
        RETURNING id, name, description, image, parent_id, depth, created_at`

    var category pb.Category
    var createdAt sql.NullTime
    var parentID, image sql.NullString

    err = tx.QueryRowContext(ctx, query,
        categoryID,
        req.Name,
        req.Description,
        req.Image,
        req.ParentId,
        depth,
    ).Scan(
        &category.Id,
        &category.Name,
        &category.Description,
        &image,
        &parentID,
        &depth,
        &createdAt,
    )

    if err != nil {
        return nil, fmt.Errorf("failed to insert category: %v", err)
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %v", err)
    }

    // Handle optional fields
    if parentID.Valid {
        category.ParentId = &parentID.String
    }
    
    if image.Valid {
        category.Image = &image.String
    }
    if createdAt.Valid {
        category.CreatedAt = timestamppb.New(createdAt.Time)
    }

    return &category, nil
}

func (s *CategoryService) GetCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {

    query := `
        WITH RECURSIVE category_tree AS (
            SELECT 
                c.id, c.name, c.description, c.image,  
                c.parent_id, c.depth, c.created_at,
                ARRAY[]::VARCHAR[] AS path,
                0 as level
            FROM categories c
            WHERE c.parent_id IS NULL
            UNION ALL
            SELECT 
                c.id, c.name, c.description, c.image,
                c.parent_id, c.depth, c.created_at,
                path || c.parent_id,
                ct.level + 1
            FROM categories c
            INNER JOIN category_tree ct ON ct.id = c.parent_id
        )
        SELECT 
            id, name, description, image, 
            parent_id, depth, created_at,
            path
        FROM category_tree
        ORDER BY path, level`

    rows, err := s.DB.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query categories: %v", err)
    }
    defer rows.Close()

    // Map to store categories by ID for building the tree
    categoryMap := make(map[string]*pb.Category)
    var rootCategories []*pb.Category

    for rows.Next() {
        var cat pb.Category
        var createdAt sql.NullTime
        var parentID, image sql.NullString
        var path []string

        err := rows.Scan(
            &cat.Id,
            &cat.Name,
            &cat.Description,
            &image,
          
            &parentID,
            &cat.Depth,
            &createdAt,
            &path,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan category: %v", err)
        }

        // Handle optional fields
        if parentID.Valid {
            cat.ParentId = &parentID.String
        }
        
        if image.Valid {
            cat.Image = &image.String
        }
        if createdAt.Valid {
            cat.CreatedAt = timestamppb.New(createdAt.Time)
        }

        categoryMap[cat.Id] = &cat

        // If it's a root category (no parent)
        if !parentID.Valid {
            rootCategories = append(rootCategories, &cat)
        } else {
            // Add to parent's children
            parent := categoryMap[parentID.String]
            if parent != nil {
                parent.Children = append(parent.Children, &cat)
            }
        }
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating categories: %v", err)
    }

    return &pb.ListCategoriesResponse{
        Categories: rootCategories,
        Total:     int32(len(categoryMap)),
    }, nil
}

func (s *CategoryService) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
    // Set default values if not provided
    if req.Page <= 0 {
        req.Page = 1
    }
    if req.Limit <= 0 {
        req.Limit = 10
    }

    // First, get all categories without pagination
    query := `
        SELECT 
            id, 
            name, 
            description, 
            image, 
            depth,
            parent_id,
            created_at
        FROM categories
        WHERE 1=1
    `
    
    var args []interface{}
    
    // Add parent_id filter if provided to only get specific tree
    if req.ParentId != nil {
        query += " AND (id = $1 OR parent_id = $1)"
        args = append(args, *req.ParentId)
    }

    rows, err := s.DB.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query categories: %v", err)
    }
    defer rows.Close()

    // Store all categories in a map for easier lookup
    categoriesMap := make(map[string]*pb.Category)
    var rootCategories []*pb.Category

    // First pass: create all category objects
    for rows.Next() {
        var cat pb.Category
        var parentID, image sql.NullString
        var createdAt sql.NullTime

        err := rows.Scan(
            &cat.Id,
            &cat.Name,
            &cat.Description,
            &image,
            &cat.Depth,
            &parentID,
            &createdAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan category: %v", err)
        }

        if image.Valid {
            cat.Image = &image.String
        }
        if parentID.Valid {
            cat.ParentId = &parentID.String
        }
        if createdAt.Valid {
            cat.CreatedAt = timestamppb.New(createdAt.Time)
        }

        categoriesMap[cat.Id] = &cat
    }

    // Second pass: build the tree structure
    for _, cat := range categoriesMap {
        if cat.ParentId == nil || *cat.ParentId == "" {
            // This is a root category
            rootCategories = append(rootCategories, cat)
        } else {
            // This category has a parent
            if parent, exists := categoriesMap[*cat.ParentId]; exists {
                if parent.Children == nil {
                    parent.Children = make([]*pb.Category, 0)
                }
                parent.Children = append(parent.Children, cat)
            }
        }
    }

    // Apply pagination to root categories
    total := int32(len(rootCategories))
    start := (req.Page - 1) * req.Limit
    end := start + req.Limit
    if start >= int32(len(rootCategories)) {
        rootCategories = []*pb.Category{}
    } else {
        if end > int32(len(rootCategories)) {
            end = int32(len(rootCategories))
        }
        rootCategories = rootCategories[start:end]
    }

    return &pb.ListCategoriesResponse{
        Categories: rootCategories,
        Total:     total,
    }, nil
}


func (s *CategoryService) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
    // Build update query dynamically based on provided fields
    query := `UPDATE categories SET `
    
    updates := []string{}
    args := []interface{}{}
    argCount := 1

    if req.Name != nil {
        updates = append(updates, fmt.Sprintf("name = $%d", argCount))
        args = append(args, *req.Name)
        argCount++
    }

    if req.Description != nil {
        updates = append(updates, fmt.Sprintf("description = $%d", argCount))
        args = append(args, *req.Description)
        argCount++
    }

    if req.Image != nil {
        updates = append(updates, fmt.Sprintf("image = $%d", argCount))
        args = append(args, *req.Image)
        argCount++
    }

    if req.ParentId != nil {
        updates = append(updates, fmt.Sprintf("parent_id = $%d", argCount))
        args = append(args, *req.ParentId)
        argCount++
    }

    // If no fields to update
    if len(updates) == 0 {
        return nil, fmt.Errorf("no fields to update")
    }

    // Combine all updates
    query += strings.Join(updates, ", ")
    query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, description, image, depth, parent_id, created_at", argCount)
    args = append(args, req.Id)

    // Execute update and get updated row
    var category pb.Category
    var parentID, image sql.NullString
    var createdAt sql.NullTime

    err := s.DB.QueryRowContext(ctx, query, args...).Scan(
        &category.Id,
        &category.Name,
        &category.Description,
        &image,
        &category.Depth,
        &parentID,
        &createdAt,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("category not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to update category: %v", err)
    }

    // Handle nullable fields
    if image.Valid {
        category.Image = &image.String
    }
    if parentID.Valid {
        category.ParentId = &parentID.String
    }
    if createdAt.Valid {
        category.CreatedAt = timestamppb.New(createdAt.Time)
    }

    return &category, nil
}



func (s *CategoryService) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
    // Start a transaction
    tx, err := s.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %v", err)
    }
    defer tx.Rollback()

    // First check if category exists
    var exists bool
    err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", req.Id).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("failed to check category existence: %v", err)
    }
    if !exists {
        return nil, fmt.Errorf("category not found")
    }

    var deletedCount int64
    if req.DeleteChildren {
        // Delete the category and all its descendants
        withDescendantsQuery := `
            WITH RECURSIVE category_tree AS (
                -- Base case: the category we want to delete
                SELECT id FROM categories WHERE id = $1
                UNION ALL
                -- Recursive case: get all children
                SELECT c.id 
                FROM categories c
                INNER JOIN category_tree ct ON c.parent_id = ct.id
            )
            DELETE FROM categories 
            WHERE id IN (SELECT id FROM category_tree)
            RETURNING id`
        
        result, err := tx.ExecContext(ctx, withDescendantsQuery, req.Id)
        if err != nil {
            return nil, fmt.Errorf("failed to delete category and its children: %v", err)
        }
        deletedCount, err = result.RowsAffected()
        if err != nil {
            return nil, fmt.Errorf("failed to get affected rows: %v", err)
        }
    } else {
        // Check if category has children
        var hasChildren bool
        err = tx.QueryRowContext(ctx, 
            "SELECT EXISTS(SELECT 1 FROM categories WHERE parent_id = $1)", 
            req.Id,
        ).Scan(&hasChildren)
        if err != nil {
            return nil, fmt.Errorf("failed to check for children: %v", err)
        }
        if hasChildren {
            return nil, fmt.Errorf("cannot delete category with children, set DeleteChildren to true to delete all")
        }

        // Delete single category
        result, err := tx.ExecContext(ctx, 
            "DELETE FROM categories WHERE id = $1", 
            req.Id,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to delete category: %v", err)
        }
        deletedCount, err = result.RowsAffected()
        if err != nil {
            return nil, fmt.Errorf("failed to get affected rows: %v", err)
        }
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %v", err)
    }

    return &pb.DeleteCategoryResponse{
        Success: true,
        DeletedCount: deletedCount,
    }, nil
}