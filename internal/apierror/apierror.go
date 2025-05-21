// Package apierror provides common error handling for Atriumn API clients.
// It defines the standard error response structure used across different Atriumn APIs.
package apierror

import "fmt"

// ErrorResponse represents a standard error response from Atriumn APIs.
// It contains the error code and an optional description returned by the API.
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
