# Atriumn SDK for Go: Component Responsibility Diagram

The diagram below illustrates the main components of the `atriumn-sdk-go` repository and their primary responsibilities.

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          ATRIUMN-SDK-GO                                     │
├────────────────┬─────────────────┬───────────────────┬────────────────────┤
│   auth/        │   storage/      │   ai/             │   ingest/          │
│                │                 │                   │                    │
│ Authentication │ Storage Service │ AI Service        │ Ingest Service     │
│ Service Client │ Client          │ Client            │ Client             │
│                │                 │                   │                    │
│ Responsibilities:                                                         │
├────────────────┼─────────────────┼───────────────────┼────────────────────┤
│ - User signup  │ - File uploads  │ - Prompt creation │ - Text ingestion   │
│ - User login   │ - File downloads│ - Prompt retrieval│ - URL ingestion    │
│ - Token mgmt   │ - Pre-signed    │ - Prompt updates  │ - File ingestion   │
│ - Password     │   URL generation│ - Prompt deletion │ - Upload URL       │
│   reset flows  │ - S3 integration│ - Prompt listing  │   generation       │
│ - Auth errors  │ - JWT auth      │ - Model settings  │ - Metadata handling│
├────────────────┴─────────────────┴───────────────────┴────────────────────┤
│                            internal/                                       │
├────────────────────────────┬───────────────────────────────────────────────┤
│   apierror/                │   clientutil/                                 │
│                            │                                               │
│ - Standard error types     │ - HTTP request execution                      │
│ - Error code definitions   │ - Response handling                           │
│ - Error message formatting │ - Status code checking                        │
│ - Error serialization      │ - Error response parsing                      │
│                            │ - Request building utilities                  │
└────────────────────────────┴───────────────────────────────────────────────┘
```

## Component Relationships

- Each service client package (`auth`, `storage`, `ai`, `ingest`) provides a client library for interacting with the corresponding Atriumn service API.
- All service clients rely on the common utilities in `internal/clientutil` for HTTP request handling and error processing.
- The `internal/apierror` package defines common error types and handling logic used across all service clients.
- Each service client follows a consistent pattern:
  - A main client struct with configuration options
  - Request/response model definitions
  - Method implementations for API endpoints
  - Service-specific error handling

## Code Organization Pattern

```
servicepackage/
  ├── client.go       # Client implementation and API methods
  ├── client_test.go  # Client testing
  ├── models.go       # Request/response data models
  └── README.md       # Documentation specific to this client
```

## Design Principles

1. **Consistency**: All service clients follow the same patterns and conventions
2. **Simplicity**: Clean, idiomatic Go interfaces that are easy to use
3. **Completeness**: Full coverage of service API functionality
4. **Robustness**: Comprehensive error handling and recovery
5. **Documentation**: Clear usage examples and API documentation