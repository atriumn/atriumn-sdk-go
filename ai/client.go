// Package ai provides a Go client for interacting with the Atriumn AI API.
// It enables managing prompts and related configurations through a simple, idiomatic Go interface.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	DefaultUserAgent = "atriumn-ai-client/1.0"
)

// Client is the main API client for Atriumn AI Service.
// It handles communication with the API endpoints for prompt management.
type Client struct {
	// BaseURL is the base URL of the Atriumn AI API
	BaseURL *url.URL

	// HTTPClient is the HTTP client used for making requests
	HTTPClient *http.Client

	// UserAgent is the user agent sent with each request
	UserAgent string
}

// NewClient creates a new Atriumn AI API client with the specified base URL.
// It returns an error if the provided URL cannot be parsed.
//
// Parameters:
//   - baseURL: The base URL for the Atriumn AI API (required)
//
// Returns:
//   - *Client: A configured AI client instance
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

// NewClientWithOptions creates a new client with custom options.
// It allows for flexible configuration of the client through functional options.
//
// Parameters:
//   - baseURL: The base URL for the Atriumn AI API (required)
//   - options: A variadic list of ClientOption functions to customize the client
//
// Returns:
//   - *Client: A configured AI client instance
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

	return req, nil
}

// do sends an API request and returns the API response
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	return clientutil.ExecuteRequest(req.Context(), c.HTTPClient, req, v)
}

// CreatePrompt creates a new prompt in the Atriumn AI system.
//
// Parameters:
//   - ctx: Context for the API request
//   - request: CreatePromptRequest containing prompt details
//
// Returns:
//   - *Prompt: The created prompt
//   - error: An error if the operation fails
func (c *Client) CreatePrompt(ctx context.Context, request *CreatePromptRequest) (*Prompt, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/prompts", request)
	if err != nil {
		return nil, err
	}

	var resp PromptResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Prompt, nil
}

// GetPrompt retrieves a prompt by its ID.
//
// Parameters:
//   - ctx: Context for the API request
//   - promptID: ID of the prompt to retrieve
//
// Returns:
//   - *Prompt: The retrieved prompt
//   - error: An error if the operation fails
func (c *Client) GetPrompt(ctx context.Context, promptID string) (*Prompt, error) {
	path := fmt.Sprintf("/prompts/%s", promptID)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp PromptResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Prompt, nil
}

// UpdatePrompt updates an existing prompt.
//
// Parameters:
//   - ctx: Context for the API request
//   - promptID: ID of the prompt to update
//   - request: UpdatePromptRequest containing the fields to update
//
// Returns:
//   - *Prompt: The updated prompt
//   - error: An error if the operation fails
func (c *Client) UpdatePrompt(ctx context.Context, promptID string, request *UpdatePromptRequest) (*Prompt, error) {
	path := fmt.Sprintf("/prompts/%s", promptID)
	req, err := c.newRequest(ctx, http.MethodPut, path, request)
	if err != nil {
		return nil, err
	}

	var resp PromptResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Prompt, nil
}

// DeletePrompt deletes a prompt by its ID.
//
// Parameters:
//   - ctx: Context for the API request
//   - promptID: ID of the prompt to delete
//
// Returns:
//   - error: An error if the operation fails
func (c *Client) DeletePrompt(ctx context.Context, promptID string) error {
	path := fmt.Sprintf("/prompts/%s", promptID)
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

// ListPrompts retrieves a list of prompts with optional filtering and pagination.
//
// Parameters:
//   - ctx: Context for the API request
//   - options: Optional ListPromptsOptions for filtering and pagination
//
// Returns:
//   - []Prompt: The list of prompts
//   - string: The next token for pagination (empty if no more pages)
//   - error: An error if the operation fails
func (c *Client) ListPrompts(ctx context.Context, options *ListPromptsOptions) ([]Prompt, string, error) {
	// Create the request with base path
	req, err := c.newRequest(ctx, http.MethodGet, "/prompts", nil)
	if err != nil {
		return nil, "", err
	}

	// Add query parameters if options are provided
	if options != nil {
		q := req.URL.Query()
		
		if options.ModelID != "" {
			q.Set("modelId", options.ModelID)
		}
		
		if len(options.Tags) > 0 {
			for _, tag := range options.Tags {
				q.Add("tags", tag)
			}
		}
		
		if options.MaxResults > 0 {
			q.Set("maxResults", strconv.Itoa(options.MaxResults))
		}
		
		if options.NextToken != "" {
			q.Set("nextToken", options.NextToken)
		}
		
		// Set the updated query parameters
		req.URL.RawQuery = q.Encode()
	}

	var resp PromptsResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, "", err
	}

	return resp.Prompts, resp.NextToken, nil
} 