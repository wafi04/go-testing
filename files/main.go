package main

import (
	"context"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/admin/search"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)



func main(){

	var cld, err = cloudinary.New()
    if err != nil {
        log.Fatalf("Failed to initialize Cloudinary, %v", err)
    }
	 var ctx = context.Background()
    // Upload an image to your Cloudinary product environment from a specified URL.
    //
    // Alternatively you can provide a path to a local file on your filesystem,
    // base64 encoded string, io.Reader and more.
    //
    // For additional information see:
    // https://cloudinary.com/documentation/upload_images
    //
    // Upload can be greatly customized by specifying uploader.UploadParams,
    // in this case we set the Public ID of the uploaded asset to "logo".
    uploadResult, err := cld.Upload.Upload(
        ctx,
        "https://res.cloudinary.com/demo/image/upload/v1598276026/docs/models.jpg",
        uploader.UploadParams{PublicID: "models",
            UniqueFilename: api.Bool(false),
            Overwrite:      api.Bool(true)})
    if err != nil {
        log.Fatalf("Failed to upload file, %v\n", err)
    }
    log.Println(uploadResult.SecureURL)
    // Prints something like:
    // https://res.cloudinary.com/<your cloud name>/image/upload/v1615875158/models.png
    // uploadResult contains useful information about the asset, like Width, Height, Format, etc.
    // See uploader.UploadResult struct for more details.
    // Now we can use Admin API to see the details about the asset.
    // The request can be customised by providing AssetParams.
    asset, err := cld.Admin.Asset(ctx, admin.AssetParams{PublicID: "models"})
    if err != nil {
        log.Fatalf("Failed to get asset details, %v\n", err)
    }
    // Print some basic information about the asset.
    log.Printf("Public ID: %v, URL: %v\n", asset.PublicID, asset.SecureURL)
    // Cloudinary also provides a very flexible Search API for filtering and retrieving
    // information on all the assets in your product environment with the help of query expressions
    // in a Lucene-like query language.
    searchQuery := search.Query{
        Expression: "resource_type:image AND uploaded_at>1d AND bytes<1m",
        SortBy:     []search.SortByField{{"created_at": search.Descending}},
        MaxResults: 30,
    }
    searchResult, err := cld.Admin.Search(ctx, searchQuery)
    if err != nil {
        log.Fatalf("Failed to search for assets, %v\n", err)
    }
    log.Printf("Assets found: %v\n", searchResult.TotalCount)
    for _, asset := range searchResult.Assets {
        log.Printf("Public ID: %v, URL: %v\n", asset.PublicID, asset.SecureURL)
    }

}