# Atriumn SDK for Go

This repository contains the official Go SDKs for Atriumn services.

## Available SDKs

### [Auth](/auth)

Go client for the Atriumn Authentication Service. Provides functionality for:

- Client Credentials authentication
- User signup, login, and management
- Password reset workflows

### [Storage](/storage)

Go client for the Atriumn Storage Service. Provides functionality for:

- Generating pre-signed S3 URLs for file uploads
- Generating pre-signed S3 URLs for file downloads
- JWT token-based authentication

## Installation

```bash
# Install the auth client
go get github.com/atriumn/atriumn-sdk-go/auth

# Install the storage client
go get github.com/atriumn/atriumn-sdk-go/storage
```

## Development

### Testing

```bash
# Run all tests
go test -v ./...

# Test a specific package
go test -v ./auth
go test -v ./storage
```

## License

[MIT](LICENSE)
