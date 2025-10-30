package utils

import (
	"context"

	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(filePath string, folder string) (string, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)

	if err != nil {
		log.Println("Cloudinary init err", err)
		return "", err
	}

	uploadParam := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "image",
	}

	uploadResult, err := cld.Upload.Upload(context.Background(), filePath, uploadParam)

	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}
