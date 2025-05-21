# Atriumn SDK for Go: Repository Boundaries

This document outlines the purpose, responsibilities, and boundaries of the `atriumn-sdk-go` repository within the broader Atriumn system architecture.

## Core Purpose

### Primary Business Functions
The `atriumn-sdk-go` repository serves as the official Go SDK for Atriumn services, providing client libraries that enable Go applications to seamlessly integrate with Atriumn's backend services. Its primary functions include:

1. **Service Client Implementations** - Providing idiomatic Go clients for all Atriumn API services
2. **Authentication Handling** - Managing authentication flows and token handling
3. **Error Standardization** - Providing consistent error handling across different service clients
4. **Request/Response Abstractions** - Abstracting HTTP-level details behind clean Go interfaces
5. **Type Safety** - Ensuring type-safe interactions with Atriumn APIs

### Key Problems Solved
This repository solves several critical problems for Atriumn's Go ecosystem:

1. **Abstraction of API Complexity** - Shields developers from low-level HTTP details, authentication complexities, and error handling
2. **Code Reusability** - Eliminates duplicate client code across Go applications and services
3. **Consistency** - Ensures all Go applications interact with Atriumn services in a consistent, standardized way
4. **Maintainability** - Centralizes API client changes, allowing service API updates without requiring changes to multiple applications
5. **Versioning** - Provides a single versioning mechanism for all Atriumn service clients

### Critical User Scenarios Supported
The SDK supports several critical integration scenarios:

1. **Authentication Workflows** - User signup, login, password management, and token-based auth
2. **Content Storage & Retrieval** - Generating pre-signed URLs for secure file uploads and downloads
3. **AI Prompt Management** - Creating, retrieving, updating, and deleting AI prompts
4. **Data Ingestion** - Ingesting text, URLs, and files into the Atriumn platform

## Domain Ownership

### Data Models Owned
This repository owns the Go representation of the following data models:

1. **Authentication Models**
   - Token responses
   - User registration data
   - Password reset workflows

2. **Storage Models**
   - File upload/download requests and responses
   - S3 pre-signed URL metadata

3. **AI Models**
   - Prompt templates and configurations
   - Prompt variables and parameters
   - Model references

4. **Ingest Models**
   - Content ingestion requests and responses
   - File metadata
   - Ingestion status information

### Business Logic Ownership Boundaries
The repository owns the following business logic:

1. **Client Configuration Logic**
   - Base URL handling
   - Client instantiation with functional options
   - Custom HTTP client configuration

2. **Authentication Logic**
   - Token refresh mechanisms
   - Access token caching
   - Authentication header management

3. **Request Building Logic**
   - Request body serialization
   - Header standardization
   - Query parameter construction

4. **Response Handling Logic**
   - Response deserialization
   - Error detection and conversion
   - Retry logic for transient errors

### Canonical Data Sources and Sinks

#### Sources
- API specifications for Atriumn services (external)
- Service endpoint documentation (external)
- API versioning information (external)

#### Sinks
- API client interfaces exposed to Go applications
- Data models and types for client applications
- Error definitions consumed by client applications

## Service Boundaries

### IN Scope
The following are explicitly within scope for this repository:

1. **Client Libraries**
   - Go client implementations for all Atriumn services
   - Request/response models for all API operations
   - Client configuration and instantiation mechanisms

2. **Authentication Mechanisms**
   - Token-based authentication support
   - Integration with Atriumn auth services
   - Credential management interfaces

3. **Error Handling**
   - Standardized error types
   - Consistent error reporting
   - Network/HTTP error translation

4. **Documentation**
   - Client usage examples
   - Integration patterns
   - API references for Go interfaces

5. **Testing Tools**
   - Client mocks for testing
   - Test utilities for integration testing
   - Examples of proper client usage

### OUT of Scope
The following are explicitly out of scope for this repository:

1. **Service Implementations**
   - Actual implementation of Atriumn services
   - Server-side business logic
   - API endpoint implementations

2. **Authentication Servers**
   - Identity provider implementation
   - Token issuance services
   - User database management

3. **Application Business Logic**
   - Application-specific workflows
   - UI/UX components
   - End-user features

4. **Infrastructure**
   - Deployment configurations
   - Service monitoring
   - Cloud provider integrations

5. **Cross-language SDKs**
   - SDK implementations for other languages (JavaScript, Python, etc.)
   - Language-specific design patterns

### Interface Contracts Provided
This repository provides the following interface contracts to other services:

1. **Client Interfaces**
   - Public methods for accessing each Atriumn service
   - Type definitions for all request/response models
   - Error types and handling patterns

2. **Authentication Integration**
   - Token provider interfaces
   - Authentication mechanism abstractions
   - Error handling for auth failures

3. **Configuration Interfaces**
   - Client option patterns
   - Environment-based configuration
   - Runtime reconfiguration capabilities

### Dependencies on Other Repositories
The SDK has the following external dependencies:

1. **Atriumn API Services**
   - Depends on the stability and contract of the various Atriumn API services
   - Changes to API contracts require corresponding SDK updates

2. **Standard Go Libraries**
   - Uses standard Go HTTP libraries and encoding packages
   - Follows Go module versioning for dependency management

3. **Minimal External Dependencies**
   - Limited to absolutely necessary third-party libraries
   - Preference for standard library solutions when possible

## Decision Guidelines

### When to Add Functionality to This Repository
New functionality should be added to this repository when:

1. **New Atriumn API Service**
   - When a new Atriumn backend service becomes available
   - When existing services add new endpoints or capabilities

2. **Client Library Improvements**
   - Enhancements to existing clients (better error handling, retries)
   - Performance optimizations for API communication
   - Additional helper methods for common operations

3. **Common Utilities**
   - Shared functionality needed across multiple service clients
   - Standardized authentication or serialization mechanisms
   - Testing utilities specific to API interactions

4. **Documentation and Examples**
   - New usage examples
   - Integration patterns
   - Best practices

### When to Delegate Functionality Elsewhere
Functionality should be delegated elsewhere when:

1. **Service Implementation Logic**
   - Core business logic should remain in the service repositories
   - API endpoint implementations belong in the service code

2. **Application-Specific Code**
   - Application workflows and use-case-specific logic
   - Custom integration patterns specific to a single application

3. **Infrastructure Concerns**
   - Deployment, scaling, and operational concerns
   - Monitoring and logging implementations

4. **Authentication Providers**
   - Identity management systems
   - Token issuance and validation services

### Migration Path for Boundary Changes
When boundaries need to change, the following migration approach should be followed:

1. **API Service Changes**
   - For breaking changes: Create a new major version of the affected client
   - For additions: Extend existing client interfaces
   - For deprecations: Mark methods as deprecated with comments and alternatives

2. **Model Changes**
   - When data models change: Provide adapters or conversion utilities
   - When adding fields: Use pointers for optional new fields
   - When removing fields: Deprecate first, then remove in next major version

3. **Client Interface Changes**
   - When changing method signatures: Introduce new methods rather than changing existing ones
   - When adding options: Use the functional options pattern for backward compatibility
   - When changing behavior: Document clearly and increment version appropriately

## Diagrams

For detailed visual representations of the repository structure and boundaries, please refer to the following diagrams:

1. [Component Responsibility Diagram](./COMPONENT_DIAGRAM.md) - Illustrates the main components of the repository and their responsibilities
2. [Service Boundary Diagram](./SERVICE_BOUNDARY_DIAGRAM.md) - Shows how the SDK interacts with client applications and backend services