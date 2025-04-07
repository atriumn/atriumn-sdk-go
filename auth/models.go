// Package auth provides a Go client for interacting with the Atriumn Auth API.
// It includes functionality for managing client credentials, user authentication,
// and accessing user profiles through a simple, idiomatic Go interface.
package auth

// ErrorResponse is now provided by the internal/apierror package.

// Common API request/response structures

// TokenResponse represents the response from a token request.
// It contains the access token and related information such as token type,
// expiration time, and scope.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// HealthResponse represents the response from the health endpoint.
// It indicates the current operational status of the Auth service.
type HealthResponse struct {
	Status string `json:"status"`
}

// ClientCredentialsRequest represents a client credentials token request.
// It is used to obtain an OAuth token using the client credentials flow.
type ClientCredentialsRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope,omitempty"`
}

// UserSignupRequest represents a user signup request.
// It contains the required information to register a new user, including
// email, password, and optional attribute map.
type UserSignupRequest struct {
	Email      string            `json:"email"`
	Password   string            `json:"password"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// UserSignupResponse represents a user signup response.
// It contains the unique identifier for the newly created user.
type UserSignupResponse struct {
	UserID string `json:"user_id"`
}

// UserLoginRequest represents a user login request.
// It contains the credentials needed to authenticate a user.
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserLogoutRequest represents a user logout request.
// It contains the access token to be invalidated during logout.
type UserLogoutRequest struct {
	AccessToken string `json:"access_token"`
}

// PasswordResetRequest represents a password reset request.
// It contains the email address of the user requesting a password reset.
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// CodeDeliveryDetails contains information about code delivery.
// It describes how and where a verification or confirmation code was sent.
type CodeDeliveryDetails struct {
	Destination    string `json:"destination"`
	DeliveryMedium string `json:"delivery_medium"`
	AttributeName  string `json:"attribute_name"`
}

// PasswordResetResponse represents a password reset response.
// It contains details about how the password reset code was delivered.
type PasswordResetResponse struct {
	CodeDeliveryDetails *CodeDeliveryDetails `json:"code_delivery_details,omitempty"`
}

// ConfirmPasswordResetRequest represents a confirm password reset request.
// It contains the email, verification code, and new password to complete the reset process.
type ConfirmPasswordResetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// ConfirmSignupRequest represents a confirm signup request.
// It contains the username and confirmation code to verify a user signup.
type ConfirmSignupRequest struct {
	Username         string `json:"username"`
	ConfirmationCode string `json:"confirmation_code"`
}

// ResendConfirmationCodeRequest represents a request to resend a confirmation code.
// It identifies the user who needs a new confirmation code.
type ResendConfirmationCodeRequest struct {
	Username string `json:"username"`
}

// UserProfileResponse represents the user profile response.
// It contains the authenticated user's information and attributes.
type UserProfileResponse struct {
	Username   string            `json:"username"`
	Attributes map[string]string `json:"attributes"`
}

// Admin Credentials API

// ClientCredentialCreateRequest represents a request to create a new client credential.
// It specifies who the credential is for, what scopes it has access to, and other metadata.
type ClientCredentialCreateRequest struct {
	IssuedTo    string   `json:"issued_to"`
	Scopes      []string `json:"scopes"`
	Description string   `json:"description,omitempty"`
	TenantID    string   `json:"tenant_id,omitempty"`
}

// ClientCredentialUpdateRequest represents a request to update a client credential.
// It contains the fields that can be updated for an existing credential.
// Pointer types are used to distinguish between zero values and not provided fields.
type ClientCredentialUpdateRequest struct {
	Active      *bool     `json:"active,omitempty"`
	Scopes      *[]string `json:"scopes,omitempty"`
	Description *string   `json:"description,omitempty"`
}

// ClientCredentialResponse represents a client credential response without the secret.
// It contains all the metadata about a client credential.
type ClientCredentialResponse struct {
	ID          string   `json:"id"`
	ClientID    string   `json:"client_id"`
	IssuedTo    string   `json:"issued_to"`
	Scopes      []string `json:"scopes"`
	Description string   `json:"description"`
	Active      bool     `json:"active"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
	TenantID    string   `json:"tenant_id"`
}

// ClientCredentialCreateResponse represents a client credential create response with the secret.
// This extends ClientCredentialResponse to include the client secret, which is only
// returned during credential creation and cannot be retrieved later.
type ClientCredentialCreateResponse struct {
	ClientCredentialResponse
	ClientSecret string `json:"client_secret"`
}

// ListClientCredentialsResponse represents the response from listing client credentials.
// It contains an array of client credential responses.
type ListClientCredentialsResponse struct {
	Credentials []ClientCredentialResponse `json:"credentials"`
}
