package storage

import "fmt"

// GenerateUploadURLRequest defines the request body for generating an upload URL.
type GenerateUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// GenerateUploadURLResponse defines the successful response body for generating an upload URL.
type GenerateUploadURLResponse struct {
	UploadURL  string `json:"uploadUrl"`
	HTTPMethod string `json:"httpMethod"` // Expected: "PUT"
	// Add other fields if the API returns them, e.g., required headers
}

// GenerateDownloadURLRequest defines the request body for generating a download URL.
type GenerateDownloadURLRequest struct {
	S3Key string `json:"s3Key"` // Full S3 key including tenant prefix
}

// GenerateDownloadURLResponse defines the successful response body for generating a download URL.
type GenerateDownloadURLResponse struct {
	DownloadURL string `json:"downloadUrl"`
	HTTPMethod  string `json:"httpMethod"` // Expected: "GET"
}

// ErrorResponse represents a standard error response from the storage API.
type ErrorResponse struct {
	ErrorCode   string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// Error satisfies the error interface.
func (e *ErrorResponse) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Description)
	}
	return e.ErrorCode
}
