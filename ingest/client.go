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
	"path"
	"strconv"
	"time"
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
// Returns an ingest response with details about the ingested content or an error.
func (c *Client) IngestURL(ctx context.Context, request *IngestURLRequest) (*IngestResponse, error) {
	httpReq, err := c.newRequest(ctx, "POST", "/ingest/url", request)
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

// IngestFile ingests content from a file through the Atriumn Ingest API.
// Parameters include tenant ID, filename, content type, user ID, and a reader
// providing the file content to be uploaded.
// Returns an ingest response with details about the ingested file or an error.
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
	u := *c.BaseURL
	u.Path = path.Join(u.Path, "/ingest/file")

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

// newRequest creates an API request with the specified method, path, and body
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u := *c.BaseURL
	u.Path = path

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
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create a new reader for future parsing
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse

		// Try to unmarshal the error response
		if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
			// If we can't parse the error response, create a generic one
			return resp, &ErrorResponse{
				ErrorCode:   "unknown_error",
				Description: fmt.Sprintf("HTTP error %d: %s", resp.StatusCode, string(bodyBytes)),
			}
		}

		// If we got an empty error response, create a formatted error based on status code
		if errResp.ErrorCode == "" {
			switch resp.StatusCode {
			case http.StatusBadRequest:
				errResp.ErrorCode = "bad_request"
				errResp.Description = "The request was invalid. Please check your input and try again."
			case http.StatusUnauthorized:
				errResp.ErrorCode = "unauthorized"
				errResp.Description = "Authentication required. Please provide valid credentials."
			case http.StatusForbidden:
				errResp.ErrorCode = "forbidden"
				errResp.Description = "You don't have permission to access this resource."
			case http.StatusNotFound:
				errResp.ErrorCode = "not_found"
				errResp.Description = "The requested resource was not found."
			case http.StatusTooManyRequests:
				errResp.ErrorCode = "rate_limited"
				errResp.Description = "Too many requests. Please try again later."
			case http.StatusInternalServerError:
				errResp.ErrorCode = "server_error"
				errResp.Description = "An internal server error occurred. Please try again later."
			default:
				errResp.ErrorCode = "unknown_error"
				errResp.Description = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
			}
		}

		return resp, &errResp
	}

	if v != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, v); err != nil {
			return resp, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return resp, nil
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