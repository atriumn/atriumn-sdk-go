// Package auth provides a Go client for interacting with the Atriumn Auth API.
package auth

import (
	"fmt"
)

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

// ClientCredentialCreateResponse represents a client credential create response with the secret
type ClientCredentialCreateResponse struct {
	ClientCredentialResponse
	ClientSecret string `json:"client_secret"`
}

// ListClientCredentialsResponse represents the response from listing client credentials
type ListClientCredentialsResponse struct {
	Credentials []ClientCredentialResponse `json:"credentials"`
}
