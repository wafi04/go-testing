package main

import (
	"net"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver

	"github.com/wafi04/go-testing/category/database"
	pb "github.com/wafi04/go-testing/category/grpc"
	"github.com/wafi04/go-testing/category/handler"
	"github.com/wafi04/go-testing/category/service"
	"github.com/wafi04/go-testing/common/pkg/logger"
	"google.golang.org/grpc"
)
const (
	port = ":50053"
)
func main(){
	log :=  logger.NewLogger()
	db ,err :=  database.New()
		if err != nil {
		log.Log(logger.ErrorLevel, "Failed to initialize database : %v: ", err)
	}
	defer db.Close()

	health := db.Health()
	log.Log(logger.InfoLevel, "Database health : %v", health["status"])


	categoryService := service.NewCategoryService(db.DB)
    categoryHandler := handler.NewCategoryHandler(categoryService)

	grpcServer := grpc.NewServer()

    pb.RegisterCategoryServiceServer(grpcServer, categoryHandler)

    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Log(logger.ErrorLevel, "Failed to listen: %v", err)
        return
    }

    log.Log(logger.InfoLevel, "gRPC server starting on port %s", port)

    if err := grpcServer.Serve(lis); err != nil {
        log.Log(logger.ErrorLevel, "Failed to serve: %v", err)
        return
    }

}