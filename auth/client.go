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
//
// Parameters:
//   - baseURL: The base URL for the Atriumn Auth API (required)
//
// Returns:
//   - *Client: A configured Auth client instance
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
//   - baseURL: The base URL for the Atriumn Auth API (required)
//   - options: A variadic list of ClientOption functions to customize the client
//
// Returns:
//   - *Client: A configured Auth client instance
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

// CreateClientCredential creates a new client credential with the provided parameters.
//
// Parameters:
//   - ctx: Context for the API request
//   - req: ClientCredentialCreateRequest containing credential details (required fields: IssuedTo, Scopes)
//
// Returns:
//   - *ClientCredentialCreateResponse: The created credential including the client ID and secret
//   - error: An error if the creation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - issuedToFilter: Optional filter to match the IssuedTo field
//   - tenantIDFilter: Optional filter to match the TenantID field
//   - scopeFilter: Optional filter to match credentials with a specific scope
//   - activeOnly: If true, return only active credentials
//   - inactiveOnly: If true, return only inactive credentials
//
// Returns:
//   - *ListClientCredentialsResponse: A list of matching credentials
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
//       - "server_error" if the API server experiences an error
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
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the credential to retrieve (required)
//
// Returns:
//   - *ClientCredentialResponse: The credential details (without the secret)
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the credential doesn't exist
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the credential to update (required)
//   - req: ClientCredentialUpdateRequest containing fields to update (Active, Scopes, Description)
//
// Returns:
//   - *ClientCredentialResponse: The updated credential details
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the credential doesn't exist
//       - "bad_request" if the request is invalid
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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
//
// Parameters:
//   - ctx: Context for the API request
//   - id: The unique identifier of the credential to delete (required)
//
// Returns:
//   - error: An error if the deletion fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "not_found" if the credential doesn't exist
//       - "unauthorized" if authentication fails
//       - "forbidden" if the caller lacks permissions
//       - "network_error" if the connection fails
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
//
// Parameters:
//   - ctx: Context for the API request
//
// Returns:
//   - *HealthResponse: The service health status, typically containing a "status" field
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "network_error" if the connection fails
//       - "server_error" if the API server experiences an error
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

