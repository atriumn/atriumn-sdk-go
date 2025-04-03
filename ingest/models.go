// Package ingest provides a Go client for interacting with the Atriumn Ingest API.
package ingest

import "fmt"

// IngestTextRequest represents a request to ingest text content
type IngestTextRequest struct {
	TenantID string `json:"tenantId,omitempty"`
	UserID   string `json:"userId,omitempty"`
	Content  string `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IngestURLRequest represents a request to ingest content from a URL
type IngestURLRequest struct {
	TenantID string `json:"tenantId,omitempty"`
	UserID   string `json:"userId,omitempty"`
	URL      string `json:"url"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IngestFileRequest represents a request to ingest content from a file
// This is not directly used in JSON marshaling/unmarshaling but represents
// the form data fields that will be sent in the multipart/form-data request
type IngestFileRequest struct {
	TenantID string
	UserID   string
	Filename string
	Metadata map[string]string
}

// IngestResponse represents the response from the ingest endpoints
type IngestResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	TenantID   string `json:"tenantId"`
	UserID     string `json:"userId,omitempty"`
	Size       int64  `json:"size,omitempty"`
	ContentURI string `json:"contentUri,omitempty"`
	Timestamp  string `json:"timestamp"`
}

// ContentItem represents a content item returned by the API
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

// ListContentResponse represents the response from the GET /content endpoint
type ListContentResponse struct {
	Items     []ContentItem `json:"items"`
	NextToken string        `json:"nextToken,omitempty"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	ErrorCode   string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// Error satisfies the error interface
func (e *ErrorResponse) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Description)
	}
	return e.ErrorCode
} 