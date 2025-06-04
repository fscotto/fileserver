package service

import (
	"context"
	"fileserver/config"
	"fmt"
	"github.com/minio/minio-go/v7"
	"os"
)

// GetFileFromMinIO retrieves a file from the specified MinIO bucket.
// It returns the file object if found, or an error if there is an issue with fetching the file.
//
// Parameters:
// - bucketName (string): The name of the MinIO bucket to fetch the file from.
// - objectName (string): The name of the object (file) to retrieve from the bucket.
//
// Returns:
// - *minio.Object: The file object retrieved from MinIO.
// - error: An error is returned if there is an issue fetching the object from MinIO.
func GetFileFromMinIO(bucketName, objectName string) (*minio.Object, error) {
	// Fetch the object from MinIO using the provided bucket name and object name
	object, err := config.MinIO.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		// Return error if there is any issue in fetching the object
		return nil, fmt.Errorf("error getting object from MinIO: %v", err)
	}
	// Return the fetched object if successful
	return object, nil
}

// UploadFileToMinIO uploads a file to MinIO under the specified bucket and object name.
// If the bucket does not exist, it is created first.
//
// Parameters:
// - ctx (context.Context): The context for the operation (to control request lifetime).
// - bucketName (string): The name of the MinIO bucket to upload the file to.
// - objectName (string): The name of the object (file) in MinIO.
// - filePath (string): The local file path of the file to upload.
//
// Returns:
// - error: An error is returned if there is any issue during file upload.
func UploadFileToMinIO(ctx context.Context, bucketName, objectName, filePath string) error {
	// Open the file from the given file path
	file, err := os.Open(filePath)
	if err != nil {
		// Return error if unable to open the file
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close() // Ensure file is closed after use

	// Check if the bucket exists, create it if not
	if err = createBucketIfNotExists(ctx, bucketName); err != nil {
		// Return error if bucket creation fails
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	// Upload the file to MinIO
	_, err = config.MinIO.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		// Return error if uploading the file fails
		return fmt.Errorf("failed to upload file: %v", err)
	}
	// Return nil if the file is successfully uploaded
	return nil
}

// createBucketIfNotExists checks if the specified bucket exists and creates it if it doesn't.
// It is called by `UploadFileToMinIO` to ensure the bucket is available before uploading.
//
// Parameters:
// - ctx (context.Context): The context for the operation (to control request lifetime).
// - bucketName (string): The name of the bucket to check/create.
//
// Returns:
// - error: An error is returned if the bucket checking or creation process fails.
func createBucketIfNotExists(ctx context.Context, bucketName string) error {
	// Check if the bucket already exists
	exists, err := config.MinIO.BucketExists(ctx, bucketName)
	if err != nil {
		// Return error if checking the bucket existence fails
		return fmt.Errorf("failed to check if bucket exists: %v", err)
	}

	// If the bucket doesn't exist, create it
	if !exists {
		fmt.Println("Bucket does not exist. Creating bucket...")
		// Create the bucket with the specified region
		err = config.MinIO.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			// Return error if bucket creation fails
			return fmt.Errorf("failed to create bucket: %v", err)
		}
		fmt.Println("Bucket created successfully!")
	}
	// Return nil if bucket already exists or is created successfully
	return nil
}

// DeleteFileFromMinIO removes a file from the specified MinIO bucket.
//
// Parameters:
// - ctx (context.Context): The context for the operation (to control request lifetime).
// - bucketName (string): The name of the MinIO bucket where the file is stored.
// - objectName (string): The name of the object (file) to delete.
//
// Returns:
// - error: An error is returned if there is an issue deleting the file from MinIO.
func DeleteFileFromMinIO(ctx context.Context, bucketName, objectName string) error {
	// Remove the object from the MinIO bucket
	err := config.MinIO.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		// Return error if deleting the object fails
		return fmt.Errorf("error deleting object from MinIO: %v", err)
	}
	// Log success message after deletion
	fmt.Println("File deleted successfully")
	// Return nil if file is deleted successfully
	return nil
}
