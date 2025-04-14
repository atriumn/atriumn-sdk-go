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
	"os"
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
//
// Parameters:
//   - baseURL: The base URL for the Atriumn Ingest API (required)
//
// Returns:
//   - *Client: A configured Ingest client instance
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
//   - baseURL: The base URL for the Atriumn Ingest API (required)
//   - options: A variadic list of ClientOption functions to customize the client
//
// Returns:
//   - *Client: A configured Ingest client instance
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

// IngestText ingests text content through the Atriumn Ingest API.
//
// Deprecated: This method is incompatible with the new upload model. Use RequestTextUpload to get a pre-signed URL,
// then perform an HTTP PUT request directly to that URL with the text content.
//
// Parameters:
//   - ctx: Context for the API request
//   - request: IngestTextRequest containing the text content and metadata (required)
//
// Returns:
//   - *IngestResponse: Details about the ingested content if successful
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - request: IngestURLRequest containing the URL to scrape and metadata (required)
//
// Returns:
//   - *IngestURLResponse: An asynchronous response with ID and status (PENDING/QUEUED)
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the URL is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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
//
// Deprecated: This method uses the old single-step multipart/form-data upload pattern
// which is no longer supported by the refactored ingest service endpoint (/ingest/file).
// Use RequestFileUpload to get a pre-signed URL, then perform an HTTP PUT request
// directly to that URL with the file content.
//
// Parameters:
//   - ctx: Context for the API request
//   - tenantID: Optional identifier for multi-tenant applications
//   - filename: The name of the file being uploaded (required)
//   - contentType: The MIME type of the file (required)
//   - userID: Optional identifier for the user who owns this content
//   - fileReader: An io.Reader providing the file content (required)
//
// Returns:
//   - *IngestResponse: Details about the ingested file if successful
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
//       - "parse_error" if there's an issue with processing the file
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
//
// Parameters:
//   - ctx: Context for the API request
//   - request: RequestFileUploadRequest containing file metadata (required fields: Filename, ContentType)
//
// Returns:
//   - *RequestFileUploadResponse: The response containing the pre-signed URL for direct S3 upload
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
//       - "server_error" if generating the upload URL fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - request: RequestTextUploadRequest containing text metadata (required field: ContentType)
//
// Returns:
//   - *RequestTextUploadResponse: The response containing the pre-signed URL for direct S3 upload
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
//       - "server_error" if generating the upload URL fails
func (c *Client) RequestTextUpload(ctx context.Context, request *RequestTextUploadRequest) (*RequestTextUploadResponse, error) {
	httpReq, err := c.newRequest(ctx, "POST", "/ingest/text", request)
	if err != nil {
		return nil, fmt.Errorf("failed to create text upload request: %w", err)
	}

	var resp RequestTextUploadResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UploadToURL uploads content directly to a pre-signed URL.
//
// Parameters:
//   - ctx: Context for the API request
//   - uploadURL: The pre-signed S3 URL to upload to (required)
//   - contentType: The MIME type of the content being uploaded (required)
//   - fileReader: An io.Reader providing the content to upload (required)
//
// Returns:
//   - *http.Response: The raw HTTP response from the upload operation
//   - error: An error if the upload fails, which can include:
//     * Network errors if the connection fails
//     * S3-specific errors if the upload is rejected
//     * Context cancellation errors
func (c *Client) UploadToURL(ctx context.Context, uploadURL string, contentType string, fileReader io.Reader) (*http.Response, error) {
	// Create a new HTTP request with the provided upload URL
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set the Content-Type header to the specified value
	req.Header.Set("Content-Type", contentType)

	// Set Content-Length if we can determine it from the fileReader (if it's an *os.File)
	if file, ok := fileReader.(*os.File); ok {
		fileInfo, err := file.Stat()
		if err == nil {
			req.ContentLength = fileInfo.Size()
		}
	}

	// Use the standard HTTP client instead of c.HTTPClient to avoid auth header conflicts
	// for direct S3 uploads with pre-signed URLs
	standardClient := &http.Client{
		Timeout: 60 * time.Second, // Set a reasonable timeout
	}
	
	resp, err := standardClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to URL: %w", err)
	}

	// Check for non-2xx status codes and return appropriate error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("upload failed with status %d, and failed to read error response: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// newRequest creates an API request with the specified method, path and body
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

// GetContentItem retrieves a specific content item by its ID.
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the content item to retrieve (required)
//
// Returns:
//   - *ContentItem: The content item details if found
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the content item doesn't exist
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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

// ListContentItems lists content items with optional filters.
//
// Parameters:
//   - ctx: Context for the API request
//   - statusFilter: Optional filter to match content items with a specific status (e.g., "COMPLETED")
//   - sourceTypeFilter: Optional filter to match content items with a specific source type (e.g., "TEXT", "URL", "FILE")
//   - limit: Optional maximum number of items to return
//   - nextToken: Optional pagination token from a previous list response
//
// Returns:
//   - *ListContentResponse: A list of content items and optional pagination token
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the query parameters are invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
func (c *Client) ListContentItems(ctx context.Context, statusFilter *string, sourceTypeFilter *string, limit *int, nextToken *string) (*ListContentResponse, error) {
	httpReq, err := c.newRequest(ctx, "GET", "/content", nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters if they are provided
	q := httpReq.URL.Query()
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
	httpReq.URL.RawQuery = q.Encode()

	var resp ListContentResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetContentDownloadURL retrieves a pre-signed URL that can be used to download the content.
//
// Parameters:
//   - ctx: Context for the API request
//   - contentID: The unique identifier of the content item (required)
//
// Returns:
//   - *DownloadURLResponse: Contains the pre-signed download URL if successful
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the content doesn't exist
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
func (c *Client) GetContentDownloadURL(ctx context.Context, contentID string) (*DownloadURLResponse, error) {
	path := fmt.Sprintf("/content/%s/download-url", contentID)
	
	req, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp DownloadURLResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateContentItem updates a content item's metadata.
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the content item to update (required)
//   - req: UpdateContentItemRequest containing the fields to update (required)
//
// Returns:
//   - *ContentItem: The updated content item if successful
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the content item doesn't exist
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
func (c *Client) UpdateContentItem(ctx context.Context, id string, req *UpdateContentItemRequest) (*ContentItem, error) {
	path := fmt.Sprintf("/content/%s", id)
	httpReq, err := c.newRequest(ctx, "PATCH", path, req)
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

// DeleteContentItem deletes a content item by its ID.
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the content item to delete (required)
//
// Returns:
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the content item doesn't exist
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
func (c *Client) DeleteContentItem(ctx context.Context, id string) error {
	path := fmt.Sprintf("/content/%s", id)
	httpReq, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = c.do(httpReq, nil)
	return err
}

// GetTextContent retrieves the raw text content of a TEXT type content item.
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the content item to retrieve text from (required)
//
// Returns:
//   - *GetTextContentResponse: Contains the raw text content if successful
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the content item doesn't exist
//       - "bad_request" if the content item is not of type TEXT
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
func (c *Client) GetTextContent(ctx context.Context, id string) (*GetTextContentResponse, error) {
	path := fmt.Sprintf("/content/%s/text", id)
	httpReq, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetTextContentResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateTextContent updates the raw text content of a TEXT type content item.
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the content item to update (required)
//   - req: UpdateTextContentRequest containing the new text content (required)
//
// Returns:
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the content item doesn't exist
//       - "bad_request" if the content item is not of type TEXT
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
func (c *Client) UpdateTextContent(ctx context.Context, id string, req *UpdateTextContentRequest) error {
	path := fmt.Sprintf("/content/%s/text", id)
	httpReq, err := c.newRequest(ctx, "PUT", path, req)
	if err != nil {
		return err
	}

	_, err = c.do(httpReq, nil)
	return err
} 