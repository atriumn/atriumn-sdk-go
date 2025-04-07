// Package storage provides a Go client for interacting with the Atriumn Storage API.
// It enables generating pre-signed URLs for uploading and downloading files
// through a simple, idiomatic Go interface.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 10 * time.Second

	// DefaultUserAgent is the user agent sent in requests
	DefaultUserAgent = "atriumn-storage-client/1.0"
)

// TokenProvider defines an interface for retrieving authentication tokens.
// Implementations should retrieve and return valid bearer tokens for the Atriumn API.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error) // Returns the Bearer token string
}

// Client is the main API client for Atriumn Storage Service.
// It handles communication with the API endpoints for generating
// pre-signed URLs for file uploads and downloads.
type Client struct {
	// BaseURL is the base URL of the Atriumn Storage API
	BaseURL *url.URL

	// HTTPClient is the HTTP client used for making requests
	HTTPClient *http.Client

	// UserAgent is the user agent sent with each request
	UserAgent string

	// tokenProvider provides authentication tokens for API requests
	tokenProvider TokenProvider
}

// NewClient creates a new Atriumn Storage API client with the specified base URL.
// It returns an error if the provided URL cannot be parsed.
func NewClient(baseURL string) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		BaseURL:    parsedURL,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
		UserAgent:  DefaultUserAgent,
	}, nil
}

// ClientOption is a function that configures a Client.
// It is used with NewClientWithOptions to customize the client behavior.
type ClientOption func(*Client)

// WithHTTPClient sets the HTTP client for the API client.
// This can be used to customize timeouts, transport settings, or to inject
// middleware/interceptors for testing or monitoring.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithUserAgent sets the user agent for the API client.
// This string is sent with each request to identify the client.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithTokenProvider sets the token provider for the API client.
// The token provider is used to obtain authentication tokens for API requests.
func WithTokenProvider(tp TokenProvider) ClientOption {
	return func(c *Client) {
		c.tokenProvider = tp
	}
}

// NewClientWithOptions creates a new client with custom options.
// It allows for flexible configuration of the client through functional options.
// Returns an error if the base URL is invalid.
func NewClientWithOptions(baseURL string, options ...ClientOption) (*Client, error) {
	client, err := NewClient(baseURL)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

// newRequest creates an API request with the specified method, path, and body
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	// Add Authorization header if TokenProvider is configured
	if c.tokenProvider != nil {
		token, tokenErr := c.tokenProvider.GetToken(ctx)
		if tokenErr != nil {
			return nil, fmt.Errorf("failed to get token from provider: %w", tokenErr)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return req, nil
}

// do sends an API request and returns the API response
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// Handle network-level errors
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				return nil, &ErrorResponse{
					ErrorCode:   "request_timeout",
					Description: "The request timed out. Please check your network connection and try again.",
				}
			} else if urlErr.Temporary() {
				return nil, &ErrorResponse{
					ErrorCode:   "temporary_error",
					Description: "A temporary network error occurred. Please try again later.",
				}
			}
		}
		return nil, &ErrorResponse{
			ErrorCode:   "network_error",
			Description: fmt.Sprintf("Failed to connect to the storage service: %v", err),
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	// Reset the body with a new ReadCloser for further processing
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse

		// We already have the body bytes, no need to read again
		if len(bodyBytes) > 0 {
			if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr != nil {
				// Not valid JSON or unexpected format
				return nil, &ErrorResponse{
					ErrorCode:   "parse_error",
					Description: fmt.Sprintf("HTTP error %d with invalid response format", resp.StatusCode),
				}
			}
		}

		// If we received an empty error response, create a user-friendly error based on status code
		if errResp.ErrorCode == "" && errResp.Description == "" {
			switch resp.StatusCode {
			case 400:
				errResp.ErrorCode = "bad_request"
				errResp.Description = "The request was invalid. Please check your input and try again."
			case 401:
				errResp.ErrorCode = "unauthorized"
				errResp.Description = "Authentication failed. Please check your credentials or login again."
			case 403:
				errResp.ErrorCode = "forbidden"
				errResp.Description = "You don't have permission to access this resource."
			case 404:
				errResp.ErrorCode = "not_found"
				errResp.Description = "The requested resource was not found."
			case 429:
				errResp.ErrorCode = "rate_limited"
				errResp.Description = "Too many requests. Please try again later."
			case 500, 502, 503, 504:
				errResp.ErrorCode = "server_error"
				errResp.Description = "The storage service is currently unavailable. Please try again later."
			default:
				errResp.ErrorCode = "unknown_error"
				errResp.Description = fmt.Sprintf("Unexpected HTTP status: %d", resp.StatusCode)
			}
		}

		return nil, &errResp
	}

	if v != nil {
		// We already have the body bytes, decode from there
		err = json.Unmarshal(bodyBytes, v)
		if err != nil {
			return nil, &ErrorResponse{
				ErrorCode:   "parse_error",
				Description: fmt.Sprintf("Failed to parse the successful response: %v", err),
			}
		}
	}

	return resp, nil
}

// GenerateUploadURL generates a pre-signed URL for uploading a file to storage.
// The request parameter contains the filename and content type of the file to upload.
// Returns a response with the upload URL and HTTP method, or an error if generation fails.
func (c *Client) GenerateUploadURL(ctx context.Context, request *GenerateUploadURLRequest) (*GenerateUploadURLResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/generate-upload-url", request)
	if err != nil {
		return nil, err
	}

	var resp GenerateUploadURLResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GenerateDownloadURL generates a pre-signed URL for downloading a file from storage.
// The request parameter contains the S3 key of the file to download.
// Returns a response with the download URL and HTTP method, or an error if generation fails.
func (c *Client) GenerateDownloadURL(ctx context.Context, request *GenerateDownloadURLRequest) (*GenerateDownloadURLResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/generate-download-url", request)
	if err != nil {
		return nil, err
	}

	var resp GenerateDownloadURLResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GenerateDownloadURLFromKey generates a pre-signed URL for downloading a file using just the S3 key.
// This is a convenience method that wraps GenerateDownloadURL.
// Returns a response with the download URL and HTTP method, or an error if generation fails.
func (c *Client) GenerateDownloadURLFromKey(ctx context.Context, s3Key string) (*GenerateDownloadURLResponse, error) {
	request := &GenerateDownloadURLRequest{
		S3Key: s3Key,
	}
	return c.GenerateDownloadURL(ctx, request)
}
