# Deployment Guide

## Overview

The Atriumn SDK for Go is distributed as a Go module through the standard Go module system. This guide covers distribution processes, integration patterns, and version management for both SDK maintainers and consumers.

## Prerequisites for Distribution

### For SDK Maintainers

- **Go 1.24.0 or later** for development and testing
- **Git** with access to the atriumn/atriumn-sdk-go repository
- **GitHub CLI** or web access for creating releases
- **Make** for build automation

### For SDK Consumers

- **Go 1.21.0 or later** (minimum supported version)
- **Network access** to GitHub and Atriumn API endpoints
- **Valid Atriumn API credentials** for authentication

## Distribution Process

### Version Management

The SDK follows semantic versioning (SemVer) with the format `vMAJOR.MINOR.PATCH`:

- **MAJOR**: Breaking API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Creating a New Release

#### 1. Prepare Release

```bash
# Ensure all tests pass
make test

# Run full test suite with coverage
make test-coverage

# Verify code quality
make lint

# Update documentation if needed
make docs-check
```

#### 2. Tag the Release

```bash
# For patch releases (bug fixes)
make tag-patch

# For minor releases (new features)
make tag-minor

# For major releases (breaking changes)
make tag-major

# Or manually create tag
git tag v1.2.3
git push origin v1.2.3
```

#### 3. Create GitHub Release

```bash
# Using GitHub CLI
gh release create v1.2.3 --title "v1.2.3" --notes "Release notes here"

# Or create through GitHub web interface
```

### Module Structure

The SDK is organized as a multi-module Go repository:

```
github.com/atriumn/atriumn-sdk-go/
├── auth          # Authentication client
├── storage       # Storage service client
├── ai            # AI service client
├── ingest        # Content ingestion client
└── internal      # Shared internal utilities
```

Each service client is importable independently:

```go
import "github.com/atriumn/atriumn-sdk-go/auth"
import "github.com/atriumn/atriumn-sdk-go/storage"
```

## Integration Instructions

### Installation for Consumers

#### Option 1: Install Specific Services

```bash
# Install only what you need
go get github.com/atriumn/atriumn-sdk-go/auth@latest
go get github.com/atriumn/atriumn-sdk-go/storage@latest
```

#### Option 2: Install All Services

```bash
# Install all service clients
go get github.com/atriumn/atriumn-sdk-go/...@latest
```

#### Option 3: Pin to Specific Version

```bash
# Pin to a specific version
go get github.com/atriumn/atriumn-sdk-go/auth@v1.2.3
```

### Basic Integration Pattern

```go
package main

import (
    "context"
    "log"
    
    "github.com/atriumn/atriumn-sdk-go/auth"
    "github.com/atriumn/atriumn-sdk-go/storage"
)

func main() {
    ctx := context.Background()
    
    // Initialize clients
    authClient := auth.NewClient("your-api-key")
    
    // Use clients
    loginResp, err := authClient.Login(ctx, auth.LoginRequest{
        Email:    "user@example.com",
        Password: "password",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Chain services
    storageClient := storage.NewClient(loginResp.Token)
    // ... use storage client
}
```

### Configuration Management

#### Environment Variables

The SDK respects standard environment variables:

```bash
# API endpoint configuration
export ATRIUMN_API_BASE_URL="https://api.atriumn.com"

# Default timeout settings
export ATRIUMN_REQUEST_TIMEOUT="30s"

# Debug logging
export ATRIUMN_DEBUG="true"
```

#### Configuration in Code

```go
// Custom client configuration
config := auth.Config{
    BaseURL: "https://api.atriumn.com",
    Timeout: 30 * time.Second,
    Debug:   true,
}

client := auth.NewClientWithConfig("api-key", config)
```

## Monitoring and Observability

### Client Metrics

The SDK provides built-in metrics for monitoring:

```go
// Enable metrics collection
client := auth.NewClient("api-key")
client.EnableMetrics()

// Access metrics
metrics := client.GetMetrics()
fmt.Printf("Requests: %d, Errors: %d\n", metrics.Requests, metrics.Errors)
```

### Logging Configuration

```go
// Configure structured logging
import "github.com/atriumn/atriumn-sdk-go/internal/logger"

logger.SetLevel(logger.DebugLevel)
logger.SetFormat(logger.JSONFormat)
```

### Health Checks

For applications using the SDK:

