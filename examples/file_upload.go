package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/atriumn/atriumn-sdk-go/ingest"
)

func main() {
	// Check if a filename was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run file_upload.go <file_to_upload>")
		os.Exit(1)
	}

	// Get the file path from command line arguments
	filePath := os.Args[1]

	// Initialize the ingest client
	client, err := ingest.NewClient("https://api.atriumn.io")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}

	// Determine content type based on file extension
	// In a production application, you might want to use a more robust method
	// such as http.DetectContentType or a library like github.com/gabriel-vasile/mimetype
	contentType := getContentType(filePath)

	fmt.Printf("Uploading file: %s (Size: %d bytes, Type: %s)\n", filepath.Base(filePath), fileInfo.Size(), contentType)

	// Step 1: Request a file upload URL
	uploadRequest := &ingest.RequestFileUploadRequest{
		Filename:    filepath.Base(filePath),
		ContentType: contentType,
		// Add tenant ID and user ID if required by your application
		// TenantID: "your-tenant-id",
		// UserID: "your-user-id",
		Metadata: map[string]string{
			"source": "sdk-example",
			"size":   fmt.Sprintf("%d", fileInfo.Size()),
		},
	}

	uploadResponse, err := client.RequestFileUpload(context.Background(), uploadRequest)
	if err != nil {
		log.Fatalf("Failed to request upload URL: %v", err)
	}

	fmt.Printf("Upload URL obtained. Content ID: %s, Status: %s\n", uploadResponse.ContentID, uploadResponse.Status)

	// Rewind the file to the beginning for upload
	if _, err := file.Seek(0, 0); err != nil {
		log.Fatalf("Failed to seek to beginning of file: %v", err)
	}

	// Step 2: Upload the file content directly to the pre-signed URL
	fmt.Println("Uploading file content to the pre-signed URL...")
	resp, err := client.UploadToURL(context.Background(), uploadResponse.UploadURL, contentType, file)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Upload complete! Status: %d\n", resp.StatusCode)
	fmt.Printf("Content item ID: %s\n", uploadResponse.ContentID)
}

// getContentType returns a MIME type based on the file extension
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".mp4":
		return "video/mp4"
	default:
		// Default to binary data if we can't determine the type
		return "application/octet-stream"
	}
} 