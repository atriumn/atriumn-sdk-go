// Package auth provides a Go client for interacting with the Atriumn Auth API.
package auth

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
	DefaultUserAgent = "atriumn-auth-client/1.0"
)

// Client is the main API client for Atriumn Auth Service
type Client struct {
	// BaseURL is the base URL of the Atriumn Auth API
	BaseURL *url.URL

	// HTTPClient is the HTTP client used for making requests
	HTTPClient *http.Client

	// UserAgent is the user agent sent with each request
	UserAgent string
}

// NewClient creates a new Atriumn Auth API client
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

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	ErrorCode   string `json:"error"`
	Description string `json:"error_description,omitempty"`
}

// Error satisfies the error interface
func (e *ErrorResponse) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.Description)
	}
	return e.ErrorCode
}

// Common API request/response structures

// TokenResponse represents the response from a token request
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// HealthResponse represents the response from the health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// ClientCredentialsRequest represents a client credentials token request
type ClientCredentialsRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
}

// UserSignupRequest represents a user signup request
type UserSignupRequest struct {
	Email      string            `json:"email"`
	Password   string            `json:"password"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// UserSignupResponse represents a user signup response
type UserSignupResponse struct {
	UserID string `json:"user_id"`
}

// UserLoginRequest represents a user login request
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserLogoutRequest represents a user logout request
type UserLogoutRequest struct {
	AccessToken string `json:"access_token"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// CodeDeliveryDetails contains information about code delivery
type CodeDeliveryDetails struct {
	Destination    string `json:"destination"`
	DeliveryMedium string `json:"delivery_medium"`
	AttributeName  string `json:"attribute_name"`
}

// PasswordResetResponse represents a password reset response
type PasswordResetResponse struct {
	CodeDeliveryDetails *CodeDeliveryDetails `json:"code_delivery_details,omitempty"`
}

// ConfirmPasswordResetRequest represents a confirm password reset request
type ConfirmPasswordResetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// ConfirmSignupRequest represents a confirm signup request
type ConfirmSignupRequest struct {
	Username         string `json:"username"`
	ConfirmationCode string `json:"confirmation_code"`
}

// ResendConfirmationCodeRequest represents a request to resend a confirmation code
type ResendConfirmationCodeRequest struct {
	Username string `json:"username"`
}

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	Username   string            `json:"username"`
	Attributes map[string]string `json:"attributes"`
}

// Admin Credentials API

// ClientCredentialCreateRequest represents a request to create a new client credential
type ClientCredentialCreateRequest struct {
	IssuedTo    string   `json:"issued_to"`
	Scopes      []string `json:"scopes"`
	Description string   `json:"description,omitempty"`
	TenantID    string   `json:"tenant_id,omitempty"`
}

// ClientCredentialUpdateRequest represents a request to update a client credential
type ClientCredentialUpdateRequest struct {
	Active      *bool     `json:"active,omitempty"`
	Scopes      *[]string `json:"scopes,omitempty"`
	Description *string   `json:"description,omitempty"`
}

// ClientCredentialResponse represents a client credential response without the secret
type ClientCredentialResponse struct {
	ID          string    `json:"id"`
	ClientID    string    `json:"client_id"`
	IssuedTo    string    `json:"issued_to"`
	Scopes      []string  `json:"scopes"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at,omitempty"`
	TenantID    string    `json:"tenant_id"`
}

// ClientCredentialCreateResponse represents a client credential create response with the secret
type ClientCredentialCreateResponse struct {
	ClientCredentialResponse
	ClientSecret string `json:"client_secret"`
}

// ListClientCredentialsResponse represents the response from listing client credentials
type ListClientCredentialsResponse struct {
	Credentials []ClientCredentialResponse `json:"credentials"`
}

// CreateClientCredential creates a new client credential
func (c *Client) CreateClientCredential(ctx context.Context, req ClientCredentialCreateRequest) (*ClientCredentialCreateResponse, error) {
	httpReq, err := c.newRequest(ctx, "POST", "/admin/credentials", req)
	if err != nil {
		return nil, err
	}

	var resp ClientCredentialCreateResponse
	httpResp, err := c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	return &resp, nil
}

