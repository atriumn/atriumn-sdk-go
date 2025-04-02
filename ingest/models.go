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