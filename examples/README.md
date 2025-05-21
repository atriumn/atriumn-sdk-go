# Atriumn SDK Go Examples

This directory contains example code demonstrating how to use the Atriumn SDK for Go.

## File Upload Example

The file upload example demonstrates the new two-step file upload process introduced in the Atriumn SDK.

### How It Works

The two-step file upload process works as follows:

1. **Request Upload URL**: The client sends a request to the Atriumn Ingest API with metadata about the file to be uploaded. The API returns a pre-signed S3 URL.

2. **Upload Content**: The client uploads the file content directly to the pre-signed S3 URL using an HTTP PUT request.

This approach has several advantages:

- Reduced server load, as file contents bypass the API server
- Improved upload performance
- Support for larger file sizes
- Better error handling and retry capabilities

### Example Code

```go
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
	// Initialize the client
	client, err := ingest.NewClient("https://api.atriumn.io")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Open the file
	filePath := "document.pdf"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Step 1: Request a file upload URL
	uploadRequest := &ingest.RequestFileUploadRequest{
		Filename:    filepath.Base(filePath),
		ContentType: "application/pdf",
		TenantID:    "your-tenant-id", // Optional
		UserID:      "your-user-id",   // Optional
		Metadata: map[string]string{   // Optional
			"source": "sdk-example",
		},
	}

	uploadResponse, err := client.RequestFileUpload(context.Background(), uploadRequest)
	if err != nil {
		log.Fatalf("Failed to request upload URL: %v", err)
	}

	fmt.Printf("Content ID: %s, Upload URL obtained\n", uploadResponse.ContentID)

	// Step 2: Upload the file content directly to the pre-signed URL
	resp, err := client.UploadToURL(context.Background(), uploadResponse.UploadURL, "application/pdf", file)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Upload complete! Status: %d\n", resp.StatusCode)
}
```

### Running the Example

To run the file upload example:

```bash
go run file_upload.go path/to/your/file.pdf
```

## Notes on Migration

The old `IngestFile` method is now deprecated in favor of the new two-step process. If you are currently using the old method, you should migrate to the new approach to benefit from the improved performance and reliability.

The old method will continue to work for backward compatibility but may be removed in future versions of the SDK.
