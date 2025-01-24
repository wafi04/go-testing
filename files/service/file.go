package service

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/wafi04/files/service/pb"
)

type Cloudinary struct {
	pb.UnimplementedFileServiceServer
    cloudinary *cloudinary.Cloudinary
}

func NewCloudinaryService(cld *cloudinary.Cloudinary) *Cloudinary {
    return &Cloudinary{cloudinary: cld}
}

func (s *Cloudinary) UploadFile(
    ctx context.Context, 
    req *pb.FileUploadRequest,
) (*pb.FileUploadResponse, error) {
    tempFile, err := ioutil.TempFile("", "upload-*")
    if err != nil {
        return nil, err
    }
    defer os.Remove(tempFile.Name())

    if _, err := tempFile.Write(req.FileData); err != nil {
        return nil, err
    }
    tempFile.Close()


    uploadResult, err := s.cloudinary.Upload.Upload(ctx, tempFile.Name(), uploader.UploadParams{
        Folder: req.Folder,
        PublicID: req.PublicId,
    })
    if err != nil {
        return nil, err
    }

    return &pb.FileUploadResponse{
        Url: uploadResult.SecureURL,
        PublicId: uploadResult.PublicID,
    }, nil
}