```go
// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // Test auth service connectivity
    if err := authClient.HealthCheck(ctx); err != nil {
        http.Error(w, "Auth service unavailable", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

## Rollback Procedures

### For SDK Consumers

#### Downgrade to Previous Version

```bash
# Downgrade to specific version
go get github.com/atriumn/atriumn-sdk-go/auth@v1.1.2

# Update go.mod
go mod tidy

# Verify downgrade
go list -m github.com/atriumn/atriumn-sdk-go/auth
```

#### Emergency Rollback

```bash
# Quick rollback to last known good version
git checkout go.mod go.sum
go mod download
```

### For SDK Maintainers

#### Retract Problematic Release

```go
// In go.mod
retract v1.2.3 // Critical bug in authentication
```

#### Emergency Patch Release

```bash
# Create hotfix branch
git checkout -b hotfix/v1.2.4

# Apply critical fix
# ... make changes

# Fast-track release
make test
git tag v1.2.4
git push origin v1.2.4
gh release create v1.2.4 --title "v1.2.4 - Critical Fix"
```

## Troubleshooting Common Issues

### Import Path Issues

**Problem**: Module not found errors

```bash
go: github.com/atriumn/atriumn-sdk-go/auth@v1.2.3: reading module: Module not found
```

**Solution**:
```bash
# Clear module cache
go clean -modcache

# Re-download modules
go mod download

# Verify module exists
go list -m -versions github.com/atriumn/atriumn-sdk-go/auth
```

### Version Conflicts

**Problem**: Dependency version conflicts

**Solution**:
```bash
# Check current versions
go list -m all | grep atriumn

# Update to latest compatible versions
go get github.com/atriumn/atriumn-sdk-go/auth@latest
go mod tidy
```

### API Connectivity Issues

**Problem**: Network timeouts or connection refused

**Solution**:
```go
// Increase timeout
config := auth.Config{
    Timeout: 60 * time.Second,
    RetryConfig: &auth.RetryConfig{
        MaxRetries: 3,
        BackoffFactor: 2,
    },
}

client := auth.NewClientWithConfig("api-key", config)
```

### Authentication Failures

**Problem**: Invalid API key or token errors

**Solution**:
```bash
# Verify API key format
echo $ATRIUMN_API_KEY | wc -c  # Should be expected length

# Test with minimal client
go run -c 'package main
import "github.com/atriumn/atriumn-sdk-go/auth"
func main() {
    client := auth.NewClient("your-key")
    // Test basic operation
}'
```

## Maintenance Tasks

### Regular Maintenance

#### Weekly Tasks

```bash
# Update dependencies
go get -u ./...
go mod tidy

# Run security audit
go list -json -deps | nancy sleuth

# Check for outdated modules
go list -u -m all
```

#### Monthly Tasks

```bash
# Full test suite across Go versions
make test-all-versions

# Performance benchmarks
make benchmark

# Documentation review
make docs-check
```

### Dependency Management

#### Security Updates

```bash
# Check for security vulnerabilities
go list -json -deps | nancy sleuth

# Update vulnerable dependencies
go get -u vulnerable/package@latest
go mod tidy
```

#### Breaking Change Management

When consuming applications need to handle breaking changes:

1. **Pin to last compatible version**
2. **Review migration guide**
3. **Test in staging environment**
4. **Gradual rollout**

### Support and Escalation

#### Getting Help

- **Documentation**: Check [docs/](../docs/) directory
- **Examples**: Review [examples/](../examples/) directory
- **Issues**: Create GitHub issue with reproduction steps
- **Security**: Contact security@atriumn.com for security issues

#### Issue Escalation

1. **Level 1**: Check documentation and examples
2. **Level 2**: Search existing GitHub issues
3. **Level 3**: Create new GitHub issue with details
4. **Level 4**: Contact Atriumn support team

#### Required Information for Issues

```bash
# Environment information
go version
go env GOOS GOARCH

# Module versions
go list -m github.com/atriumn/atriumn-sdk-go/...

# Minimal reproduction case
# Include code that demonstrates the issue
```

## Performance Considerations

### Connection Pooling

```go
// Configure HTTP client for better performance
config := auth.Config{
    HTTPClient: &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:       100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:    90 * time.Second,
        },
    },
}
```

### Caching

```go
// Enable token caching for auth client
authClient := auth.NewClient("api-key")
authClient.EnableTokenCaching(30 * time.Minute)
```

### Batch Operations

```go
// Use batch operations when available
ingestClient := ingest.NewClient("token")
results, err := ingestClient.IngestBatch(ctx, []ingest.IngestRequest{
    // ... multiple requests
})
```
