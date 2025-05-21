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
	// AccessToken is the JWT token used for API authorization
	AccessToken string `json:"access_token"`
	// IDToken is an optional OpenID Connect ID token
	IDToken string `json:"id_token,omitempty"`
	// RefreshToken is an optional token used to obtain a new access token
	RefreshToken string `json:"refresh_token,omitempty"`
	// TokenType specifies the token type, typically "Bearer"
	TokenType string `json:"token_type"`
	// ExpiresIn is the validity period of the token in seconds
	ExpiresIn int64 `json:"expires_in"`
	// Scope defines the permissions granted by this token
	Scope string `json:"scope,omitempty"`
}

// HealthResponse represents the response from the health endpoint.
// It indicates the current operational status of the Auth service.
type HealthResponse struct {
	// Status indicates the health of the service (e.g., "ok", "error")
	Status string `json:"status"`
}

// ClientCredentialsRequest represents a client credentials token request.
// It is used to obtain an OAuth token using the client credentials flow.
type ClientCredentialsRequest struct {
	// GrantType must be set to "client_credentials" for this flow
	GrantType string `json:"grant_type"`
	// ClientID is the unique identifier for the client application
	ClientID string `json:"client_id"`
	// ClientSecret is the secret key for the client application
	ClientSecret string `json:"client_secret"`
	// Scope is an optional space-delimited list of requested permissions
	Scope string `json:"scope,omitempty"`
}

// UserSignupRequest represents a user signup request.
// It contains the required information to register a new user, including
// email, password, and optional attribute map.
type UserSignupRequest struct {
	// Email is the user's email address (required)
	Email string `json:"email"`
	// Password is the user's chosen password (required)
	Password string `json:"password"`
	// Attributes is an optional map of additional user attributes
	Attributes map[string]string `json:"attributes,omitempty"`
}

// UserSignupResponse represents a user signup response.
// It contains the unique identifier for the newly created user.
type UserSignupResponse struct {
	// UserID is the unique identifier for the newly created user
	UserID string `json:"user_id"`
}

// UserLoginRequest represents a user login request.
// It contains the credentials needed to authenticate a user.
type UserLoginRequest struct {
	// Username is the user's email address or username (required)
	Username string `json:"username"`
	// Password is the user's password (required)
	Password string `json:"password"`
}

// UserLogoutRequest represents a user logout request.
// It contains the access token to be invalidated during logout.
type UserLogoutRequest struct {
	// AccessToken is the JWT token to invalidate (required)
	AccessToken string `json:"access_token"`
}

// PasswordResetRequest represents a password reset request.
// It contains the email address of the user requesting a password reset.
type PasswordResetRequest struct {
	// Email is the email address of the account to reset (required)
	Email string `json:"email"`
}

// CodeDeliveryDetails contains information about code delivery.
// It describes how and where a verification or confirmation code was sent.
type CodeDeliveryDetails struct {
	// Destination is the masked version of where the code was sent (e.g., "e***@example.com")
	Destination string `json:"destination"`
	// DeliveryMedium is the method used to deliver the code (e.g., "EMAIL", "SMS")
	DeliveryMedium string `json:"delivery_medium"`
	// AttributeName is the user attribute used for delivery (e.g., "email", "phone_number")
	AttributeName string `json:"attribute_name"`
}

// PasswordResetResponse represents a password reset response.
// It contains details about how the password reset code was delivered.
type PasswordResetResponse struct {
	// CodeDeliveryDetails contains information about how the reset code was delivered
	CodeDeliveryDetails *CodeDeliveryDetails `json:"code_delivery_details,omitempty"`
}

// ConfirmPasswordResetRequest represents a confirm password reset request.
// It contains the email, verification code, and new password to complete the reset process.
type ConfirmPasswordResetRequest struct {
	// Email is the email address of the account being reset (required)
	Email string `json:"email"`
	// Code is the verification code sent to the user (required)
	Code string `json:"code"`
	// NewPassword is the new password to set for the account (required)
	NewPassword string `json:"new_password"`
}

