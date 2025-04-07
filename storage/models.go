// Package storage provides a Go client for interacting with the Atriumn Storage API.
// It enables generating pre-signed URLs for uploading and downloading files
// through a simple, idiomatic Go interface.
package storage

import "fmt"

// GenerateUploadURLRequest defines the request body for generating an upload URL.
// It specifies the filename and content type of the file to be uploaded.
type GenerateUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// GenerateUploadURLResponse defines the successful response body for generating an upload URL.
// It contains the pre-signed URL for uploading a file and the HTTP method to use.
type GenerateUploadURLResponse struct {
	UploadURL  string `json:"uploadUrl"`
	HTTPMethod string `json:"httpMethod"` // Expected: "PUT"
	// Add other fields if the API returns them, e.g., required headers
}

// GenerateDownloadURLRequest defines the request body for generating a download URL.
// It specifies the S3 key of the file to be downloaded.
type GenerateDownloadURLRequest struct {
	S3Key string `json:"s3Key"` // Full S3 key including tenant prefix
}

// GenerateDownloadURLResponse defines the successful response body for generating a download URL.
// It contains the pre-signed URL for downloading a file and the HTTP method to use.
type GenerateDownloadURLResponse struct {
	DownloadURL string `json:"downloadUrl"`
	HTTPMethod  string `json:"httpMethod"` // Expected: "GET"
}

// ErrorResponse represents a standard error response from the storage API.
// It contains the error code and an optional description returned by the Atriumn Storage API.
type ErrorResponse struct {
	ErrorCode   string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// Error satisfies the error interface by returning a formatted error message.
// If a description is available, it will be included in the error message.
func (e *ErrorResponse) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Description)
	}
	return e.ErrorCode
}
