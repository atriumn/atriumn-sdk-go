package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/atriumn/atriumn-sdk-go/internal/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test HTTP server and client for testing
func setupTestServer(handler http.Handler) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client, _ := NewClient(server.URL)
	return server, client
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		wantErr  bool
		errCheck func(err error) bool
	}{
		{
			name:    "valid URL",
			baseURL: "https://api.example.com",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			baseURL: ":",
			wantErr: true,
			errCheck: func(err error) bool {
				return err != nil && err.Error() != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errCheck != nil && !tt.errCheck(err) {
					t.Errorf("NewClient() unexpected error format: %v", err)
				}
				return
			}
			if client == nil {
				t.Errorf("NewClient() returned nil client")
				return
			}
			if client.BaseURL == nil {
				t.Errorf("client.BaseURL is nil")
				return
			}
			if client.HTTPClient == nil {
				t.Errorf("client.HTTPClient is nil")
				return
			}
			if client.UserAgent != DefaultUserAgent {
				t.Errorf("client.UserAgent = %v, want %v", client.UserAgent, DefaultUserAgent)
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	customHTTPClient := &http.Client{Timeout: 20 * time.Second}
	customUserAgent := "custom-agent/1.0"

	client, err := NewClientWithOptions("https://api.example.com",
		WithHTTPClient(customHTTPClient),
		WithUserAgent(customUserAgent),
	)
	if err != nil {
		t.Fatalf("NewClientWithOptions() error = %v", err)
	}

	if client.HTTPClient != customHTTPClient {
		t.Errorf("client.HTTPClient = %v, want %v", client.HTTPClient, customHTTPClient)
	}

	if client.UserAgent != customUserAgent {
		t.Errorf("client.UserAgent = %v, want %v", client.UserAgent, customUserAgent)
	}
}

func TestHealth(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/health" {
			t.Errorf("Expected /health path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer server.Close()

	health, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if health.Status != "ok" {
		t.Errorf("health.Status = %v, want %v", health.Status, "ok")
	}
}

func TestGetClientCredentialsToken(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/auth/token" {
			t.Errorf("Expected /auth/token path, got %s", r.URL.Path)
		}

		// Verify request body
		var req ClientCredentialsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.GrantType != "client_credentials" {
			t.Errorf("req.GrantType = %v, want %v", req.GrantType, "client_credentials")
		}
		if req.ClientID != "test-client" {
			t.Errorf("req.ClientID = %v, want %v", req.ClientID, "test-client")
		}
		if req.ClientSecret != "test-secret" {
			t.Errorf("req.ClientSecret = %v, want %v", req.ClientSecret, "test-secret")
		}
		if req.Scope != "test-scope" {
			t.Errorf("req.Scope = %v, want %v", req.Scope, "test-scope")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"access_token": "test-token",
			"token_type": "Bearer",
			"expires_in": 3600,
			"scope": "test-scope"
		}`)
	}))
	defer server.Close()

	token, err := client.GetClientCredentialsToken(context.Background(), "test-client", "test-secret", "test-scope")
	if err != nil {
		t.Fatalf("GetClientCredentialsToken() error = %v", err)
	}
	if token.AccessToken != "test-token" {
		t.Errorf("token.AccessToken = %v, want %v", token.AccessToken, "test-token")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("token.TokenType = %v, want %v", token.TokenType, "Bearer")
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("token.ExpiresIn = %v, want %v", token.ExpiresIn, 3600)
	}
	if token.Scope != "test-scope" {
		t.Errorf("token.Scope = %v, want %v", token.Scope, "test-scope")
	}
}

func TestSignupUser(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/auth/signup" {
			t.Errorf("Expected /auth/signup path, got %s", r.URL.Path)
		}

		// Verify request body
		var req UserSignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.Email != "test@example.com" {
			t.Errorf("req.Email = %v, want %v", req.Email, "test@example.com")
		}
		if req.Password != "password123" {
			t.Errorf("req.Password = %v, want %v", req.Password, "password123")
		}
		if req.Attributes["name"] != "Test User" {
			t.Errorf("req.Attributes[\"name\"] = %v, want %v", req.Attributes["name"], "Test User")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"user_id": "user-123"}`)
	}))
	defer server.Close()

	attributes := map[string]string{"name": "Test User"}
	response, err := client.SignupUser(context.Background(), "test@example.com", "password123", attributes)
	if err != nil {
		t.Fatalf("SignupUser() error = %v", err)
	}
	if response.UserID != "user-123" {
		t.Errorf("response.UserID = %v, want %v", response.UserID, "user-123")
	}
}

