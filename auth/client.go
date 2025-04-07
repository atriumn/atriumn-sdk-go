// Package auth provides a Go client for interacting with the Atriumn Auth API.
// It includes functionality for managing client credentials, user authentication,
// and accessing user profiles through a simple, idiomatic Go interface.
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

	"github.com/atriumn/atriumn-sdk-go/internal/clientutil"
)

const (
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 10 * time.Second

	// DefaultUserAgent is the user agent sent in requests
	DefaultUserAgent = "atriumn-auth-client/1.0"
)

// Client is the main API client for Atriumn Auth Service.
// It handles communication with the API endpoints, including
// authentication, client credential management, and user operations.
type Client struct {
	// BaseURL is the base URL of the Atriumn Auth API
	BaseURL *url.URL

	// HTTPClient is the HTTP client used for making requests
	HTTPClient *http.Client

	// UserAgent is the user agent sent with each request
	UserAgent string
}

// NewClient creates a new Atriumn Auth API client with the specified base URL.
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

// CreateClientCredential creates a new client credential with the provided parameters.
// It returns the created credential including the client ID and secret, or an error if the operation fails.
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

// ListClientCredentials lists client credentials with optional filters.
// Parameters allow filtering by issuedTo, tenantID, scope, and active status.
// Returns a list of matching credentials or an error if the operation fails.
func (c *Client) ListClientCredentials(ctx context.Context, issuedToFilter, tenantIDFilter, scopeFilter string, activeOnly, inactiveOnly bool) (*ListClientCredentialsResponse, error) {
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
	if scopeFilter != "" {
		q.Add("scope", scopeFilter)
	}
	if activeOnly {
		q.Add("active", "true")
	} else if inactiveOnly {
		q.Add("active", "false")
	}
	httpReq.URL.RawQuery = q.Encode()

	var resp ListClientCredentialsResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetClientCredential gets a client credential by its ID.
// Returns the credential details or an error if not found or operation fails.
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

// UpdateClientCredential updates a client credential with the specified ID.
// The req parameter specifies which fields to update.
// Returns the updated credential details or an error if operation fails.
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

// DeleteClientCredential deletes a client credential with the specified ID.
// Returns an error if the deletion fails or the credential does not exist.
func (c *Client) DeleteClientCredential(ctx context.Context, id string) error {
	path := fmt.Sprintf("/admin/credentials/%s", id)
	httpReq, err := c.newRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(httpReq, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// newRequest creates an API request
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// Create the URL for the request
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

// do sends an API request and returns the API response.
// The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred.
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	return clientutil.ExecuteRequest(req.Context(), c.HTTPClient, req, v)
}

// Health checks the health status of the Auth API.
// Returns the service health status or an error if the operation fails.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}

	var resp HealthResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetClientCredentialsToken obtains an OAuth token using client credentials flow.
// Parameters include clientID, clientSecret, and optional scope.
// Returns a token response or an error if authentication fails.
func (c *Client) GetClientCredentialsToken(ctx context.Context, clientID, clientSecret, scope string) (*TokenResponse, error) {
	tokenReq := ClientCredentialsRequest{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/token", tokenReq)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SignupUser registers a new user with email, password, and optional attributes.
// Returns a signup response containing the user ID, or an error if signup fails.
func (c *Client) SignupUser(ctx context.Context, email, password string, attributes map[string]string) (*UserSignupResponse, error) {
	signupReq := UserSignupRequest{
		Email:      email,
		Password:   password,
		Attributes: attributes,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/signup", signupReq)
	if err != nil {
		return nil, err
	}

	var resp UserSignupResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConfirmSignup confirms a user registration with the verification code.
// Returns an error if confirmation fails or the code is invalid.
func (c *Client) ConfirmSignup(ctx context.Context, username, code string) error {
	confirmReq := ConfirmSignupRequest{
		Username:         username,
		ConfirmationCode: code,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/signup/confirm", confirmReq)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

// ResendConfirmationCode resends the confirmation code to the specified user.
// Returns details about the code delivery or an error if the operation fails.
func (c *Client) ResendConfirmationCode(ctx context.Context, username string) (*CodeDeliveryDetails, error) {
	resendReq := ResendConfirmationCodeRequest{
		Username: username,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/signup/resend", resendReq)
	if err != nil {
		return nil, err
	}

	var resp struct {
		CodeDeliveryDetails CodeDeliveryDetails `json:"code_delivery_details"`
	}
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.CodeDeliveryDetails, nil
}

// LoginUser authenticates a user with username and password.
// Returns a token response if authentication succeeds, or an error if it fails.
func (c *Client) LoginUser(ctx context.Context, username, password string) (*TokenResponse, error) {
	loginReq := UserLoginRequest{
		Username: username,
		Password: password,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/login", loginReq)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// LogoutUser logs out a user by invalidating their access token.
// Returns an error if the logout operation fails.
func (c *Client) LogoutUser(ctx context.Context, accessToken string) error {
	logoutReq := UserLogoutRequest{
		AccessToken: accessToken,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/logout", logoutReq)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

// RequestPasswordReset initiates a password reset process for the specified email.
// Returns password reset response with delivery details, or an error if the operation fails.
func (c *Client) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResponse, error) {
	resetReq := PasswordResetRequest{
		Email: email,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/password/reset", resetReq)
	if err != nil {
		return nil, err
	}

	var resp PasswordResetResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConfirmPasswordReset completes a password reset with the verification code and new password.
// Returns an error if the password reset fails or the code is invalid.
func (c *Client) ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error {
	confirmReq := ConfirmPasswordResetRequest{
		Email:       email,
		Code:        code,
		NewPassword: newPassword,
	}

	req, err := c.newRequest(ctx, "POST", "/auth/password/confirm", confirmReq)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	return err
}

// GetUserProfile retrieves the profile of the authenticated user.
// Requires a valid access token for authorization.
// Returns the user profile or an error if retrieval fails.
func (c *Client) GetUserProfile(ctx context.Context, accessToken string) (*UserProfileResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/auth/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	var resp UserProfileResponse
	_, err = c.do(req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
