// Package ingest provides a Go client for interacting with the Atriumn Ingest API.
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
	"time"
)

const (
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 10 * time.Second

	// DefaultUserAgent is the user agent sent in requests
	DefaultUserAgent = "atriumn-ingest-client/1.0"
)

// TokenProvider defines an interface for retrieving authentication tokens
type TokenProvider interface {
	GetToken(ctx context.Context) (string, error) // Returns the Bearer token string
}

// Client is the main API client for Atriumn Ingest Service
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

// NewClient creates a new Atriumn Ingest API client
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

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets the HTTP client for the API client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithUserAgent sets the user agent for the API client
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithTokenProvider sets the token provider for the API client
func WithTokenProvider(tp TokenProvider) ClientOption {
	return func(c *Client) {
		c.tokenProvider = tp
	}
}

// NewClientWithOptions creates a new client with custom options
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

// IngestText ingests text content through the Atriumn Ingest API
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

// IngestURL ingests content from a URL through the Atriumn Ingest API
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

// IngestFile ingests content from a file through the Atriumn Ingest API
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