func TestLoginUser(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/auth/login" {
			t.Errorf("Expected /auth/login path, got %s", r.URL.Path)
		}

		// Verify request body
		var req UserLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.Username != "test@example.com" {
			t.Errorf("req.Username = %v, want %v", req.Username, "test@example.com")
		}
		if req.Password != "password123" {
			t.Errorf("req.Password = %v, want %v", req.Password, "password123")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"access_token": "access-token",
			"id_token": "id-token",
			"refresh_token": "refresh-token",
			"token_type": "Bearer",
			"expires_in": 3600
		}`)
	}))
	defer server.Close()

	token, err := client.LoginUser(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("LoginUser() error = %v", err)
	}
	if token.AccessToken != "access-token" {
		t.Errorf("token.AccessToken = %v, want %v", token.AccessToken, "access-token")
	}
	if token.IDToken != "id-token" {
		t.Errorf("token.IDToken = %v, want %v", token.IDToken, "id-token")
	}
	if token.RefreshToken != "refresh-token" {
		t.Errorf("token.RefreshToken = %v, want %v", token.RefreshToken, "refresh-token")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("token.TokenType = %v, want %v", token.TokenType, "Bearer")
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("token.ExpiresIn = %v, want %v", token.ExpiresIn, 3600)
	}
}

func TestLogoutUser(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/auth/logout" {
			t.Errorf("Expected /auth/logout path, got %s", r.URL.Path)
		}

		// Verify request body
		var req UserLogoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.AccessToken != "access-token" {
			t.Errorf("req.AccessToken = %v, want %v", req.AccessToken, "access-token")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer server.Close()

	err := client.LogoutUser(context.Background(), "access-token")
	if err != nil {
		t.Fatalf("LogoutUser() error = %v", err)
	}
}

func TestRequestPasswordReset(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/auth/password/reset" {
			t.Errorf("Expected /auth/password/reset path, got %s", r.URL.Path)
		}

		// Verify request body
		var req PasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.Email != "test@example.com" {
			t.Errorf("req.Email = %v, want %v", req.Email, "test@example.com")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"code_delivery_details": {
				"destination": "t***@example.com",
				"delivery_medium": "EMAIL",
				"attribute_name": "email"
			}
		}`)
	}))
	defer server.Close()

	response, err := client.RequestPasswordReset(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}
	if response.CodeDeliveryDetails.Destination != "t***@example.com" {
		t.Errorf("response.CodeDeliveryDetails.Destination = %v, want %v",
			response.CodeDeliveryDetails.Destination, "t***@example.com")
	}
	if response.CodeDeliveryDetails.DeliveryMedium != "EMAIL" {
		t.Errorf("response.CodeDeliveryDetails.DeliveryMedium = %v, want %v",
			response.CodeDeliveryDetails.DeliveryMedium, "EMAIL")
	}
	if response.CodeDeliveryDetails.AttributeName != "email" {
		t.Errorf("response.CodeDeliveryDetails.AttributeName = %v, want %v",
			response.CodeDeliveryDetails.AttributeName, "email")
	}
}