// ListClientCredentials lists client credentials with optional filters
func (c *Client) ListClientCredentials(ctx context.Context, issuedToFilter, tenantIDFilter string) (*ListClientCredentialsResponse, error) {
	httpReq, err := c.newRequest(ctx, "GET", "/admin/credentials", nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters if they are provided
	q := httpReq.URL.Query()
	if issuedToFilter != "" {
		q.Add("issuedTo", issuedToFilter)
	}
	if tenantIDFilter != "" {
		q.Add("tenantId", tenantIDFilter)
	}
	httpReq.URL.RawQuery = q.Encode()

	var resp ListClientCredentialsResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetClientCredential gets a client credential by ID
func (c *Client) GetClientCredential(ctx context.Context, id string) (*ClientCredentialResponse, error) {
	path := fmt.Sprintf("/admin/credentials/%s", id)
	httpReq, err := c.newRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp ClientCredentialResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateClientCredential updates a client credential by ID
func (c *Client) UpdateClientCredential(ctx context.Context, id string, req ClientCredentialUpdateRequest) (*ClientCredentialResponse, error) {
	path := fmt.Sprintf("/admin/credentials/%s", id)
	httpReq, err := c.newRequest(ctx, "PATCH", path, req)
	if err != nil {
		return nil, err
	}

	var resp ClientCredentialResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteClientCredential deletes a client credential by ID
func (c *Client) DeleteClientCredential(ctx context.Context, id string) error {
	path := fmt.Sprintf("/admin/credentials/%s", id)
	httpReq, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	httpResp, err := c.do(httpReq, nil)
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	return nil
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
			Description: fmt.Sprintf("Failed to connect to the authentication service: %v", err),
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
				errResp.Description = "The authentication service is currently unavailable. Please try again later."
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

// Health checks the health of the API
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	_, err = c.do(req, &health)
	if err != nil {
		return nil, err
	}

	return &health, nil
}

// GetClientCredentialsToken obtains a token using the client credentials flow
func (c *Client) GetClientCredentialsToken(ctx context.Context, clientID, clientSecret, scope string) (*TokenResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/auth/token", ClientCredentialsRequest{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
	})
	if err != nil {
		return nil, err
	}

	var token TokenResponse
	_, err = c.do(req, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// SignupUser registers a new user
func (c *Client) SignupUser(ctx context.Context, email, password string, attributes map[string]string) (*UserSignupResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/auth/signup", UserSignupRequest{
		Email:      email,
		Password:   password,
		Attributes: attributes,
	})
	if err != nil {
		return nil, err
	}

	var response UserSignupResponse
	_, err = c.do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ConfirmSignup confirms a user registration with the verification code
func (c *Client) ConfirmSignup(ctx context.Context, username, code string) error {
	req, err := c.newRequest(ctx, "POST", "/auth/signup/confirm", ConfirmSignupRequest{
		Username:         username,
		ConfirmationCode: code,
	})
	if err != nil {
		return err
	}

	// We don't need to parse a response body, just check for success
	_, err = c.do(req, nil)
	return err
}

// ResendConfirmationCode requests a new confirmation code for a user's signup
func (c *Client) ResendConfirmationCode(ctx context.Context, username string) (*CodeDeliveryDetails, error) {
	req, err := c.newRequest(ctx, "POST", "/auth/signup/resend", ResendConfirmationCodeRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}

	var response CodeDeliveryDetails
	_, err = c.do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// LoginUser authenticates a user and returns tokens
func (c *Client) LoginUser(ctx context.Context, username, password string) (*TokenResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/auth/login", UserLoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	var token TokenResponse
	_, err = c.do(req, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// LogoutUser logs out a user from all devices
func (c *Client) LogoutUser(ctx context.Context, accessToken string) error {
	req, err := c.newRequest(ctx, "POST", "/auth/logout", UserLogoutRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

// RequestPasswordReset initiates a password reset process
func (c *Client) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResponse, error) {
	req, err := c.newRequest(ctx, "POST", "/auth/password/reset", PasswordResetRequest{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	var response PasswordResetResponse
	_, err = c.do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ConfirmPasswordReset completes the password reset process
func (c *Client) ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error {
	req, err := c.newRequest(ctx, "POST", "/auth/password/confirm", ConfirmPasswordResetRequest{
		Email:       email,
		Code:        code,
		NewPassword: newPassword,
	})
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

// GetUserProfile retrieves the profile of the currently authenticated user.
// 
// Note: The caller is responsible for adding the appropriate "Authorization: Bearer <token>" header
// to the request before calling this method. This can be done by passing a token directly 
// to this method or by configuring the HTTP client passed via WithHTTPClient with a transport 
// that injects the token automatically.
func (c *Client) GetUserProfile(ctx context.Context, accessToken string) (*UserProfileResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/auth/me", nil)
	if err != nil {
		return nil, err
	}

	// Add Bearer token authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	var profile UserProfileResponse
	_, err = c.do(req, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
} 