# Atriumn SDK for Go: Service Boundary Diagram

The diagram below illustrates the boundaries between the `atriumn-sdk-go` repository and its surrounding ecosystem, including the services it interacts with and the applications that consume it.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               GO CLIENT APPLICATIONS                             │
│                                                                                 │
│  ┌───────────────┐  ┌───────────────┐  ┌────────────────┐  ┌───────────────┐    │
│  │ Web Services  │  │ CLI Tools     │  │ Microservices  │  │ Other Go Apps │    │
│  └───────┬───────┘  └───────┬───────┘  └────────┬───────┘  └───────┬───────┘    │
│          │                  │                   │                  │            │
└──────────┼──────────────────┼───────────────────┼──────────────────┼────────────┘
           │                  │                   │                  │
           │                  │  imports          │                  │
           ▼                  ▼                   ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               ATRIUMN-SDK-GO                                     │
│                                                                                 │
│  ┌───────────────┐  ┌───────────────┐  ┌────────────────┐  ┌───────────────┐    │
│  │ Auth Client   │  │ Storage Client│  │ AI Client      │  │ Ingest Client │    │
│  └───────┬───────┘  └───────┬───────┘  └────────┬───────┘  └───────┬───────┘    │
│          │                  │                   │                  │            │
│  ┌───────┴───────────────────────────────────────────────────────────────────┐  │
│  │                          Internal Utilities                               │  │
│  │                                                                           │  │
│  │  ┌──────────────────────────┐     ┌──────────────────────────────┐        │  │
│  │  │ Error Handling (apierror)│     │ HTTP Client Utilities         │        │  │
│  │  └──────────────────────────┘     └──────────────────────────────┘        │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                 │
└──────────┼──────────────────┼───────────────────┼──────────────────┼────────────┘
           │                  │                   │                  │
           │ HTTP/REST        │ HTTP/REST         │ HTTP/REST        │ HTTP/REST
           │ API Calls        │ API Calls         │ API Calls        │ API Calls
           ▼                  ▼                   ▼                  ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Auth Service API │  │ Storage API      │  │ AI Service API   │  │ Ingest API       │
└──────────┬───────┘  └──────────┬───────┘  └──────────┬───────┘  └──────────┬───────┘
           │                     │                     │                     │
           ▼                     ▼                     ▼                     ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Auth Service     │  │ Storage Service  │  │ AI Service       │  │ Ingest Service   │
│ Implementation   │  │ Implementation   │  │ Implementation   │  │ Implementation   │
└──────────────────┘  └──────────────────┘  └──────────────────┘  └──────────────────┘
      Managed in          Managed in             Managed in            Managed in
     other repos         other repos            other repos           other repos
```

## Boundary Interactions

### Upstream Boundaries (SDK to Client Applications)
- **Interface**: Go package imports
- **Contract**: Public methods, types, and errors exposed by the SDK
- **Responsibilities**:
  - Providing a stable, well-documented API for Go applications
  - Maintaining backward compatibility for public interfaces
  - Handling authentication, error management, and API communication

### Downstream Boundaries (SDK to Atriumn Services)
- **Interface**: HTTP/REST API calls
- **Contract**: API endpoints, request/response formats, error codes
- **Responsibilities**:
  - Mapping Go types to HTTP request parameters
  - Translating API responses to Go return values
  - Converting service-specific errors to SDK error types
  - Implementing retry and recovery logic for transient failures

## Cross-Cutting Concerns

### Authentication
Authentication crosses multiple boundaries:
1. Client applications provide credentials to SDK
2. SDK exchanges credentials for tokens with Auth Service
3. SDK uses tokens to authenticate requests to other services

### Error Handling
Error handling is standardized across all services:
1. Service API errors are converted to SDK error types
2. Network and HTTP errors are wrapped in consistent format
3. SDK exposes typed errors to client applications

## Integration Points

### Integration with Other Atriumn Repositories
- **Auth Service Integration**: Token generation and validation
- **Storage Service Integration**: File uploads and downloads
- **AI Service Integration**: Prompt management and AI configuration
- **Ingest Service Integration**: Content ingestion and processing

### Integration with Client Applications
- **Dependency Management**: Go modules
- **Configuration**: Environment variables, explicit client configuration
- **Error Handling**: Type assertions for specific error handling
- **Authentication**: Token provider interfaces