# Atriumn SDK for Go: Architecture

This document describes the architecture of the `atriumn-sdk-go` repository, including its component structure, data flows, infrastructure considerations, and key technical decisions.

## 1. Component Structure

### 1.1 Major Components and Their Responsibilities

The SDK is organized into distinct service client packages, each responsible for interacting with a specific Atriumn backend service:

#### Service Client Packages

| Package | Responsibility |
|---------|----------------|
| `auth/` | Authentication service client for user management and token-based authentication |
| `storage/` | Storage service client for managing file uploads/downloads and pre-signed URL generation |
| `ai/` | AI service client for prompt creation, management, and configuration |
| `ingest/` | Ingest service client for content ingestion (text, URLs, files) |

#### Internal Utilities

| Package | Responsibility |
|---------|----------------|
| `internal/apierror/` | Common error types, error handling, and error response formatting |
| `internal/clientutil/` | Shared HTTP request handling, response processing, and error conversion |

### 1.2 Layer Boundaries

The SDK follows a clean separation of concerns with well-defined layers:

1. **Public Interface Layer**
   - Client structs with public methods (`Client` in each service package)
   - Domain models representing API resources
   - Request/response structs for API operations

2. **HTTP Communication Layer**
   - HTTP request building and execution
   - Authentication header management
   - Request/response serialization
   - Status code handling

3. **Error Handling Layer**
   - Standard error types
   - Error response parsing
   - User-friendly error messages

### 1.3 Design Patterns

The SDK employs several key design patterns for consistency and extensibility:

1. **Client-Builder Pattern**
   - Core `Client` struct in each service package
   - `NewClient` constructor for basic initialization
   - `NewClientWithOptions` for advanced configuration
   - Functional options pattern for flexible client configuration

2. **Interface Segregation**
   - Each service client focuses on a specific domain
   - Clear separation between authentication, storage, AI, and ingest operations
   - `TokenProvider` interface for pluggable authentication mechanisms

3. **Consistent Method Signatures**
   - Context-based request cancellation
   - Domain model return types
   - Error handling conventions

4. **Resource-Oriented API Design**
   - Methods organized around resource operations (Create, Get, Update, Delete, List)
   - Resource-specific request and response types
   - Consistent parameter and return value patterns

## 2. Data Flows

### 2.1 Request/Response Patterns

The SDK follows a consistent request/response flow across all service clients:

1. **Request Initiation**
   - Client method called with domain-specific parameters
   - Parameters validated and converted to API request format
   - HTTP request constructed with proper headers and authentication

2. **Request Execution**
   - HTTP request sent to Atriumn API
   - Context handling for timeouts and cancellation
   - Network-level error handling

3. **Response Processing**
   - Status code validation
   - Error response detection and conversion
   - Successful response deserialization into domain models
   - Domain model returned to caller

A typical request flow is illustrated below:

```
┌─────────────┐     ┌─────────────┐     ┌───────────────┐     ┌────────────┐
│ Client API  │────▶│ HTTP Request│────▶│ Atriumn API   │────▶│ HTTP       │
│ Method Call │     │ Construction│     │ Endpoint      │     │ Response   │
└─────────────┘     └─────────────┘     └───────────────┘     └────────────┘
                                                                    │
┌─────────────┐     ┌─────────────┐     ┌───────────────┐          │
│ Return      │◀────│ Domain Model│◀────│ Response      │◀─────────┘
│ To Caller   │     │ Creation    │     │ Processing    │
└─────────────┘     └─────────────┘     └───────────────┘
```

### 2.2 Error Handling Flow

Error handling follows a consistent pattern across all service clients:

1. Network errors are captured and translated to user-friendly `apierror.ErrorResponse` types
2. HTTP error status codes are converted to appropriate error types with descriptive messages
3. API-specific error responses are parsed and preserved
4. Consistent error interface is exposed to client applications

Error flow:

```
┌─────────────┐     ┌─────────────┐     ┌───────────────┐
│ HTTP Error  │────▶│ Error       │────▶│ apierror.     │
│ Occurs      │     │ Detection   │     │ ErrorResponse │
└─────────────┘     └─────────────┘     └───────────────┘
                                               │
┌─────────────┐     ┌─────────────┐            │
│ Return      │◀────│ Error       │◀────────────┘
│ To Caller   │     │ Propagation │
└─────────────┘     └─────────────┘
```

### 2.3 Authentication Flow

The SDK supports token-based authentication through the `TokenProvider` interface:

