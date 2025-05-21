// Package clientutil provides common utilities for Atriumn API clients.
// It includes shared HTTP request execution and error handling functionality.
package clientutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/atriumn/atriumn-sdk-go/internal/apierror"
)

// ExecuteRequest sends an API request and returns the API response.
// It handles:
// - Sending the request using httpClient.Do(req)
// - Network error handling and wrapping into apierror.ErrorResponse
// - Reading the response body exactly once
// - Closing the response body
// - Status code checking
// - Parsing error responses into apierror.ErrorResponse
// - Generating fallback error messages for empty/unparsable error responses
// - Unmarshalling successful responses into the provided value
func ExecuteRequest(ctx context.Context, httpClient *http.Client, req *http.Request, v interface{}) (*http.Response, error) {
	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		// Handle network-level errors
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				return nil, &apierror.ErrorResponse{
					ErrorCode:   "request_timeout",
					Description: "The request timed out. Please check your network connection and try again.",
				}
			} else if urlErr.Temporary() {
				return nil, &apierror.ErrorResponse{
					ErrorCode:   "temporary_error",
					Description: "A temporary network error occurred. Please try again later.",
				}
			}
		}
		return nil, &apierror.ErrorResponse{
			ErrorCode:   "network_error",
			Description: fmt.Sprintf("Failed to connect to the service: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, &apierror.ErrorResponse{
			ErrorCode:   "read_error",
			Description: fmt.Sprintf("Failed to read response body: %v", err),
		}
	}

	// Reset the body with a new ReadCloser for further processing if needed
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Handle non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp apierror.ErrorResponse

		// Try to unmarshal the error response
		if len(bodyBytes) > 0 {
			if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil &&
				(errResp.ErrorCode != "" || errResp.Description != "") {
				// Successfully parsed error with at least some data
				return nil, &errResp
			}
		}

		// Create a user-friendly error based on status code if parsing failed
		// or the error response was empty
		switch resp.StatusCode {
		case http.StatusBadRequest:
			errResp.ErrorCode = "bad_request"
			errResp.Description = "The request was invalid. Please check your input and try again."
		case http.StatusUnauthorized:
			errResp.ErrorCode = "unauthorized"
			errResp.Description = "Authentication failed. Please check your credentials or login again."
		case http.StatusForbidden:
			errResp.ErrorCode = "forbidden"
			errResp.Description = "You don't have permission to access this resource."
		case http.StatusNotFound:
			errResp.ErrorCode = "not_found"
			errResp.Description = "The requested resource was not found."
		case http.StatusTooManyRequests:
			errResp.ErrorCode = "rate_limited"
			errResp.Description = "Too many requests. Please try again later."
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			errResp.ErrorCode = "server_error"
			errResp.Description = "The service is currently unavailable. Please try again later."
		default:
			errResp.ErrorCode = "unknown_error"
			errResp.Description = fmt.Sprintf("Unexpected HTTP status: %d", resp.StatusCode)
		}

		// Include response body for unknown errors if available
		if errResp.ErrorCode == "unknown_error" && len(bodyBytes) > 0 {
			errResp.Description += fmt.Sprintf(" Body: %s", string(bodyBytes))
		}

		return nil, &errResp
	}

	// Handle successful response
	if v != nil && len(bodyBytes) > 0 {
		err = json.Unmarshal(bodyBytes, v)
		if err != nil {
			return nil, &apierror.ErrorResponse{
				ErrorCode:   "parse_error",
				Description: fmt.Sprintf("Failed to parse the successful response: %v", err),
			}
		}
	}

	return resp, nil
}
