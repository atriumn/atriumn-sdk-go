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

	"github.com/atriumn/atriumn-sdk-go/internal/clientutil"
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
//
// Parameters:
//   - baseURL: The base URL for the Atriumn Storage API (required)
//
// Returns:
//   - *Client: A configured Storage client instance
//   - error: An error if the URL cannot be parsed
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
//
// Parameters:
//   - httpClient: The custom HTTP client to use for making API requests
//
// Returns:
//   - ClientOption: A functional option to configure the client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithUserAgent sets the user agent for the API client.
// This string is sent with each request to identify the client.
//
// Parameters:
//   - userAgent: The user agent string to send with API requests
//
// Returns:
//   - ClientOption: A functional option to configure the client
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithTokenProvider sets the token provider for the API client.
// The token provider is used to obtain authentication tokens for API requests.
//
// Parameters:
//   - tp: The TokenProvider implementation to use for authentication
//
// Returns:
//   - ClientOption: A functional option to configure the client
func WithTokenProvider(tp TokenProvider) ClientOption {
	return func(c *Client) {
		c.tokenProvider = tp
	}
}

// NewClientWithOptions creates a new client with custom options.
// It allows for flexible configuration of the client through functional options.
//
// Parameters:
//   - baseURL: The base URL for the Atriumn Storage API (required)
//   - options: A variadic list of ClientOption functions to customize the client
//
// Returns:
//   - *Client: A configured Storage client instance
//   - error: An error if the URL cannot be parsed
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
	u := c.BaseURL.JoinPath(path)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
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
	return clientutil.ExecuteRequest(req.Context(), c.HTTPClient, req, v)
}

// GenerateUploadURL generates a pre-signed URL for uploading a file to storage.
//
// Parameters:
//   - ctx: Context for the API request
//   - request: GenerateUploadURLRequest containing file metadata (required fields: Filename, ContentType)
//
// Returns:
//   - *GenerateUploadURLResponse: The response containing the pre-signed URL for upload
//   - error: An error if the operation fails, which can be:
//   - apierror.ErrorResponse with codes like:
//   - "bad_request" if the request is invalid
//   - "unauthorized" if authentication fails
//   - "forbidden" if the caller lacks permissions
//   - "network_error" if the connection fails
//   - "server_error" if generating the upload URL fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - request: GenerateDownloadURLRequest containing the S3 key of the file (required field: S3Key)
//
// Returns:
//   - *GenerateDownloadURLResponse: The response containing the pre-signed URL for download
//   - error: An error if the operation fails, which can be:
//   - apierror.ErrorResponse with codes like:
//   - "bad_request" if the request is invalid
//   - "unauthorized" if authentication fails
//   - "forbidden" if the caller lacks permissions
//   - "not_found" if the file doesn't exist
//   - "network_error" if the connection fails
//   - "server_error" if generating the download URL fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - s3Key: The S3 key of the file to download (required)
//
// Returns:
//   - *GenerateDownloadURLResponse: The response containing the pre-signed URL for download
//   - error: An error if the operation fails, which can be:
//   - apierror.ErrorResponse with codes like:
//   - "bad_request" if the key is invalid
//   - "unauthorized" if authentication fails
//   - "forbidden" if the caller lacks permissions
//   - "not_found" if the file doesn't exist
//   - "network_error" if the connection fails
//   - "server_error" if generating the download URL fails
func (c *Client) GenerateDownloadURLFromKey(ctx context.Context, s3Key string) (*GenerateDownloadURLResponse, error) {
	request := &GenerateDownloadURLRequest{
		S3Key: s3Key,
	}
	return c.GenerateDownloadURL(ctx, request)
}