// ConfirmSignupRequest represents a confirm signup request.
// It contains the username and confirmation code to verify a user signup.
type ConfirmSignupRequest struct {
	// Username is the email address or username of the account to confirm (required)
	Username string `json:"username"`
	// ConfirmationCode is the verification code sent to the user during signup (required)
	ConfirmationCode string `json:"confirmation_code"`
}

// ResendConfirmationCodeRequest represents a request to resend a confirmation code.
// It identifies the user who needs a new confirmation code.
type ResendConfirmationCodeRequest struct {
	// Username is the email address or username of the account (required)
	Username string `json:"username"`
}

// UserProfileResponse represents the user profile response.
// It contains the authenticated user's information and attributes.
type UserProfileResponse struct {
	// Username is the user's username or email
	Username string `json:"username"`
	// Attributes is a map of user attributes and values
	Attributes map[string]string `json:"attributes"`
}

// Admin Credentials API

// ClientCredentialCreateRequest represents a request to create a new client credential.
// It specifies who the credential is for, what scopes it has access to, and other metadata.
type ClientCredentialCreateRequest struct {
	// IssuedTo identifies who this credential is being issued to (required)
	IssuedTo string `json:"issued_to"`
	// Scopes is a list of permission scopes for this credential (required)
	Scopes []string `json:"scopes"`
	// Description is an optional human-readable description of the credential
	Description string `json:"description,omitempty"`
	// TenantID is an optional tenant identifier for multi-tenant applications
	TenantID string `json:"tenant_id,omitempty"`
}

// ClientCredentialUpdateRequest represents a request to update a client credential.
// It contains the fields that can be updated for an existing credential.
// Pointer types are used to distinguish between zero values and not provided fields.
type ClientCredentialUpdateRequest struct {
	// Active indicates whether the credential should be active
	Active *bool `json:"active,omitempty"`
	// Scopes is a list of permission scopes for this credential
	Scopes *[]string `json:"scopes,omitempty"`
	// Description is a human-readable description of the credential
	Description *string `json:"description,omitempty"`
}

// ClientCredentialResponse represents a client credential response without the secret.
// It contains all the metadata about a client credential.
type ClientCredentialResponse struct {
	// ID is the internal unique identifier for the credential
	ID string `json:"id"`
	// ClientID is the OAuth client ID (public identifier)
	ClientID string `json:"client_id"`
	// IssuedTo identifies who this credential was issued to
	IssuedTo string `json:"issued_to"`
	// Scopes is a list of permission scopes for this credential
	Scopes []string `json:"scopes"`
	// Description is a human-readable description of the credential
	Description string `json:"description"`
	// Active indicates whether the credential is currently active
	Active bool `json:"active"`
	// CreatedAt is the UTC timestamp when the credential was created
	CreatedAt string `json:"created_at"`
	// UpdatedAt is the UTC timestamp when the credential was last updated
	UpdatedAt string `json:"updated_at,omitempty"`
	// TenantID is the tenant identifier for multi-tenant applications
	TenantID string `json:"tenant_id"`
}

// ClientCredentialCreateResponse represents a client credential create response with the secret.
// This extends ClientCredentialResponse to include the client secret, which is only
// returned during credential creation and cannot be retrieved later.
type ClientCredentialCreateResponse struct {
	// Embedded ClientCredentialResponse
	ClientCredentialResponse
	// ClientSecret is the secret key for the client application (only returned once during creation)
	ClientSecret string `json:"client_secret"`
}

// ListClientCredentialsResponse represents the response from listing client credentials.
// It contains an array of client credential responses.
type ListClientCredentialsResponse struct {
	// Credentials is an array of client credentials without their secrets
	Credentials []ClientCredentialResponse `json:"credentials"`
}