1. Application code provides a `TokenProvider` implementation
2. The SDK requests tokens when making authenticated API calls
3. Tokens are included in Authorization headers
4. Token errors are propagated back to client applications

## 3. Infrastructure

### 3.1 Deployment Architecture

The `atriumn-sdk-go` is deployed as a Go module that client applications import. Key infrastructure considerations include:

1. **Module Versioning**
   - Semantic versioning for compatibility guarantees
   - Go module system for dependency management
   - Backward compatibility within major versions

2. **Dependency Management**
   - Minimal external dependencies to reduce supply chain risk
   - Strict versioning of dependencies
   - Preference for standard library solutions

3. **Distribution**
   - Public Go module repository
   - CI/CD for automated testing and release

### 3.2 Scaling Considerations

While the SDK itself isn't directly responsible for scaling, it supports scalable applications through:

1. **Connection Pooling**
   - HTTP connection reuse for efficient API communication
   - Configurable HTTP client for custom connection settings

2. **Context Support**
   - All API methods accept a context.Context for request cancellation
   - Timeout management to prevent resource exhaustion

3. **Resource Efficiency**
   - Minimized memory allocations for request/response handling
   - Efficient error handling without excessive allocations

### 3.3 Infrastructure Dependencies

The SDK has the following infrastructure dependencies:

1. **Atriumn Backend Services**
   - Authentication Service API
   - Storage Service API
   - AI Service API
   - Ingest Service API

2. **Network Infrastructure**
   - HTTP/HTTPS connectivity
   - DNS resolution
   - TLS certificate validation

## 4. Technical Decisions

### 4.1 Key Architectural Decisions

1. **Separate Service Clients**
   - **Decision**: Organize code into distinct service-specific packages
   - **Rationale**: Clear separation of concerns, allows selective imports, matches backend service boundaries
   - **Trade-offs**: More packages to maintain vs. cleaner organization

2. **Shared Internal Utilities**
   - **Decision**: Common functionality in internal packages
   - **Rationale**: Code reuse, consistent behavior across service clients
   - **Trade-offs**: Internal coupling vs. duplication

3. **Context-Based API Design**
   - **Decision**: All API methods accept a context.Context parameter
   - **Rationale**: Enables timeout and cancellation control, follows Go best practices
   - **Trade-offs**: Additional parameter vs. better control flow

4. **Functional Options Pattern**
   - **Decision**: Client customization via functional options
   - **Rationale**: Flexible configuration without breaking changes, idiomatic Go
   - **Trade-offs**: More complex than simple struct initialization

### 4.2 Technology Choices

1. **Standard Library HTTP Client**
   - **Choice**: Use Go's standard `net/http` package
   - **Rationale**: Minimal dependencies, well-maintained, familiar to Go developers
   - **Alternatives Considered**: Custom HTTP clients, third-party libraries

2. **JSON for Serialization**
   - **Choice**: Standard library `encoding/json` for request/response serialization
   - **Rationale**: Simple, universally supported, matches API requirements
   - **Alternatives Considered**: Protocol Buffers, custom binary formats

3. **Error Type System**
   - **Choice**: Custom `apierror.ErrorResponse` type implementing error interface
   - **Rationale**: Preserves API error details while satisfying Go error handling
   - **Alternatives Considered**: Sentinel errors, error wrapping

### 4.3 Performance Considerations

1. **Connection Pooling**
   - HTTP connection reuse via `http.Client`
   - Configurable timeouts to prevent resource exhaustion

2. **Efficient Request/Response Handling**
   - Request body streaming
   - Response body read exactly once
   - Proper body closing to prevent leaks

3. **Memory Efficiency**
   - Appropriate use of pointers vs. values
   - Minimized allocations in request processing
   - Error handling without excessive allocations

## 5. API Design Principles

The Atriumn SDK for Go follows these core API design principles:

1. **Idiomatic Go**
   - Method names follow Go conventions (e.g., `CreatePrompt` not `create_prompt`)
   - Error handling using Go's standard approach
   - Context-based cancellation and timeouts

2. **Consistency**
   - Uniform method signatures across all service clients
   - Consistent parameter ordering and return values
   - Standardized error handling

3. **Usability**
   - Clear, descriptive method and parameter names
   - Comprehensive documentation with examples
   - Intuitive resource-oriented organization

4. **Extensibility**
   - Functional options for future extensibility
   - Interface-based design for pluggable components
   - Versioned API for controlled evolution