func TestConfirmPasswordReset(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/auth/password/confirm" {
			t.Errorf("Expected /auth/password/confirm path, got %s", r.URL.Path)
		}

		// Verify request body
		var req ConfirmPasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.Email != "test@example.com" {
			t.Errorf("req.Email = %v, want %v", req.Email, "test@example.com")
		}
		if req.Code != "123456" {
			t.Errorf("req.Code = %v, want %v", req.Code, "123456")
		}
		if req.NewPassword != "newpassword123" {
			t.Errorf("req.NewPassword = %v, want %v", req.NewPassword, "newpassword123")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"status": "ok"}`)
	}))
	defer server.Close()

	err := client.ConfirmPasswordReset(context.Background(), "test@example.com", "123456", "newpassword123")
	if err != nil {
		t.Fatalf("ConfirmPasswordReset() error = %v", err)
	}
}

func TestErrorHandling(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintln(w, `{
			"error": "invalid_client",
			"error_description": "Client authentication failed"
		}`)
	}))
	defer server.Close()

	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	// Check that error is properly cast to ErrorResponse
	errResp, ok := err.(*apierror.ErrorResponse)
	if !ok {
		t.Fatalf("Expected error to be *apierror.ErrorResponse but got %T", err)
	}
	if errResp.ErrorCode != "invalid_client" {
		t.Errorf("errResp.ErrorCode = %v, want %v", errResp.ErrorCode, "invalid_client")
	}
	if errResp.Description != "Client authentication failed" {
		t.Errorf("errResp.Description = %v, want %v", errResp.Description, "Client authentication failed")
	}

	// Test Error() method
	expected := "invalid_client: Client authentication failed"
	if errResp.Error() != expected {
		t.Errorf("errResp.Error() = %v, want %v", errResp.Error(), expected)
	}
}

func TestInvalidURLError(t *testing.T) {
	client, _ := NewClient("https://invalid.example.com")

	// Create a server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentionally empty to force timeout
	}))
	server.Close() // Close server immediately to force connection error

	// Set invalid URL
	parsedURL, _ := url.Parse(server.URL)
	client.BaseURL = parsedURL

	// Set very short timeout to ensure fast test
	client.HTTPClient = &http.Client{Timeout: 10 * time.Millisecond}

	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
}

func TestClient_GetUserProfile(t *testing.T) {
	cases := []struct {
		name            string
		accessToken     string
		responseStatus  int
		responseBody    string
		expectedError   bool
		expectedProfile *UserProfileResponse
	}{
		{
			name:           "Success",
			accessToken:    "valid-token",
			responseStatus: http.StatusOK,
			responseBody:   `{"username":"testuser@example.com","attributes":{"email":"testuser@example.com","email_verified":"true"}}`,
			expectedError:  false,
			expectedProfile: &UserProfileResponse{
				Username: "testuser@example.com",
				Attributes: map[string]string{
					"email":          "testuser@example.com",
					"email_verified": "true",
				},
			},
		},
		{
			name:            "Unauthorized",
			accessToken:     "invalid-token",
			responseStatus:  http.StatusUnauthorized,
			responseBody:    `{"error":"invalid_token","error_description":"The access token is invalid"}`,
			expectedError:   true,
			expectedProfile: nil,
		},
		{
			name:            "Server Error",
			accessToken:     "valid-token",
			responseStatus:  http.StatusInternalServerError,
			responseBody:    `{"error":"server_error","error_description":"An error occurred on the server"}`,
			expectedError:   true,
			expectedProfile: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/auth/me", r.URL.Path)
				assert.Equal(t, fmt.Sprintf("Bearer %s", tc.accessToken), r.Header.Get("Authorization"))

				// Return response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.responseStatus)
				_, _ = w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			client, err := NewClient(server.URL)
			require.NoError(t, err)

			// Test
			profile, err := client.GetUserProfile(context.Background(), tc.accessToken)

			// Verify
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, profile)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedProfile.Username, profile.Username)
				assert.Equal(t, tc.expectedProfile.Attributes, profile.Attributes)
			}
		})
	}
}

func TestCreateClientCredential_Success(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req ClientCredentialCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Test App", req.IssuedTo)
		assert.Equal(t, []string{"read:users", "write:users"}, req.Scopes)
		assert.Equal(t, "Test credential", req.Description)
		assert.Equal(t, "tenant-123", req.TenantID)

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := `{
			"id": "cred-123",
			"client_id": "client-123",
			"client_secret": "secret-abc",
			"issued_to": "Test App",
			"scopes": ["read:users", "write:users"],
			"description": "Test credential",
			"active": true,
			"created_at": "2023-01-01T00:00:00Z",
			"tenant_id": "tenant-123"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create request
	req := ClientCredentialCreateRequest{
		IssuedTo:    "Test App",
		Scopes:      []string{"read:users", "write:users"},
		Description: "Test credential",
		TenantID:    "tenant-123",
	}

	// Call the method
	resp, err := client.CreateClientCredential(context.Background(), req)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "cred-123", resp.ID)
	assert.Equal(t, "client-123", resp.ClientID)
	assert.Equal(t, "secret-abc", resp.ClientSecret)
	assert.Equal(t, "Test App", resp.IssuedTo)
	assert.Equal(t, []string{"read:users", "write:users"}, resp.Scopes)
	assert.Equal(t, "Test credential", resp.Description)
	assert.True(t, resp.Active)
	assert.Equal(t, "2023-01-01T00:00:00Z", resp.CreatedAt)
	assert.Equal(t, "tenant-123", resp.TenantID)
}

