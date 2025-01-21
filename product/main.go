package main

import (
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/wafi04/go-testing/common/pkg/logger"
	"github.com/wafi04/go-testing/product/database"
	pb "github.com/wafi04/go-testing/product/grpc"
	"github.com/wafi04/go-testing/product/handler"
	"github.com/wafi04/go-testing/product/service"
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


	productService := service.NewProductService(db.DB)
    categoryHandler := handler.NewProductHandler(productService)

	grpcServer := grpc.NewServer()

    pb.RegisterProductServiceServer(grpcServer, categoryHandler)

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