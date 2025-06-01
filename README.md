# Atriumn SDK for Go

The official Go SDK for Atriumn services, providing idiomatic Go clients for authentication, storage, AI, and content ingestion.

## Documentation Index

- **[APPROACH.md](APPROACH.md)** - Technical strategy, patterns, and development decisions
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design, components, and data flows
- **[docs/TESTING.md](docs/TESTING.md)** - Testing strategy and execution guide
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Distribution and integration instructions
- **[TEST_AUDIT_REPORT.md](TEST_AUDIT_REPORT.md)** - Current testing status and coverage
- **[examples/](examples/)** - Working code examples for all services

## Prerequisites

- **Go 1.24.0 or later**
- Network access to Atriumn API endpoints
- Valid Atriumn API credentials

## Getting Started

### Installation

Install individual service clients as needed:

```bash
# Install the auth client
go get github.com/atriumn/atriumn-sdk-go/auth

# Install the storage client  
go get github.com/atriumn/atriumn-sdk-go/storage

# Install the AI client
go get github.com/atriumn/atriumn-sdk-go/ai

# Install the ingest client
go get github.com/atriumn/atriumn-sdk-go/ingest

# Or install all at once
go get github.com/atriumn/atriumn-sdk-go/...
```

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/atriumn/atriumn-sdk-go/auth"
    "github.com/atriumn/atriumn-sdk-go/storage"
)

func main() {
    ctx := context.Background()
    
    // Initialize auth client
    authClient := auth.NewClient("your-api-key")
    
    // Authenticate user
    loginResp, err := authClient.Login(ctx, auth.LoginRequest{
        Email:    "user@example.com", 
        Password: "password",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize storage client with JWT token
    storageClient := storage.NewClient(loginResp.Token)
    
    // Generate upload URL
    uploadResp, err := storageClient.GenerateUploadURL(ctx, storage.GenerateUploadURLRequest{
        Filename: "document.pdf",
        Size:     1024000,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Upload URL: %s\n", uploadResp.UploadURL)
}
```

## Available SDKs

### [Auth Client](auth/)

Authentication service client providing:
- Client credentials authentication  
- User signup, login, and management
- Password reset workflows
- JWT token management

### [Storage Client](storage/)

Storage service client providing:
- Pre-signed S3 URL generation for uploads
- Pre-signed S3 URL generation for downloads  
- JWT token-based authentication
- File metadata management

### [AI Client](ai/)

AI service client providing:
- Prompt creation and management
- AI model configuration
- Request/response handling
- Version control for prompts

### [Ingest Client](ingest/)

Content ingestion client providing:
- Text content ingestion
- URL content extraction
- File processing workflows
- Bulk ingestion operations

## Basic Usage Examples

### Authentication

```go
authClient := auth.NewClient("your-api-key")

// User signup
signupResp, err := authClient.Signup(ctx, auth.SignupRequest{
    Email:    "user@example.com",
    Password: "secure-password",
    Name:     "John Doe",
})

// User login  
loginResp, err := authClient.Login(ctx, auth.LoginRequest{
    Email:    "user@example.com",
    Password: "secure-password", 
})
```

### Storage Operations

```go
storageClient := storage.NewClient("jwt-token")

// Generate upload URL
uploadResp, err := storageClient.GenerateUploadURL(ctx, storage.GenerateUploadURLRequest{
    Filename: "document.pdf",
    Size:     1024000,
})

// Generate download URL
downloadResp, err := storageClient.GenerateDownloadURL(ctx, storage.GenerateDownloadURLRequest{
    FileID: "file-id-123",
})
```

### AI Operations

```go
aiClient := ai.NewClient("jwt-token")

// Create prompt
promptResp, err := aiClient.CreatePrompt(ctx, ai.CreatePromptRequest{
    Name:    "document-summary",
    Content: "Summarize the following document: {{content}}",
    Model:   "gpt-4",
})

// Update prompt
updateResp, err := aiClient.UpdatePrompt(ctx, "prompt-id", ai.UpdatePromptRequest{
    Content: "Provide a detailed summary of: {{content}}",
})
```

### Content Ingestion

```go
ingestClient := ingest.NewClient("jwt-token")

// Ingest text content
textResp, err := ingestClient.IngestText(ctx, ingest.IngestTextRequest{
    Content:  "This is the content to ingest",
    Metadata: map[string]string{"source": "api"},
})

// Ingest from URL
urlResp, err := ingestClient.IngestURL(ctx, ingest.IngestURLRequest{
    URL:      "https://example.com/document.html",
    Metadata: map[string]string{"source": "web"},
})
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output  
make test-verbose

# Run tests with coverage
make test-coverage

# Run specific package tests
go test -v ./auth
go test -v ./storage  
go test -v ./ai
go test -v ./ingest
```

### Code Quality

```bash
# Run linters
make lint

# Format code
make fmt

# Run test audit
make test-audit
```

### Building Examples

```bash
# Build all examples
cd examples && go build ./...

# Run specific example
cd examples && go run file_upload.go
```

## Distribution

The SDK is distributed as a Go module via the standard Go module system. Release versions follow semantic versioning (e.g., v1.2.3).

### Release Process

```bash
# Tag new patch version
make tag-patch

# Tag new minor version  
make tag-minor

# Tag new major version
make tag-major

# Push tags
git push origin <tag-name>
```

## License

[MIT](LICENSE)
