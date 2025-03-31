# Atriumn Storage SDK

The Atriumn Storage SDK provides a Go client for interacting with the Atriumn Storage service. It allows you to generate pre-signed URLs for uploading and downloading files from S3.

## Installation

```bash
go get github.com/atriumn/atriumn-sdk-go
```

## Usage

### Initializing the Client

```go
import (
    "github.com/atriumn/atriumn-sdk-go/storage"
)

// Create a client with default options
client, err := storage.NewClient("https://api.atriumn.com/storage")
if err != nil {
    // Handle error
}

// Create a client with custom options
client, err := storage.NewClientWithOptions("https://api.atriumn.com/storage",
    storage.WithHTTPClient(&http.Client{Timeout: 20 * time.Second}),
    storage.WithUserAgent("custom-user-agent/1.0"),
)
```

### Authentication

The storage service requires JWT authentication. You need to provide a token provider that implements the `TokenProvider` interface:

```go
// Example token provider using the auth package
type AuthTokenProvider struct {
    authClient *auth.Client
    clientID   string
    clientSecret string
}

func (a *AuthTokenProvider) GetToken(ctx context.Context) (string, error) {
    // Get token from the auth service
    tokenResp, err := a.authClient.GetClientCredentialsToken(ctx, a.clientID, a.clientSecret, "storage")
    if err != nil {
        return "", err
    }
    return tokenResp.AccessToken, nil
}

// Create a client with the token provider
tokenProvider := &AuthTokenProvider{
    authClient: authClient,
    clientID: "your-client-id",
    clientSecret: "your-client-secret",
}
client, err := storage.NewClientWithOptions("https://api.atriumn.com/storage",
    storage.WithTokenProvider(tokenProvider),
)
```

### Generating an Upload URL

```go
// Generate a pre-signed URL for uploading a file
uploadResp, err := client.GenerateUploadURL(ctx, &storage.GenerateUploadURLRequest{
    Filename:    "document.pdf",
    ContentType: "application/pdf",
})
if err != nil {
    // Handle error
}

// Use the uploadResp.UploadURL with the uploadResp.HTTPMethod (typically PUT)
// to upload the file to S3
```

### Generating a Download URL

```go
// Generate a pre-signed URL for downloading a file
downloadResp, err := client.GenerateDownloadURL(ctx, &storage.GenerateDownloadURLRequest{
    S3Key: "tenant-123/files/document.pdf",
})
if err != nil {
    // Handle error
}

// Use the downloadResp.DownloadURL with the downloadResp.HTTPMethod (typically GET)
// to download the file from S3
```

### Error Handling

The SDK returns errors that implement the standard error interface. For API errors, the error will be of type `*storage.ErrorResponse`:

```go
resp, err := client.GenerateDownloadURL(ctx, req)
if err != nil {
    if apiErr, ok := err.(*storage.ErrorResponse); ok {
        // Handle API error
        fmt.Printf("API Error: %s - %s\n", apiErr.ErrorCode, apiErr.Description)
    } else {
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Complete Example

Here's a complete example showing how to upload a file using the storage SDK:

```go
package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"

    "github.com/atriumn/atriumn-sdk-go/auth"
    "github.com/atriumn/atriumn-sdk-go/storage"
)

// Simple token provider that caches tokens
type CachingTokenProvider struct {
    authClient   *auth.Client
    clientID     string
    clientSecret string
    token        string
    expiresAt    time.Time
}

func (c *CachingTokenProvider) GetToken(ctx context.Context) (string, error) {
    // If we have a valid token, return it
    if c.token != "" && time.Now().Before(c.expiresAt) {
        return c.token, nil
    }

    // Otherwise, get a new token
    tokenResp, err := c.authClient.GetClientCredentialsToken(ctx, c.clientID, c.clientSecret, "storage")
    if err != nil {
        return "", err
    }

    // Cache the token with a buffer before expiry
    c.token = tokenResp.AccessToken
    c.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

    return c.token, nil
}

func main() {
    ctx := context.Background()

    // Initialize auth client
    authClient, err := auth.NewClient("https://api.atriumn.com/auth")
    if err != nil {
        fmt.Printf("Error creating auth client: %v\n", err)
        return
    }

    // Create token provider
    tokenProvider := &CachingTokenProvider{
        authClient:   authClient,
        clientID:     "your-client-id",
        clientSecret: "your-client-secret",
    }

    // Initialize storage client with token provider
    storageClient, err := storage.NewClientWithOptions("https://api.atriumn.com/storage",
        storage.WithTokenProvider(tokenProvider),
    )
    if err != nil {
        fmt.Printf("Error creating storage client: %v\n", err)
        return
    }

    // 1. Generate upload URL
    uploadResp, err := storageClient.GenerateUploadURL(ctx, &storage.GenerateUploadURLRequest{
        Filename:    "example.pdf",
        ContentType: "application/pdf",
    })
    if err != nil {
        fmt.Printf("Error generating upload URL: %v\n", err)
        return
    }

    fmt.Printf("Generated upload URL: %s\n", uploadResp.UploadURL)

    // 2. Upload file using the pre-signed URL
    file, err := os.Open("path/to/your/file.pdf")
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        return
    }
    defer file.Close()

    req, err := http.NewRequestWithContext(ctx, uploadResp.HTTPMethod, uploadResp.UploadURL, file)
    if err != nil {
        fmt.Printf("Error creating upload request: %v\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/pdf")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Printf("Error uploading file: %v\n", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        fmt.Printf("Upload failed with status %d: %s\n", resp.StatusCode, string(body))
        return
    }

    fmt.Println("File uploaded successfully!")

    // 3. Generate download URL (assuming we know the S3 key)
    s3Key := "tenant/files/example.pdf" // The actual S3 key might be generated by the service

    downloadResp, err := storageClient.GenerateDownloadURL(ctx, &storage.GenerateDownloadURLRequest{
        S3Key: s3Key,
    })
    if err != nil {
        fmt.Printf("Error generating download URL: %v\n", err)
        return
    }

    fmt.Printf("Generated download URL: %s\n", downloadResp.DownloadURL)
}
```
