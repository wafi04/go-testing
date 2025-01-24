package main

import (
	"log"
	"net"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/joho/godotenv"
	"github.com/wafi04/files/service"
	"github.com/wafi04/files/service/pb"
	"google.golang.org/grpc"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    cld, err := cloudinary.NewFromParams(
        os.Getenv("CLOUDINARY_CLOUD_NAME"),
        os.Getenv("CLOUDINARY_API_KEY"),
        os.Getenv("CLOUDINARY_API_SECRET"),
    )
	
    if err != nil {
        log.Fatalf("Failed to initialize Cloudinary: %v", err)
    }

	grpcServer := grpc.NewServer()
    fileService := service.NewCloudinaryService(cld)

    // cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
    // publicID := "cld-sample"
    
    // url := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", cloudName, publicID)
    
    // urlWithTransform := fmt.Sprintf(
    //     "https://res.cloudinary.com/%s/image/upload/c_scale,w_500/%s", 
    //     cloudName, 
    //     publicID,
    // )

    // fmt.Println("Image URL:", url)
    // fmt.Println("Transformed URL:", urlWithTransform)
	 // Daftarkan service
    pb.RegisterFileServiceServer(grpcServer, fileService)

    // Mendengarkan pada port tertentu
    lis, err := net.Listen("tcp", ":50054")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    log.Println("Server is running on port :50054")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}