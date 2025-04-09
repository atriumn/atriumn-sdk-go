// Package ingest provides a Go client for interacting with the Atriumn Ingest API.
// It enables uploading and managing various types of content (text, URLs, files)
// through a simple, idiomatic Go interface.
package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/atriumn/atriumn-sdk-go/internal/clientutil"
)

const (
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 10 * time.Second

	// DefaultUserAgent is the user agent sent in requests
	DefaultUserAgent = "atriumn-ingest-client/1.0"
)

// TokenProvider defines an interface for retrieving authentication tokens.
// Implementations should retrieve and return valid bearer tokens for the Atriumn API.
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error) // Returns the Bearer token string
}

// Client is the main API client for Atriumn Ingest Service.
// It handles communication with the API endpoints for content ingestion
// and retrieval operations.
type Client struct {
	// BaseURL is the base URL of the Atriumn Ingest API
	BaseURL *url.URL

	// HTTPClient is the HTTP client used for making requests
	HTTPClient *http.Client

	// UserAgent is the user agent sent with each request
	UserAgent string

	// tokenProvider provides authentication tokens for API requests
	tokenProvider TokenProvider
}

// NewClient creates a new Atriumn Ingest API client with the specified base URL.
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

// IngestText ingests text content through the Atriumn Ingest API.
// The request parameter contains the text to ingest along with metadata.
// Returns an ingest response with details about the ingested content or an error.
//
// Deprecated: This method is incompatible with the new upload model. Use RequestTextUpload to get a pre-signed URL, 
// then perform an HTTP PUT request directly to that URL with the text content.
func (c *Client) IngestText(ctx context.Context, request *IngestTextRequest) (*IngestResponse, error) {
	httpReq, err := c.newRequest(ctx, "POST", "/ingest/text", request)
	if err != nil {
		return nil, err
	}

	var resp IngestResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// IngestURL ingests content from a URL through the Atriumn Ingest API.
// The request parameter contains the URL to ingest along with metadata.
// Returns an asynchronous response with the ID and status (PENDING/QUEUED) of the ingestion job.
func (c *Client) IngestURL(ctx context.Context, request *IngestURLRequest) (*IngestURLResponse, error) {
	httpReq, err := c.newRequest(ctx, "POST", "/ingest/url", request)
	if err != nil {
		return nil, err
	}

	var resp IngestURLResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// IngestFile ingests content from a file through the Atriumn Ingest API.
// Parameters include tenant ID, filename, content type, user ID, and a reader
// providing the file content to be uploaded.
// Returns an ingest response with details about the ingested file or an error.
//
// Deprecated: This method uses the old single-step multipart/form-data upload pattern
// which is no longer supported by the refactored ingest service endpoint (/ingest/file).
// Use RequestFileUpload to get a pre-signed URL, then perform an HTTP PUT request
// directly to that URL with the file content.
func (c *Client) IngestFile(ctx context.Context, tenantID string, filename string, contentType string, userID string, fileReader io.Reader) (*IngestResponse, error) {
	// Create multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	if tenantID != "" {
		if err := writer.WriteField("tenantId", tenantID); err != nil {
			return nil, fmt.Errorf("failed to write tenantId field: %w", err)
		}
	}

	if userID != "" {
		if err := writer.WriteField("userId", userID); err != nil {
			return nil, fmt.Errorf("failed to write userId field: %w", err)
		}
	}

	// Create form file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content to form file
	if _, err := io.Copy(part, fileReader); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	u := c.BaseURL.JoinPath("ingest", "file")

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

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

	// Send request and process response
	var resp IngestResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// RequestFileUpload initiates a file upload by sending metadata to the ingest service.
// It returns the pre-signed URL required for the client to upload the file directly to S3.
func (c *Client) RequestFileUpload(ctx context.Context, request *RequestFileUploadRequest) (*RequestFileUploadResponse, error) {
	// Use the internal newRequest helper to create the POST request
	// The path should now be `/ingest/file` based on service refactor. Double-check service route.
	httpReq, err := c.newRequest(ctx, "POST", "/ingest/file", request) // Pass the RequestFileUploadRequest struct directly
	if err != nil {
		return nil, fmt.Errorf("failed to create file upload request: %w", err)
	}

	// Execute the request using the internal 'do' helper, expecting RequestFileUploadResponse
	var resp RequestFileUploadResponse
	_, err = c.do(httpReq, &resp) // Pass pointer to the response struct
	if err != nil {
		return nil, err // Error handling (including 4xx/5xx) is done within c.do
	}

	// Return the successful response
	return &resp, nil
}

// RequestTextUpload initiates a text upload by sending metadata to the ingest service.
// It returns the pre-signed URL required for the client to upload the text content directly to S3.
func (c *Client) RequestTextUpload(ctx context.Context, request *RequestTextUploadRequest) (*RequestTextUploadResponse, error) {
	// Create the POST request to /ingest/text endpoint
	httpReq, err := c.newRequest(ctx, "POST", "/ingest/text", request)
	if err != nil {
		return nil, fmt.Errorf("failed to create text upload request: %w", err)
	}

	// Execute the request and parse the response
	var resp RequestTextUploadResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	// Return the successful response
	return &resp, nil
}

// UploadToURL uploads file content directly to a pre-signed URL.
// This is a utility method that can be used after RequestFileUpload to complete the two-step upload process.
// It handles making the PUT request with the correct headers and returns the HTTP response.
func (c *Client) UploadToURL(ctx context.Context, uploadURL string, contentType string, fileReader io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	// Don't follow redirects for pre-signed URLs
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	
	// Execute request directly without using the c.do method since we want the raw response
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to pre-signed URL: %w", err)
	}

	// Check for successful upload
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return resp, nil
}

// newRequest creates an API request with the specified method, path, and body
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u := c.BaseURL.JoinPath(path)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

// GetContentItem retrieves a content item by its ID.
// Returns the content item details or an error if not found or operation fails.
func (c *Client) GetContentItem(ctx context.Context, id string) (*ContentItem, error) {
	path := fmt.Sprintf("/content/%s", id)
	httpReq, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ContentItem
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ListContentItems retrieves a list of content items with optional filtering.
// Parameters allow filtering by status, source type, and pagination through limit and next token.
// Returns a list of content items or an error if the operation fails.
func (c *Client) ListContentItems(ctx context.Context, statusFilter *string, sourceTypeFilter *string, limit *int, nextToken *string) (*ListContentResponse, error) {
	path := "/content"
	
	// Create query parameters
	q := url.Values{}
	if statusFilter != nil {
		q.Add("status", *statusFilter)
	}
	if sourceTypeFilter != nil {
		q.Add("sourceType", *sourceTypeFilter)
	}
	if limit != nil {
		q.Add("limit", strconv.Itoa(*limit))
	}
	if nextToken != nil {
		q.Add("nextToken", *nextToken)
	}

	// Create request
	httpReq, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters to the URL
	httpReq.URL.RawQuery = q.Encode()

	var resp ListContentResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
} 