// GetClientCredentialsToken obtains an OAuth token using the client credentials flow.
//
// Parameters:
//   - ctx: Context for the API request
//   - clientID: The client identifier (required)
//   - clientSecret: The client secret (required)
//   - scope: Optional space-delimited list of requested permission scopes
//
// Returns:
//   - *TokenResponse: The token response containing access_token, token_type, and expires_in
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the credentials are invalid
//       - "unauthorized" if authentication fails
//       - "network_error" if the connection fails
//       - "server_error" if the API server experiences an error
func (c *Client) GetClientCredentialsToken(ctx context.Context, clientID, clientSecret, scope string) (*TokenResponse, error) {
	req := ClientCredentialsRequest{
		GrantType:    "client_credentials",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/oauth/token", req)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SignupUser registers a new user with the provided email and password.
//
// Parameters:
//   - ctx: Context for the API request
//   - email: The user's email address (required)
//   - password: The user's chosen password (required)
//   - attributes: Optional map of additional user attributes
//
// Returns:
//   - *UserSignupResponse: The signup response containing the user ID
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the email or password is invalid
//       - "conflict" if the user already exists
//       - "network_error" if the connection fails
//       - "server_error" if the API server experiences an error
func (c *Client) SignupUser(ctx context.Context, email, password string, attributes map[string]string) (*UserSignupResponse, error) {
	req := UserSignupRequest{
		Email:      email,
		Password:   password,
		Attributes: attributes,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/signup", req)
	if err != nil {
		return nil, err
	}

	var resp UserSignupResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConfirmSignup confirms a user signup with a verification code.
//
// Parameters:
//   - ctx: Context for the API request
//   - username: The email address or username of the account to confirm (required)
//   - code: The verification code sent to the user during signup (required)
//
// Returns:
//   - error: An error if the confirmation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the username or code is invalid
//       - "not_found" if the user doesn't exist
//       - "expired_code" if the confirmation code has expired
//       - "network_error" if the connection fails
func (c *Client) ConfirmSignup(ctx context.Context, username, code string) error {
	req := ConfirmSignupRequest{
		Username:         username,
		ConfirmationCode: code,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/confirm-signup", req)
	if err != nil {
		return err
	}

	_, err = c.do(httpReq, nil)
	return err
}

// ResendConfirmationCode resends a confirmation code to a user.
//
// Parameters:
//   - ctx: Context for the API request
//   - username: The email address or username of the account (required)
//
// Returns:
//   - *CodeDeliveryDetails: Information about how the code was delivered
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the username is invalid
//       - "not_found" if the user doesn't exist
//       - "rate_limited" if too many codes have been requested
//       - "network_error" if the connection fails
func (c *Client) ResendConfirmationCode(ctx context.Context, username string) (*CodeDeliveryDetails, error) {
	req := ResendConfirmationCodeRequest{
		Username: username,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/resend-confirmation-code", req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		CodeDeliveryDetails *CodeDeliveryDetails `json:"codeDeliveryDetails"`
	}
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return resp.CodeDeliveryDetails, nil
}

// LoginUser authenticates a user with username/email and password.
//
// Parameters:
//   - ctx: Context for the API request
//   - username: The email address or username (required)
//   - password: The user's password (required)
//
// Returns:
//   - *TokenResponse: The token response containing access_token, id_token, refresh_token
//   - error: An error if the login fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the username or password is invalid
//       - "unauthorized" if authentication fails
//       - "not_confirmed" if the user account is not confirmed
//       - "user_disabled" if the account is disabled
//       - "network_error" if the connection fails
func (c *Client) LoginUser(ctx context.Context, username, password string) (*TokenResponse, error) {
	req := UserLoginRequest{
		Username: username,
		Password: password,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/login", req)
	if err != nil {
		return nil, err
	}

	var resp TokenResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// LogoutUser logs out a user by invalidating their access token.
//
// Parameters:
//   - ctx: Context for the API request
//   - accessToken: The JWT token to invalidate (required)
//
// Returns:
//   - error: An error if the logout fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the token is invalid
//       - "unauthorized" if the token is already invalid
//       - "network_error" if the connection fails
func (c *Client) LogoutUser(ctx context.Context, accessToken string) error {
	req := UserLogoutRequest{
		AccessToken: accessToken,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/logout", req)
	if err != nil {
		return err
	}

	_, err = c.do(httpReq, nil)
	return err
}

// RequestPasswordReset initiates a password reset for a user.
//
// Parameters:
//   - ctx: Context for the API request
//   - email: The email address of the account to reset (required)
//
// Returns:
//   - *PasswordResetResponse: Information about how the reset code was delivered
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the email is invalid
//       - "not_found" if the user doesn't exist
//       - "rate_limited" if too many resets have been requested
//       - "network_error" if the connection fails
func (c *Client) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResponse, error) {
	req := PasswordResetRequest{
		Email: email,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/forgot-password", req)
	if err != nil {
		return nil, err
	}

	var resp PasswordResetResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConfirmPasswordReset completes a password reset with a verification code.
//
// Parameters:
//   - ctx: Context for the API request
//   - email: The email address of the account being reset (required)
//   - code: The verification code sent to the user (required)
//   - newPassword: The new password to set for the account (required)
//
// Returns:
//   - error: An error if the confirmation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "bad_request" if the email, code, or password is invalid
//       - "not_found" if the user doesn't exist
//       - "expired_code" if the reset code has expired
//       - "network_error" if the connection fails
func (c *Client) ConfirmPasswordReset(ctx context.Context, email, code, newPassword string) error {
	req := ConfirmPasswordResetRequest{
		Email:       email,
		Code:        code,
		NewPassword: newPassword,
	}

	httpReq, err := c.newRequest(ctx, "POST", "/auth/confirm-forgot-password", req)
	if err != nil {
		return err
	}

	_, err = c.do(httpReq, nil)
	return err
}

// GetUserProfile retrieves the profile of an authenticated user.
//
// Parameters:
//   - ctx: Context for the API request
//   - accessToken: The JWT access token of the authenticated user (required)
//
// Returns:
//   - *UserProfileResponse: The user profile containing username and attributes
//   - error: An error if the operation fails, which can be:
//     * apierror.ErrorResponse with codes like:
//       - "unauthorized" if the token is invalid or expired
//       - "not_found" if the user doesn't exist
//       - "network_error" if the connection fails
func (c *Client) GetUserProfile(ctx context.Context, accessToken string) (*UserProfileResponse, error) {
	httpReq, err := c.newRequest(ctx, "GET", "/auth/profile", nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	var resp UserProfileResponse
	_, err = c.do(httpReq, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
