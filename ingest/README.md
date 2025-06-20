# Atriumn Ingest Client

A Go client library for the Atriumn Ingest Service. This package provides an easy way to interact with the Atriumn Ingest API from other Go services.

## Installation

```bash
go get github.com/atriumn/atriumn-sdk-go/ingest
```

## Usage

### Creating a Client

```go
import (
    "context"
    "fmt"
    "log"
    "time"
    "os"

    "github.com/atriumn/atriumn-sdk-go/auth"
    "github.com/atriumn/atriumn-sdk-go/ingest"
)

func main() {
    // Create a client with default settings
    client, err := ingest.NewClient("https://api.example.com")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Or create with custom options
    customClient, err := ingest.NewClientWithOptions(
        "https://api.example.com",
        ingest.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
        ingest.WithUserAgent("my-service/1.0"),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Use the client...
}
```

### Authentication

The ingest service requires JWT authentication. You need to provide a token provider that implements the `TokenProvider` interface:

```go
// Example token provider using the auth package
type AuthTokenProvider struct {
    authClient   *auth.Client
    clientID     string
    clientSecret string
}

func (a *AuthTokenProvider) GetToken(ctx context.Context) (string, error) {
    // Get token from the auth service
    tokenResp, err := a.authClient.GetClientCredentialsToken(ctx, a.clientID, a.clientSecret, "ingest")
    if err != nil {
        return "", err
    }
    return tokenResp.AccessToken, nil
}

// Create a client with the token provider
authClient, _ := auth.NewClient("https://auth.example.com")
tokenProvider := &AuthTokenProvider{
    authClient:   authClient,
    clientID:     "your-client-id",
    clientSecret: "your-client-secret",
}

client, err := ingest.NewClientWithOptions(
    "https://api.example.com",
    ingest.WithTokenProvider(tokenProvider),
)
```

### Ingesting Text

```go
ctx := context.Background()

// Create the request
request := &ingest.IngestTextRequest{
    TenantID: "tenant-123",
    UserID:   "user-456",
    Content:  "This is the text content to be ingested.",
    Metadata: map[string]string{
        "source": "manual-input",
        "category": "notes",
    },
}

// Send the request
response, err := client.IngestText(ctx, request)
if err != nil {
    log.Fatalf("Failed to ingest text: %v", err)
}

fmt.Printf("Ingestion ID: %s\n", response.ID)
fmt.Printf("Status: %s\n", response.Status)
```

### Ingesting URL

```go
ctx := context.Background()

// Create the request
request := &ingest.IngestURLRequest{
    TenantID: "tenant-123",
    UserID:   "user-456",
    URL:      "https://example.com/document.pdf",
    Metadata: map[string]string{
        "source": "web",
        "category": "document",
    },
}

// Send the request
response, err := client.IngestURL(ctx, request)
if err != nil {
    log.Fatalf("Failed to ingest URL: %v", err)
}

fmt.Printf("Ingestion ID: %s\n", response.ID)
fmt.Printf("Status: %s\n", response.Status)
```

### Uploading Files (Two-Step Process)

The SDK uses a two-step process for file uploads:

1. Request a pre-signed URL by sending file metadata
2. Upload the file content directly to the pre-signed URL

```go
ctx := context.Background()

// Open the file
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatalf("Failed to open file: %v", err)
}
defer file.Close()

// Step 1: Request a file upload URL
uploadRequest := &ingest.RequestFileUploadRequest{
    Filename:    "document.pdf",
    ContentType: "application/pdf",
    TenantID:    "tenant-123",
    UserID:      "user-456",
    Metadata: map[string]string{
        "source": "local-drive",
        "category": "document",
    },
}

uploadResponse, err := client.RequestFileUpload(ctx, uploadRequest)
if err != nil {
    log.Fatalf("Failed to request file upload URL: %v", err)
}

fmt.Printf("Content ID: %s\n", uploadResponse.ContentID)
fmt.Printf("Upload URL: %s\n", uploadResponse.UploadURL)

// Ensure file cursor is at the beginning
if _, err := file.Seek(0, 0); err != nil {
    log.Fatalf("Failed to seek to beginning of file: %v", err)
}

// Step 2: Upload the file directly to the pre-signed URL
resp, err := client.UploadToURL(ctx, uploadResponse.UploadURL, "application/pdf", file)
if err != nil {
    log.Fatalf("Failed to upload file: %v", err)
}
defer resp.Body.Close()

fmt.Printf("Upload status: %d\n", resp.StatusCode)
```

### Error Handling

The SDK returns errors that implement the standard error interface. For API errors, the error will be of type `*ingest.ErrorResponse`:

```go
resp, err := client.IngestText(ctx, request)
if err != nil {
    if apiErr, ok := err.(*ingest.ErrorResponse); ok {
        // Handle API error
        fmt.Printf("API Error: %s - %s\n", apiErr.ErrorCode, apiErr.Description)
    } else {
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```
