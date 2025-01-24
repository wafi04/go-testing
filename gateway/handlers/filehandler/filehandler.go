package filehandler

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/wafi04/files/service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Filehandler struct {
	fileClient pb.FileServiceClient
}

func NewFileGateway(ctx context.Context)(*Filehandler,error){
	conn, err := grpc.DialContext(ctx,
            "192.168.100.81:5054",
            grpc.WithTransportCredentials(insecure.NewCredentials()),
            grpc.WithBlock(),
    )

	if err != nil {
		return nil, err
	}

	return &Filehandler{
		fileClient: pb.NewFileServiceClient(conn),
	},nil

}


func (s *Filehandler) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	// Pastikan method adalah POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // Maksimal 10 MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Ambil file dari request
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Baca isi file
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Siapkan context dengan timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	PublicID :=  fmt.Sprintf("%06d",rand.Intn(1000000))

	uploadRequest := &pb.FileUploadRequest{
		FileData: fileBytes,
		Folder:   "testing", 
		PublicId: PublicID,
	}

	response, err := s.fileClient.UploadFile(ctx, uploadRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"url": "%s", "public_id": "%s"}`, 
		response.Url, 
		response.PublicId,
	)))
}
