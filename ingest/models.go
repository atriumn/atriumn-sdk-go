// Package ingest provides a Go client for interacting with the Atriumn Ingest API.
// It enables uploading and managing various types of content (text, URLs, files)
// through a simple, idiomatic Go interface.
package ingest

// IngestTextRequest represents a request to ingest text content.
// It contains the text content to be ingested along with optional
// tenant ID, user ID, and metadata.
type IngestTextRequest struct {
	TenantID string `json:"tenantId,omitempty"`
	UserID   string `json:"userId,omitempty"`
	Content  string `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IngestURLRequest represents a request to ingest content from a URL.
// It contains the URL to be scraped and ingested along with optional
// tenant ID, user ID, and metadata.
type IngestURLRequest struct {
	TenantID string `json:"tenantId,omitempty"`
	UserID   string `json:"userId,omitempty"`
	URL      string `json:"url"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IngestFileRequest represents a request to ingest content from a file.
// This is not directly used in JSON marshaling/unmarshaling but represents
// the form data fields that will be sent in the multipart/form-data request.
// It contains metadata about the file being uploaded.
type IngestFileRequest struct {
	TenantID string
	UserID   string
	Filename string
	Metadata map[string]string
}

// IngestResponse represents the response from the ingest endpoints.
// It contains details about the ingested content, including its unique ID,
// processing status, and associated metadata.
type IngestResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	TenantID   string `json:"tenantId"`
	UserID     string `json:"userId,omitempty"`
	Size       int64  `json:"size,omitempty"`
	ContentURI string `json:"contentUri,omitempty"`
	Timestamp  string `json:"timestamp"`
}

// ContentItem represents a content item returned by the API.
// It contains comprehensive metadata about the ingested content,
// including its source, processing status, and storage information.
type ContentItem struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenantId"`
	UserID      string            `json:"userId,omitempty"`
	SourceType  string            `json:"sourceType"`
	SourceURI   string            `json:"sourceUri,omitempty"`
	S3Key       string            `json:"s3Key,omitempty"`
	Status      string            `json:"status"`
	ContentType string            `json:"contentType,omitempty"`
	Size        int64             `json:"size,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

// ListContentResponse represents the response from the GET /content endpoint.
// It contains a list of content items and an optional token for pagination.
type ListContentResponse struct {
	Items     []ContentItem `json:"items"`
	NextToken string        `json:"nextToken,omitempty"`
}

// ErrorResponse is now provided by the internal/apierror package. 