func TestCreateClientCredential_Error(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := `{
			"error": "invalid_request",
			"error_description": "Missing required fields"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create request
	req := ClientCredentialCreateRequest{
		IssuedTo: "Test App",
		// Intentionally omit required fields
	}

	// Call the method, should get an error
	resp, err := client.CreateClientCredential(context.Background(), req)

	// Verify error response
	require.Error(t, err)
	require.Nil(t, resp)
	errorResp, ok := err.(*apierror.ErrorResponse)
	require.True(t, ok, "Expected error to be *apierror.ErrorResponse")
	assert.Equal(t, "invalid_request", errorResp.ErrorCode)
	assert.Equal(t, "Missing required fields", errorResp.Description)
}

func TestListClientCredentials_Success(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)

		// Verify query parameters
		queryParams := r.URL.Query()
		assert.Equal(t, "TestApp", queryParams.Get("issuedTo"))
		assert.Equal(t, "tenant-123", queryParams.Get("tenantId"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"credentials": [
				{
					"id": "cred-123",
					"client_id": "client-123",
					"issued_to": "TestApp",
					"scopes": ["read:users"],
					"description": "Test credential 1",
					"active": true,
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-02T00:00:00Z",
					"tenant_id": "tenant-123"
				},
				{
					"id": "cred-456",
					"client_id": "client-456",
					"issued_to": "TestApp2",
					"scopes": ["write:users"],
					"description": "Test credential 2",
					"active": false,
					"created_at": "2023-01-01T00:00:00Z",
					"tenant_id": "tenant-123"
				}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method with filters
	resp, err := client.ListClientCredentials(context.Background(), "TestApp", "tenant-123", "", false, false)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Credentials, 2)

	// Verify first credential
	assert.Equal(t, "cred-123", resp.Credentials[0].ID)
	assert.Equal(t, "client-123", resp.Credentials[0].ClientID)
	assert.Equal(t, "TestApp", resp.Credentials[0].IssuedTo)
	assert.Equal(t, []string{"read:users"}, resp.Credentials[0].Scopes)
	assert.Equal(t, "Test credential 1", resp.Credentials[0].Description)
	assert.True(t, resp.Credentials[0].Active)
	assert.Equal(t, "2023-01-01T00:00:00Z", resp.Credentials[0].CreatedAt)
	assert.Equal(t, "2023-01-02T00:00:00Z", resp.Credentials[0].UpdatedAt)
	assert.Equal(t, "tenant-123", resp.Credentials[0].TenantID)

	// Verify second credential
	assert.Equal(t, "cred-456", resp.Credentials[1].ID)
	assert.Equal(t, "client-456", resp.Credentials[1].ClientID)
	assert.Equal(t, "TestApp2", resp.Credentials[1].IssuedTo)
	assert.Equal(t, []string{"write:users"}, resp.Credentials[1].Scopes)
	assert.Equal(t, "Test credential 2", resp.Credentials[1].Description)
	assert.False(t, resp.Credentials[1].Active)
	assert.Equal(t, "2023-01-01T00:00:00Z", resp.Credentials[1].CreatedAt)
	assert.Equal(t, "", resp.Credentials[1].UpdatedAt)
	assert.Equal(t, "tenant-123", resp.Credentials[1].TenantID)
}

func TestListClientCredentials_NoFilters(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)

		// Verify no query parameters
		queryParams := r.URL.Query()
		assert.Equal(t, "", queryParams.Get("issuedTo"))
		assert.Equal(t, "", queryParams.Get("tenantId"))
		assert.Equal(t, "", queryParams.Get("scope"))
		assert.Equal(t, "", queryParams.Get("active"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{"credentials": []}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method without filters
	resp, err := client.ListClientCredentials(context.Background(), "", "", "", false, false)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Credentials)
}

func TestListClientCredentials_ScopeFilter(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)

		// Verify query parameters
		queryParams := r.URL.Query()
		assert.Equal(t, "read:users", queryParams.Get("scope"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"credentials": [
				{
					"id": "cred-123",
					"client_id": "client-123",
					"issued_to": "TestApp",
					"scopes": ["read:users"],
					"description": "Test credential with read:users scope",
					"active": true,
					"created_at": "2023-01-01T00:00:00Z",
					"tenant_id": "tenant-123"
				}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method with scope filter
	resp, err := client.ListClientCredentials(context.Background(), "", "", "read:users", false, false)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Credentials, 1)
	assert.Equal(t, "cred-123", resp.Credentials[0].ID)
	assert.Equal(t, []string{"read:users"}, resp.Credentials[0].Scopes)
}

func TestListClientCredentials_ActiveOnlyFilter(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)

		// Verify query parameters
		queryParams := r.URL.Query()
		assert.Equal(t, "true", queryParams.Get("active"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"credentials": [
				{
					"id": "cred-123",
					"client_id": "client-123",
					"issued_to": "TestApp",
					"scopes": ["read:users"],
					"description": "Active test credential",
					"active": true,
					"created_at": "2023-01-01T00:00:00Z",
					"tenant_id": "tenant-123"
				}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method with activeOnly filter
	resp, err := client.ListClientCredentials(context.Background(), "", "", "", true, false)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Credentials, 1)
	assert.Equal(t, "cred-123", resp.Credentials[0].ID)
	assert.True(t, resp.Credentials[0].Active)
}

func TestListClientCredentials_InactiveOnlyFilter(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)

		// Verify query parameters
		queryParams := r.URL.Query()
		assert.Equal(t, "false", queryParams.Get("active"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"credentials": [
				{
					"id": "cred-456",
					"client_id": "client-456",
					"issued_to": "TestApp2",
					"scopes": ["write:users"],
					"description": "Inactive test credential",
					"active": false,
					"created_at": "2023-01-01T00:00:00Z",
					"tenant_id": "tenant-123"
				}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method with inactiveOnly filter
	resp, err := client.ListClientCredentials(context.Background(), "", "", "", false, true)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Credentials, 1)
	assert.Equal(t, "cred-456", resp.Credentials[0].ID)
	assert.False(t, resp.Credentials[0].Active)
}

func TestListClientCredentials_MultipleCriteria(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials", r.URL.Path)

		// Verify query parameters
		queryParams := r.URL.Query()
		assert.Equal(t, "TestApp", queryParams.Get("issuedTo"))
		assert.Equal(t, "tenant-123", queryParams.Get("tenantId"))
		assert.Equal(t, "read:users", queryParams.Get("scope"))
		assert.Equal(t, "true", queryParams.Get("active"))

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"credentials": [
				{
					"id": "cred-123",
					"client_id": "client-123",
					"issued_to": "TestApp",
					"scopes": ["read:users"],
					"description": "Active test credential",
					"active": true,
					"created_at": "2023-01-01T00:00:00Z",
					"tenant_id": "tenant-123"
				}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method with multiple filters
	resp, err := client.ListClientCredentials(context.Background(), "TestApp", "tenant-123", "read:users", true, false)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Credentials, 1)
	assert.Equal(t, "cred-123", resp.Credentials[0].ID)
	assert.Equal(t, "TestApp", resp.Credentials[0].IssuedTo)
	assert.Equal(t, "tenant-123", resp.Credentials[0].TenantID)
	assert.Equal(t, []string{"read:users"}, resp.Credentials[0].Scopes)
	assert.True(t, resp.Credentials[0].Active)
}

func TestGetClientCredential_Success(t *testing.T) {
	credentialID := "cred-123"

	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials/"+credentialID, r.URL.Path)

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": "cred-123",
			"client_id": "client-123",
			"issued_to": "TestApp",
			"scopes": ["read:users", "write:users"],
			"description": "Test credential",
			"active": true,
			"created_at": "2023-01-01T00:00:00Z",
			"updated_at": "2023-01-02T00:00:00Z",
			"tenant_id": "tenant-123"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method
	resp, err := client.GetClientCredential(context.Background(), credentialID)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "cred-123", resp.ID)
	assert.Equal(t, "client-123", resp.ClientID)
	assert.Equal(t, "TestApp", resp.IssuedTo)
	assert.Equal(t, []string{"read:users", "write:users"}, resp.Scopes)
	assert.Equal(t, "Test credential", resp.Description)
	assert.True(t, resp.Active)
	assert.Equal(t, "2023-01-01T00:00:00Z", resp.CreatedAt)
	assert.Equal(t, "2023-01-02T00:00:00Z", resp.UpdatedAt)
	assert.Equal(t, "tenant-123", resp.TenantID)
}

func TestGetClientCredential_NotFound(t *testing.T) {
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/admin/credentials/not-exist", r.URL.Path)

		// Return error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := `{
			"error": "not_found",
			"error_description": "Credential not found"
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Call the method, should get a not found error
	resp, err := client.GetClientCredential(context.Background(), "not-exist")

	// Verify error response
	require.Error(t, err)
	require.Nil(t, resp)
	errorResp, ok := err.(*apierror.ErrorResponse)
	require.True(t, ok, "Expected error to be *apierror.ErrorResponse")
	assert.Equal(t, "not_found", errorResp.ErrorCode)
	assert.Equal(t, "Credential not found", errorResp.Description)
}
