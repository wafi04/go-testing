package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	pb "github.com/wafi04/go-testing/category/grpc"
)

func TestCreateCategory(t *testing.T) {
    // Buat mock DB
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("failed to create mock db: %v", err)
    }
    defer db.Close()

    service := NewCategoryService(db)

    // Setup test cases
    tests := []struct {
        name        string
        request     *pb.CreateCategoryRequest
        mockSetup   func(mock sqlmock.Sqlmock)
        expectError bool
    }{
        {
            name: "successful creation without parent",
            request: &pb.CreateCategoryRequest{
                Name:        "Test Category",
                Description: "Test Description",
                Image:      nil,
                ParentId:   nil,
            },
            mockSetup: func(mock sqlmock.Sqlmock) {
                // Expect begin transaction
                mock.ExpectBegin()

                // Expect insert query
                mock.ExpectQuery(regexp.QuoteMeta(`
                    INSERT INTO categories (
                        id,
                        name,
                        description,
                        image,
                        parent_id,
                        depth,
                        created_at
                    ) VALUES (
                        $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP
                    )
                    RETURNING id, name, description, image, parent_id, depth, created_at`)).
                    WithArgs(
                        sqlmock.AnyArg(),
                        "Test Category",
                        "Test Description",
                        nil,
                        nil,
                        int32(0),
                    ).
                    WillReturnRows(
                        sqlmock.NewRows([]string{
                            "id",
                            "name",
                            "description",
                            "image",
                            "parent_id",
                            "depth",
                            "created_at",
                        }).
                            AddRow(
                                "test-uuid",
                                "Test Category",
                                "Test Description",
                                nil,
                                nil,
                                0,
                                time.Now(),
                            ),
                    )

                // Expect commit
                mock.ExpectCommit()
            },
            expectError: false,
        },
        {
            name: "successful creation with parent",
            request: &pb.CreateCategoryRequest{
                Name:        "Child Category",
                Description: "Child Description",
                ParentId:    stringPtr("parent-uuid"),
            },
            mockSetup: func(mock sqlmock.Sqlmock) {
                // Expect begin transaction
                mock.ExpectBegin()

                // Expect parent depth query
                mock.ExpectQuery(regexp.QuoteMeta(`SELECT depth FROM categories WHERE id = $1`)).
                    WithArgs("parent-uuid").
                    WillReturnRows(sqlmock.NewRows([]string{"depth"}).AddRow(1))

                // Expect insert query
                mock.ExpectQuery(regexp.QuoteMeta(`
                    INSERT INTO categories (
                        id,
                        name,
                        description,
                        image,
                        parent_id,
                        depth,
                        created_at
                    ) VALUES (
                        $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP
                    )
                    RETURNING id, name, description, image, parent_id, depth, created_at`)).
                    WithArgs(
                        sqlmock.AnyArg(),
                        "Child Category",
                        "Child Description",
                        nil,
                        "parent-uuid",
                        int32(2),
                    ).
                    WillReturnRows(
                        sqlmock.NewRows([]string{
                            "id",
                            "name",
                            "description",
                            "image",
                            "parent_id",
                            "depth",
                            "created_at",
                        }).
                            AddRow(
                                "child-uuid",
                                "Child Category",
                                "Child Description",
                                nil,
                                "parent-uuid",
                                2,
                                time.Now(),
                            ),
                    )

                // Expect commit
                mock.ExpectCommit()
            },
            expectError: false,
        },
        {
            name: "parent category not found",
            request: &pb.CreateCategoryRequest{
                Name:        "Child Category",
                Description: "Child Description",
                ParentId:    stringPtr("non-existent-uuid"),
            },
            mockSetup: func(mock sqlmock.Sqlmock) {
                // Expect begin transaction
                mock.ExpectBegin()

                // Expect parent depth query to return no rows
                mock.ExpectQuery(regexp.QuoteMeta(`SELECT depth FROM categories WHERE id = $1`)).
                    WithArgs("non-existent-uuid").
                    WillReturnError(sql.ErrNoRows)

                // Expect rollback
                mock.ExpectRollback()
            },
            expectError: true,
        },
    }

    // Run test cases
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock expectations
            tt.mockSetup(mock)

            // Execute test
            category, err := service.CreateCategory(context.Background(), tt.request)

            // Verify results
            if tt.expectError {
                assert.Error(t, err)
                assert.Nil(t, category)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, category)
                assert.Equal(t, tt.request.Name, category.Name)
                assert.Equal(t, tt.request.Description, category.Description)
                if tt.request.ParentId != nil {
                    assert.Equal(t, *tt.request.ParentId, *category.ParentId)
                }
            }

            // Verify all expectations were met
            if err := mock.ExpectationsWereMet(); err != nil {
                t.Errorf("unfulfilled expectations: %s", err)
            }
        })
    }
}

// Helper function untuk membuat pointer ke string
func stringPtr(s string) *string {
    return &s
}