package util

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"os"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// InitializeGCP initializes GCP and creates a bucket if it doesn't exist
func InitializeGCP(projectID, bucketName, creds string) (*storage.Client, error) {
	// Initialize the GCP client
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP client: %v", err)
	}
	return client, nil
}

func SaveSRTToBucket(ctx context.Context, bucket *storage.BucketHandle, objectPath string, srtContent string) (string, error) {
	obj := bucket.Object(objectPath)
	writer := obj.NewWriter(ctx)
	_, err := writer.Write([]byte(srtContent))
	if err != nil {
		return "", fmt.Errorf("error writing SRT to bucket: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	// Make the object publicly accessible
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("error setting ACL: %v", err)
	}

	// Get the public URL
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting object attributes: %v", err)
	}

	return attrs.MediaLink, nil
}

func DeleteFolderFromBucket(ctx context.Context, client *storage.Client, bucketName, folderName string) error {
	bucket := client.Bucket(bucketName)

	if err := bucket.Object(folderName).Delete(ctx); err != nil {
		log.Printf("failed to delete folder: %v", err)
		return err
	}
	return nil
}

func GetGCPClient() (*storage.Client, error) {
	creds := "/Users/aditya/Documents/OSS/zappush/shortpro/backend/gcp_credentials.json"
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP client: %v", err)
	}
	return client, nil
}

func DownloadFile(ctx context.Context, storageClient *storage.Client, bucketName string, objectName string, destPath string) error {
	bucket := storageClient.Bucket(bucketName)

	src := bucket.Object(objectName)
	reader, err := src.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, reader)
	return err
}

// UploadImageToGCP uploads a base64 encoded image to GCP and returns the URL.
func UploadImageToGCP(client *storage.Client, bucketName, objectName, base64Image string) (string, error) {
	log.Printf("Uploading image to GCP bucket: %s", bucketName)

	b64data := base64Image[strings.IndexByte(base64Image, ',')+1:]

	size, err := CalculateBase64ImageSizeMB(b64data)
	if err != nil {
		return "", fmt.Errorf("failed to calculate image size: %v", err)
	}

	log.Printf("Image size: %.2f MB", size)

	if size > 5 {
		return "", fmt.Errorf("image size exceeds 5 MB")
	}

	// Decode the base64 image
	imageData, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %v", err)
	}

	// save object in /campaign_images folder
	objectName = "campaign_images/" + objectName

	// Upload the image to the bucket
	ctx := context.Background()
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	writer := object.NewWriter(ctx)
	if _, err := writer.Write(imageData); err != nil {
		return "", fmt.Errorf("failed to write image to bucket: %v", err)
	}
	if err := writer.Close(); err != nil {
		// return "", fmt.Errorf("failed to close writer: %v", err)
		log.Printf("failed to close writer: %v", err)
		// try to make image public anyway
	}

	// Make the image publicly accessible
	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set object ACL: %v", err)
	}

	// Return the image URL
	imageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
	return imageURL, nil
}
