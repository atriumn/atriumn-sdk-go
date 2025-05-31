# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Test

```bash
# Run all tests
go test ./...

# Run tests for a specific package 
go test ./auth
go test ./storage
go test ./ai
go test ./ingest

# Run tests with verbose output
go test -v ./...

# Run tests with coverage reporting
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint code
golangci-lint run ./...

# Format code
find . -name "*.go" -not -path "./vendor/*" | xargs gofmt -s -w
```

## Architecture

The Atriumn SDK for Go provides clients for various Atriumn services:

### SDK Structure

1. **Service-specific packages**: Each service has its own package with a client, models, and tests:
   - `/auth` - Authentication service client
   - `/storage` - Storage service client for file operations
   - `/ai` - AI service client for prompt management
   - `/ingest` - Ingest service client for content ingestion

2. **Internal shared components**:
   - `/internal/clientutil` - Common HTTP request handling and error processing
   - `/internal/apierror` - Standardized error response structure

3. **Examples**: Implementation examples for each service in `/examples`

### Client Implementation Pattern

All service clients follow the same implementation pattern:

1. **Client Creation**:
   - Basic constructor: `NewClient(baseURL)`
   - Options-based constructor: `NewClientWithOptions(baseURL, ...options)`
   - Functional options for configuration (e.g., `WithHTTPClient`, `WithUserAgent`)

2. **Request/Response Flow**:
   - `newRequest()` - Creates HTTP requests with proper headers and JSON encoding
   - `do()` - Executes the request and processes responses using `clientutil.ExecuteRequest`
   - Service-specific methods that wrap these internal methods

3. **Error Handling**:
   - Consistent error types via `apierror.ErrorResponse`
   - Detailed error codes and descriptions
   - Type assertions to detect specific error types

### Authentication Mechanisms

The SDK supports different authentication mechanisms:

1. **Client Credentials** (via Auth client)
2. **TokenProvider interface** (for authenticated services)
3. **JWT tokens** for user authentication
4. **AWS IAM SigV4** for some services (like AI)

## Development Guidelines

1. **API Client Consistency**:
   - Follow existing patterns for new services
   - Use the same option pattern for client configuration
   - Implement consistent error handling

2. **Documentation**:
   - Maintain doc comments for public APIs
   - Include parameter descriptions, return values, and error cases
   - Provide usage examples in README.md files

3. **Testing**:
   - Write tests for all client methods
   - Include both success and error cases
   - Use the same test patterns across services

4. **Error Handling**:
   - Return typed errors using `apierror.ErrorResponse`
   - Include detailed error messages
   - Document possible error codes