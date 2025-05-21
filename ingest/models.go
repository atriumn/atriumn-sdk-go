// Package ingest provides a Go client for interacting with the Atriumn Ingest API.
// It enables uploading and managing various types of content (text, URLs, files)
// through a simple, idiomatic Go interface.
package ingest

// IngestTextRequest represents a request to ingest text content.
// It contains the text content to be ingested along with optional
// tenant ID, user ID, and metadata.
type IngestTextRequest struct {
	// TenantID is an optional identifier for multi-tenant applications
	TenantID string `json:"tenantId,omitempty"`
	// UserID is an optional identifier for the user who owns this content
	UserID string `json:"userId,omitempty"`
	// Content is the text content to be ingested (required)
	Content string `json:"content"`
	// Metadata is an optional map of key-value pairs with additional information about the content
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IngestURLRequest represents a request to ingest content from a URL.
// It contains the URL to be scraped and ingested along with optional
// tenant ID, user ID, and metadata.
type IngestURLRequest struct {
	// TenantID is an optional identifier for multi-tenant applications
	TenantID string `json:"tenantId,omitempty"`
	// UserID is an optional identifier for the user who owns this content
	UserID string `json:"userId,omitempty"`
	// URL is the web address to scrape and ingest (required)
	URL string `json:"url"`
	// SourceSubType is an optional hint about the nature of the URL (e.g., "linkedin_profile")
	SourceSubType *string `json:"sourceSubType,omitempty"`
	// Metadata is an optional map of key-value pairs with additional information about the content
	Metadata map[string]string `json:"metadata,omitempty"`
	// UserNotes is an optional field containing free-form text notes provided by the user
	UserNotes *string `json:"userNotes,omitempty"`
}

// IngestFileRequest represents a request to ingest content from a file.
// This is not directly used in JSON marshaling/unmarshaling but represents
// the form data fields that will be sent in the multipart/form-data request.
// It contains metadata about the file being uploaded.
type IngestFileRequest struct {
	// TenantID is an optional identifier for multi-tenant applications
	TenantID string
	// UserID is an optional identifier for the user who owns this content
	UserID string
	// Filename is the name of the file being uploaded (required)
	Filename string
	// Metadata is an optional map of key-value pairs with additional information about the file
	Metadata map[string]string
}

// RequestFileUploadRequest represents a request to initiate a file upload session.
// It sends metadata to the ingest service to obtain an upload URL.
type RequestFileUploadRequest struct {
	// Filename is the name of the file to be uploaded (required)
	Filename string `json:"filename"`
	// ContentType is the MIME type of the file (required)
	ContentType string `json:"contentType"`
	// TenantID is an optional identifier for multi-tenant applications
	TenantID string `json:"tenantId,omitempty"`
	// UserID is an optional identifier for the user who owns this content
	UserID string `json:"userId,omitempty"`
	// Metadata is an optional map of key-value pairs with additional information about the file
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RequestFileUploadResponse defines the successful response body after requesting a file upload.
// It contains the pre-signed URL for uploading the file and the unique content item ID.
type RequestFileUploadResponse struct {
	// ContentID is the unique ID assigned to the content item
	ContentID string `json:"id"`
	// Status is the status of the content item (should be UPLOADING)
	Status string `json:"status"`
	// UploadURL is the pre-signed URL to use for the HTTP PUT upload
	UploadURL string `json:"uploadUrl"`
	// TenantID is the tenant ID associated with this upload
	TenantID string `json:"tenantId,omitempty"`
	// UserID is the user ID associated with this upload, if provided
	UserID string `json:"userId,omitempty"`
	// Timestamp is when the request was processed
	Timestamp string `json:"timestamp,omitempty"`
}

// RequestTextUploadRequest represents a request to initiate a text upload session.
// It sends metadata to the ingest service to obtain an upload URL, without the content itself.
type RequestTextUploadRequest struct {
	// TenantID is an optional identifier for multi-tenant applications
	TenantID string `json:"tenantId,omitempty"`
	// ContentType is the MIME type of the text content (optional, defaults to text/plain)
	ContentType string `json:"contentType,omitempty"`
	// UserID is an optional identifier for the user who owns this content
	UserID string `json:"userId,omitempty"`
	// Metadata is an optional map of key-value pairs with additional information
	Metadata map[string]string `json:"metadata,omitempty"`
	// CallbackURL is an optional URL to be notified when processing completes
	CallbackURL string `json:"callbackUrl,omitempty"`
}

// RequestTextUploadResponse defines the successful response body after requesting a text upload.
// It contains the pre-signed URL for uploading the text and the unique content item ID.
type RequestTextUploadResponse struct {
	// ContentID is the unique ID assigned to the content item
	ContentID string `json:"id"`
	// Status is the status of the content item (should be UPLOADING)
	Status string `json:"status"`
	// UploadURL is the pre-signed URL to use for the HTTP PUT upload
	UploadURL string `json:"uploadUrl"`
	// TenantID is the tenant ID associated with this upload
	TenantID string `json:"tenantId,omitempty"`
	// UserID is the user ID associated with this upload, if provided
	UserID string `json:"userId,omitempty"`
	// Timestamp is when the request was processed
	Timestamp string `json:"timestamp,omitempty"`
}

// IngestResponse represents the response from the ingest endpoints.
// It contains details about the ingested content, including its unique ID,
// processing status, and associated metadata.
type IngestResponse struct {
	// ID is the unique identifier for the content item
	ID string `json:"id"`
	// Status indicates the processing status of the content (e.g., "PROCESSING", "COMPLETED")
	Status string `json:"status"`
	// TenantID is the tenant identifier for multi-tenant applications
	TenantID string `json:"tenantId"`
	// UserID is the identifier for the user who owns this content
	UserID string `json:"userId,omitempty"`
	// Size is the content size in bytes
	Size int64 `json:"size,omitempty"`
	// ContentURI is the URI where the processed content can be accessed
	ContentURI string `json:"contentUri,omitempty"`
	// Timestamp is when the content was ingested
	Timestamp string `json:"timestamp"`
}

// ContentItem represents a content item returned by the API.
// It contains comprehensive metadata about the ingested content,
// including its source, processing status, and storage information.
type ContentItem struct {
	// ID is the unique identifier for the content item
	ID string `json:"id"`
	// TenantID is the tenant identifier for multi-tenant applications
	TenantID string `json:"tenantId"`
	// UserID is the identifier for the user who owns this content
	UserID string `json:"userId,omitempty"`
	// SourceType indicates how the content was ingested (e.g., "TEXT", "URL", "FILE")
	SourceType string `json:"sourceType"`
	// SourceURI is the original source URI for URL content
	SourceURI string `json:"sourceUri,omitempty"`
	// S3Key is the internal storage key in S3
	S3Key string `json:"s3Key,omitempty"`
	// S3Bucket is the S3 bucket where the content is stored
	S3Bucket string `json:"s3Bucket,omitempty"`
	// Status indicates the processing status of the content
	Status string `json:"status"`
	// ContentType is the MIME type of the content
	ContentType string `json:"contentType,omitempty"`
	// Size is the content size in bytes
	Size int64 `json:"size,omitempty"`
	// Metadata is a map of custom metadata associated with this content
	Metadata map[string]string `json:"metadata,omitempty"`
	// CreatedAt is the UTC timestamp when the content was created
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the UTC timestamp when the content was last updated
	UpdatedAt string `json:"updatedAt"`
}

// ListContentResponse represents the response from the GET /content endpoint.
// It contains a list of content items and an optional token for pagination.
type ListContentResponse struct {
	// Items is an array of content items matching the query criteria
	Items []ContentItem `json:"items"`
	// NextToken is an optional pagination token for retrieving the next set of results
	NextToken string `json:"nextToken,omitempty"`
}

// ErrorResponse is now provided by the internal/apierror package.

// IngestURLResponse represents the response from the ingest URL endpoint.
// After Task 7.1/7.3 changes, this is an immediate, asynchronous response
// indicating that URL processing has been queued.
type IngestURLResponse struct {
	// ID is the unique identifier assigned to the content item
	ID string `json:"id"`
	// Status should be PENDING/QUEUED, indicating asynchronous processing
	Status string `json:"status"`
}

// DownloadURLResponse represents the response from the GET /content/{id}/download-url endpoint.
// It contains a pre-signed URL that can be used to download the content.
type DownloadURLResponse struct {
	// DownloadURL is the pre-signed URL that can be used to download the content
	DownloadURL string `json:"downloadUrl"`
}

// UpdateContentItemRequest represents the payload for updating a content item.
// It uses pointers for fields that are optional in the update to distinguish
// between empty values and fields not provided for update.
type UpdateContentItemRequest struct {
	// SourceURI is the original source URI for the content
	SourceURI *string `json:"sourceUri,omitempty"`
	// Metadata is an optional map of key-value pairs with additional information about the content
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GetTextContentResponse represents the response from the GET /content/{id}/text endpoint.
// It contains the raw text content of a TEXT type content item.
type GetTextContentResponse struct {
	// Content is the raw text content
	Content string `json:"content"`
}

// UpdateTextContentRequest represents the request to update text content via PUT /content/{id}/text.
// It contains the new text content to be stored.
type UpdateTextContentRequest struct {
	// Content is the new text content to store
	Content string `json:"content"`
}
