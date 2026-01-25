package utils

import (
	"context"

	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(filePath string, folder string) (string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)

	if err != nil {
		log.Println("Cloudinary init err", err)
		return "", err
	}

	uploadParam := uploader.UploadParams{
		Folder:         folder,
		ResourceType:   "image",
		AllowedFormats: []string{"jpg", "png", "jpeg"},
		Transformation: "w_1280,q_auto",
	}

	// Open the file to ensure we are sending the content, not a path string that might be misinterpreted
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file for cloudinary upload: %v", err)
		return "", err
	}
	defer file.Close()

	uploadResult, err := cld.Upload.Upload(context.Background(), file, uploadParam)

	if err != nil {
		log.Printf("Cloudinary Upload Error: %v", err)
		return "", err
	}

	// Log the full result for debugging
	// log.Printf("Cloudinary Upload Result ID: %s, PublicID: %s, URL: %s, SecureURL: %s", uploadResult.AssetID, uploadResult.PublicID, uploadResult.URL, uploadResult.SecureURL)

	if uploadResult.SecureURL == "" {
		// log.Printf("[Error] Cloudinary returned empty SecureURL. Full Result: %+v", uploadResult)
		if uploadResult.URL != "" {
			return uploadResult.URL, nil // Fallback to non-secure URL if available
		}
	}

	return uploadResult.SecureURL, nil
}
