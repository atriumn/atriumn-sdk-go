# Atriumn SDK for Go: Integration Guide

This document outlines all integration points between the `atriumn-sdk-go` library and other repositories or external systems, including API contracts, event patterns, service dependencies, and data exchange formats.

## Table of Contents

1. [External APIs](#1-external-apis)
   - [Auth Service](#11-auth-service)
   - [AI Service](#12-ai-service)
   - [Storage Service](#13-storage-service)
   - [Ingest Service](#14-ingest-service)
   - [Amazon S3](#15-amazon-s3)

2. [Event Patterns](#2-event-patterns)
   - [Events Published](#21-events-published)
   - [Events Consumed](#22-events-consumed)
   - [Message Formats](#23-message-formats)
   - [Delivery Guarantees](#24-delivery-guarantees)

3. [Service Dependencies](#3-service-dependencies)
   - [External Services](#31-external-services)
   - [Required Capabilities](#32-required-capabilities)
   - [Failure Handling](#33-failure-handling)

4. [Data Exchange](#4-data-exchange)
   - [Shared Data Models](#41-shared-data-models)
   - [Serialization Formats](#42-serialization-formats)
   - [Schema Evolution Approach](#43-schema-evolution-approach)

5. [Integration Sequence Diagrams](#5-integration-sequence-diagrams)

---

## 1. External APIs

### 1.1 Auth Service

The Auth Service provides authentication and user management capabilities.

#### REST Endpoints

| Endpoint | Method | Description | Auth Required | Rate Limit |
|----------|--------|-------------|--------------|------------|
| `/auth/token` | POST | OAuth token requests | Client Credentials | 10 req/min |
| `/auth/signup` | POST | User registration | None | 5 req/min |
| `/auth/signup/confirm` | POST | Confirm user registration | None | 5 req/min |
| `/auth/signup/resend` | POST | Resend confirmation code | None | 3 req/5min |
| `/auth/login` | POST | User authentication | Basic Auth | 10 req/min |
| `/auth/logout` | POST | User logout | Bearer Token | 10 req/min |
| `/auth/password/reset` | POST | Password reset request | None | 3 req/10min |
| `/auth/password/confirm` | POST | Confirm password reset | None | 3 req/10min |
| `/auth/me` | GET | User profile access | Bearer Token | 20 req/min |
| `/admin/credentials` | POST | Create client credentials | Bearer Token | 5 req/min |
| `/admin/credentials` | GET | List client credentials | Bearer Token | 10 req/min |
| `/admin/credentials/{id}` | GET | Get client credential | Bearer Token | 20 req/min |
| `/admin/credentials/{id}` | PATCH | Update client credential | Bearer Token | 5 req/min |
| `/admin/credentials/{id}` | DELETE | Delete client credential | Bearer Token | 5 req/min |
| `/health` | GET | Service health check | None | 60 req/min |

#### Authentication Requirements

- **OAuth 2.0 Client Credentials Flow**
  - Used for server-to-server authentication
  - Requires `client_id` and `client_secret`
  - Returns an access token with configurable expiration

- **JWT Bearer Token Authentication**
  - Format: `Authorization: Bearer {token}`
  - Used for authenticated API requests
  - Tokens typically expire after 1 hour

- **HTTP Basic Authentication**
  - Used only for the login endpoint
  - Format: `Authorization: Basic {base64(username:password)}`

#### Request/Response Formats

Example of token request:

```json
// POST /auth/token
// Request
{
  "grant_type": "client_credentials",
  "client_id": "abc123",
  "client_secret": "xyz789",
  "scope": "auth:admin storage:read"
}

// Response
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "bearer",
  "expires_in": 3600
}
```

#### Rate Limiting and Quotas

- Rate limits are specified per endpoint as shown in the endpoints table
- Exceeded rate limits return HTTP 429 (Too Many Requests)
- Headers provide rate limit information:
  - `X-RateLimit-Limit`: Requests allowed per period
  - `X-RateLimit-Remaining`: Requests remaining in current period
  - `X-RateLimit-Reset`: Seconds until rate limit resets

### 1.2 AI Service

The AI Service provides capabilities for managing AI prompts and configurations.

#### REST Endpoints

| Endpoint | Method | Description | Auth Required | Rate Limit |
|----------|--------|-------------|--------------|------------|
| `/prompts` | POST | Create a new prompt | Bearer Token | 10 req/min |
| `/prompts` | GET | List prompts | Bearer Token | 20 req/min |
| `/prompts/{id}` | GET | Get a specific prompt | Bearer Token | 30 req/min |
| `/prompts/{id}` | PUT | Update a prompt | Bearer Token | 10 req/min |
| `/prompts/{id}` | DELETE | Delete a prompt | Bearer Token | 10 req/min |

#### Authentication Requirements

- **JWT Bearer Token Authentication**
  - Format: `Authorization: Bearer {token}`
  - Token must have appropriate scopes (e.g., `ai:read`, `ai:write`)

#### Request/Response Formats

Example of creating a prompt:

```json
// POST /prompts
// Request
{
  "name": "Product Description Generator",
  "description": "Generates compelling product descriptions for e-commerce",
  "template": "Create a compelling description for {{product_name}} that highlights its {{feature}}.",
  "modelId": "gpt-4",
  "parameters": {
    "temperature": 0.7,
    "max_tokens": 300
  },
  "variables": [
    {
      "name": "product_name",
      "description": "The name of the product",
      "required": true
    },
    {
      "name": "feature",
      "description": "Main feature to highlight",
      "defaultValue": "unique selling point",
      "required": true
    }
  ],
  "tags": ["e-commerce", "marketing", "product"]
}

// Response
{
  "prompt": {
    "id": "prompt-123",
    "name": "Product Description Generator",
    "description": "Generates compelling product descriptions for e-commerce",
    "template": "Create a compelling description for {{product_name}} that highlights its {{feature}}.",
    "modelId": "gpt-4",
    "parameters": {
      "temperature": 0.7,
      "max_tokens": 300
    },
    "variables": [
      {
        "name": "product_name",
        "description": "The name of the product",
        "required": true
      },
      {
        "name": "feature",
        "description": "Main feature to highlight",
        "defaultValue": "unique selling point",
        "required": true
      }
    ],
    "tags": ["e-commerce", "marketing", "product"],
    "version": 1,
    "createdAt": "2025-05-21T04:55:58Z",
    "updatedAt": "2025-05-21T04:55:58Z"
  }
}
```

#### Rate Limiting and Quotas

- Rate limits are specified per endpoint as shown in the endpoints table
- Default quotas:
  - Maximum 1000 prompts per tenant
  - Maximum 100KB size per prompt template

### 1.3 Storage Service

The Storage Service provides secure file upload and download capabilities through pre-signed URLs.

#### REST Endpoints

| Endpoint | Method | Description | Auth Required | Rate Limit |
|----------|--------|-------------|--------------|------------|
| `/generate-upload-url` | POST | Generate pre-signed URL for file upload | Bearer Token | 30 req/min |
| `/generate-download-url` | POST | Generate pre-signed URL for file download | Bearer Token | 60 req/min |

#### Authentication Requirements

- **JWT Bearer Token Authentication**
  - Format: `Authorization: Bearer {token}`
  - Token must have appropriate scopes (e.g., `storage:read`, `storage:write`)

#### Request/Response Formats

Example of generating an upload URL:

```json
// POST /generate-upload-url
// Request
{
  "filename": "document.pdf",
  "contentType": "application/pdf",
  "tenantId": "tenant-123"
}

// Response
{
  "uploadUrl": "https://storage.atriumn.io/bucket/tenant-123/abc123/document.pdf?signed=...",
  "s3Key": "tenant-123/abc123/document.pdf",
  "httpMethod": "PUT"
}
```

#### Rate Limiting and Quotas

- Rate limits are specified per endpoint as shown in the endpoints table
- Default quotas:
  - Maximum file size: 100MB
  - Maximum storage per tenant: 10GB
  - Pre-signed URLs expire after 15 minutes

### 1.4 Ingest Service

The Ingest Service enables uploading and processing various types of content (text, URLs, files).

#### REST Endpoints

| Endpoint | Method | Description | Auth Required | Rate Limit |
|----------|--------|-------------|--------------|------------|
| `/ingest/text` | POST | Ingest text content | Bearer Token | 20 req/min |
| `/ingest/url` | POST | Ingest content from URL | Bearer Token | 10 req/min |
| `/ingest/file` | POST | Request file upload URL | Bearer Token | 20 req/min |
| `/content` | GET | List content items | Bearer Token | 30 req/min |
| `/content/{id}` | GET | Get content item metadata | Bearer Token | 60 req/min |
| `/content/{id}` | PATCH | Update content item metadata | Bearer Token | 20 req/min |
| `/content/{id}` | DELETE | Delete content item | Bearer Token | 10 req/min |
| `/content/{id}/download-url` | GET | Get pre-signed download URL | Bearer Token | 60 req/min |
| `/content/{id}/text` | GET | Get text content | Bearer Token | 30 req/min |
| `/content/{id}/text` | PUT | Update text content | Bearer Token | 10 req/min |

#### Authentication Requirements

- **JWT Bearer Token Authentication**
  - Format: `Authorization: Bearer {token}`
  - Token must have appropriate scopes (e.g., `ingest:read`, `ingest:write`)

#### Request/Response Formats

Example of URL ingestion request:

```json
// POST /ingest/url
// Request
{
  "tenantId": "tenant-123",
  "userId": "user-456",
  "url": "https://example.com/article",
  "metadata": {
    "source": "web-crawler",
    "category": "news"
  }
}

// Response
{
  "id": "content-789",
  "status": "PENDING"
}
```

#### Rate Limiting and Quotas

- Rate limits are specified per endpoint as shown in the endpoints table
- Default quotas:
  - Maximum text content size: 1MB
  - Maximum URL length: 2048 characters
  - Maximum file size: 100MB
  - Maximum items per tenant: 10,000

### 1.5 Amazon S3

The SDK interacts with Amazon S3 or S3-compatible storage for direct file uploads and downloads.

#### Endpoints

Pre-signed URLs generated by Storage and Ingest services point to S3 buckets.

#### Authentication Requirements

- No additional authentication needed when using pre-signed URLs
- Pre-signed URLs include temporary credentials and signature within the URL parameters

#### Request/Response Formats

File upload to S3 pre-signed URL:

```
PUT https://s3.amazonaws.com/bucket-name/key?AWSAccessKeyId=...&Expires=...&Signature=...
Content-Type: application/pdf
Content-Length: 12345

[Binary File Data]
```

#### Rate Limiting and Quotas

- S3 service quotas apply:
  - Max object size: 5TB
  - S3 rate limits depend on AWS account configuration
  - Pre-signed URLs have expiration times (typically 15 minutes)

## 2. Event Patterns

### 2.1 Events Published

The SDK does not directly publish events to external systems. Event publishing is handled by the backend services.

### 2.2 Events Consumed

The SDK does not directly consume events from external systems. Instead, it polls for status changes or receives webhooks through API endpoints.

#### Asynchronous Processing Status Updates

While not technically events, the SDK handles asynchronous processing status updates for:

1. **URL Ingestion Processing**
   - URL content is processed asynchronously
   - Initial status is `PENDING` or `PROCESSING`
   - Client application must poll for status changes

2. **File Processing**
   - Large files may undergo processing after upload
   - Initial status is `UPLOADING` or `PROCESSING`
   - Final status becomes `COMPLETED` or `ERROR`

### 2.3 Message Formats

No direct message consumption occurs within the SDK. Status updates are delivered through standard API responses.

### 2.4 Delivery Guarantees

The SDK relies on HTTP response codes and retries for reliable communication:

- Automatic retries for network errors (configurable in client options)
- No guaranteed delivery for asynchronous operations
- Client applications should implement their retry logic for status polling

## 3. Service Dependencies

### 3.1 External Services

The Atriumn SDK for Go depends on the following external services:

1. **Atriumn Auth Service**
   - Provides authentication and user management
   - Required for all authenticated operations
   - Base URL format: `https://{auth-service-hostname}`

2. **Atriumn AI Service**
   - Provides AI prompt management capabilities
   - Base URL format: `https://{ai-service-hostname}`

3. **Atriumn Storage Service**
   - Provides secure file storage capabilities
   - Base URL format: `https://{storage-service-hostname}`

4. **Atriumn Ingest Service**
   - Provides content ingestion and processing
   - Base URL format: `https://{ingest-service-hostname}`

5. **Amazon S3 / S3-Compatible Storage**
   - Used for direct file uploads and downloads
   - Accessed via pre-signed URLs

### 3.2 Required Capabilities

Required capabilities from external services:

1. **Auth Service**
   - OAuth 2.0 client credentials flow
   - JWT token issuance and validation
   - User management (signup, login, password reset)

2. **AI Service**
   - Prompt template storage and retrieval
   - Prompt versioning
   - Tagging and filtering

3. **Storage Service**
   - Generation of S3 pre-signed URLs
   - Access control based on tenant ID
   - File metadata management

4. **Ingest Service**
   - Content ingestion and processing
   - Content indexing and retrieval
   - Multi-tenant content isolation

5. **S3-Compatible Storage**
   - Binary file storage
   - Support for pre-signed URLs
   - Content-type preservation

### 3.3 Failure Handling

The SDK implements the following failure handling strategies:

1. **Connection Failures**
   - Timeouts default to 10 seconds (configurable)
   - Network errors are converted to `apierror.ErrorResponse` with code `network_error`

2. **Authentication Failures**
   - 401 responses are converted to `apierror.ErrorResponse` with code `unauthorized`
   - 403 responses are converted to `apierror.ErrorResponse` with code `forbidden`

3. **Service Unavailability**
   - 5xx responses are converted to `apierror.ErrorResponse` with code `server_error`
   - Default behavior does not include automatic retries, but clients can implement this

4. **Rate Limiting**
   - 429 responses are converted to `apierror.ErrorResponse` with code `rate_limited`
   - No built-in backoff, client applications should implement

5. **Context Cancellation**
   - All API calls accept a context.Context parameter
   - Allows client applications to implement timeouts and cancellation

## 4. Data Exchange

### 4.1 Shared Data Models

The SDK defines Go struct representations of the following shared data models:

1. **Authentication Models**
   - `TokenResponse`: Contains access token, token type, and expiration
   - `ClientCredentialResponse`: Client credential details
   - `UserProfileResponse`: User profile information

2. **AI Prompt Models**
   - `Prompt`: Complete prompt configuration
   - `PromptVariable`: Variable definition for prompt templates

3. **Storage Models**
   - `GenerateUploadURLResponse`: Pre-signed URL for uploads
   - `GenerateDownloadURLResponse`: Pre-signed URL for downloads

4. **Ingest Models**
   - `ContentItem`: Metadata about ingested content
   - `IngestResponse`: Response to content ingestion requests
   - `ListContentResponse`: Paginated list of content items

5. **Error Models**
   - `apierror.ErrorResponse`: Standardized error format

### 4.2 Serialization Formats

The SDK uses the following serialization formats:

1. **JSON**
   - Used for all HTTP request and response bodies
   - Implemented using Go's `encoding/json` package
   - All structs have appropriate json tags

2. **HTTP Headers**
   - Standard HTTP headers for authentication and content negotiation
   - Content-Type: application/json
   - Accept: application/json
   - Authorization: Bearer {token}

3. **Query Parameters**
   - Used for filtering, pagination, and search operations
   - URL-encoded according to RFC 3986

4. **Binary Data**
   - Used for file uploads and downloads
   - Transferred directly to/from S3 via pre-signed URLs

### 4.3 Schema Evolution Approach

The SDK follows these principles for handling schema evolution:

1. **Backward Compatibility**
   - Optional fields use pointer types to distinguish between zero values and unset fields
   - New fields are always added as optional
   - Field removal follows a deprecation period

2. **Versioning Strategy**
   - The SDK uses semantic versioning (SemVer)
   - Breaking changes are limited to major version increments
   - API endpoints are versioned separately from the SDK

3. **Extensibility Patterns**
   - Functional options pattern for client configuration
   - Maps for extensible metadata (`map[string]string`)
   - Interface-based design for pluggable components (e.g., `TokenProvider`)

4. **Field Type Flexibility**
   - Some complex fields use `map[string]interface{}` for flexibility
   - Custom JSON unmarshaling for fields with multiple possible types

## 5. Integration Sequence Diagrams

### Authentication Flow

```
┌──────────┐                  ┌──────────┐                ┌──────────────┐
│  Client  │                  │  SDK     │                │  Auth API    │
│ App      │                  │          │                │              │
└────┬─────┘                  └────┬─────┘                └──────┬───────┘
     │                             │                             │
     │ Initialize with credentials │                             │
     │ ──────────────────────────>│                             │
     │                             │                             │
     │                             │ OAuth Token Request         │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Access Token Response       │
     │                             │ <──────────────────────────│
     │                             │                             │
     │ SDK Client ready            │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
     │ Call API method             │                             │
     │ ──────────────────────────>│                             │
     │                             │ API Request with token      │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ API Response                │
     │                             │ <──────────────────────────│
     │ API Method Response         │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
```

### File Upload Flow

```
┌──────────┐                  ┌──────────┐            ┌─────────────┐            ┌──────────┐
│  Client  │                  │  SDK     │            │ Ingest API  │            │   S3     │
│  App     │                  │          │            │             │            │          │
└────┬─────┘                  └────┬─────┘            └──────┬──────┘            └────┬─────┘
     │                             │                         │                        │
     │ RequestFileUpload()         │                         │                        │
     │ ──────────────────────────>│                         │                        │
     │                             │ POST /ingest/file       │                        │
     │                             │ ──────────────────────>│                        │
     │                             │                         │                        │
     │                             │ Pre-signed URL          │                        │
     │                             │ <──────────────────────│                        │
     │                             │                         │                        │
     │ Pre-signed URL Response     │                         │                        │
     │ <──────────────────────────│                         │                        │
     │                             │                         │                        │
     │ UploadToURL()               │                         │                        │
     │ ──────────────────────────>│                         │                        │
     │                             │ PUT to pre-signed URL   │                        │
     │                             │ ───────────────────────────────────────────────>│
     │                             │                         │                        │
     │                             │ Upload Success Response │                        │
     │                             │ <──────────────────────────────────────────────│
     │ Upload Result               │                         │                        │
     │ <──────────────────────────│                         │                        │
     │                             │                         │                        │
```

### URL Ingestion Flow (Asynchronous)

```
┌──────────┐                  ┌──────────┐                ┌──────────────┐
│  Client  │                  │  SDK     │                │  Ingest API  │
│  App     │                  │          │                │              │
└────┬─────┘                  └────┬─────┘                └──────┬───────┘
     │                             │                             │
     │ IngestURL()                 │                             │
     │ ──────────────────────────>│                             │
     │                             │ POST /ingest/url            │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Accepted (202) PENDING      │
     │                             │ <──────────────────────────│
     │ IngestURLResponse           │                             │
     │ (PENDING with ID)           │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
     │                             │ [Asynchronous Processing]   │
     │                             │                             │
     │ GetContentItem()            │                             │
     │ ──────────────────────────>│                             │
     │                             │ GET /content/{id}           │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Content Status (PROCESSING) │
     │                             │ <──────────────────────────│
     │ ContentItem                 │                             │
     │ (status = PROCESSING)       │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
     │ [Client waits and retries]  │                             │
     │                             │                             │
     │ GetContentItem()            │                             │
     │ ──────────────────────────>│                             │
     │                             │ GET /content/{id}           │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Content Status (COMPLETED)  │
     │                             │ <──────────────────────────│
     │ ContentItem                 │                             │
     │ (status = COMPLETED)        │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
```

### Prompt Management Flow

```
┌──────────┐                  ┌──────────┐                ┌──────────────┐
│  Client  │                  │  SDK     │                │  AI API      │
│  App     │                  │          │                │              │
└────┬─────┘                  └────┬─────┘                └──────┬───────┘
     │                             │                             │
     │ CreatePrompt()              │                             │
     │ ──────────────────────────>│                             │
     │                             │ POST /prompts               │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Created Prompt Response     │
     │                             │ <──────────────────────────│
     │ Prompt Object               │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
     │ UpdatePrompt()              │                             │
     │ ──────────────────────────>│                             │
     │                             │ PUT /prompts/{id}           │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Updated Prompt Response     │
     │                             │ <──────────────────────────│
     │ Updated Prompt Object       │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
     │ ListPrompts()               │                             │
     │ ──────────────────────────>│                             │
     │                             │ GET /prompts?filters        │
     │                             │ ──────────────────────────>│
     │                             │                             │
     │                             │ Prompt List + Pagination    │
     │                             │ <──────────────────────────│
     │ Prompt List + Next Token    │                             │
     │ <──────────────────────────│                             │
     │                             │                             │
```