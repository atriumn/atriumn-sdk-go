package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// mockTokenProvider implements the TokenProvider interface for testing
type mockTokenProvider struct {
	token string
	err   error
}

func (m *mockTokenProvider) GetToken(ctx context.Context) (string, error) {
	return m.token, m.err
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
			baseURL: "https://storage.example.com",
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
	tokenProvider := &mockTokenProvider{token: "test-token"}

	client, err := NewClientWithOptions("https://storage.example.com",
		WithHTTPClient(customHTTPClient),
		WithUserAgent(customUserAgent),
		WithTokenProvider(tokenProvider),
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

	if client.tokenProvider != tokenProvider {
		t.Errorf("client.tokenProvider not set correctly")
	}
}

func TestGenerateUploadURL_Success(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-upload-url", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req GenerateUploadURLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-file.txt", req.Filename)
		assert.Equal(t, "text/plain", req.ContentType)

		// Verify no auth header is present when no token provider
		assert.Empty(t, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"uploadUrl": "https://example-bucket.s3.amazonaws.com/test-file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
			"httpMethod": "PUT"
		}`)
	}))
	defer server.Close()

	request := &GenerateUploadURLRequest{
		Filename:    "test-file.txt",
		ContentType: "text/plain",
	}
	resp, err := client.GenerateUploadURL(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.UploadURL, "https://example-bucket.s3.amazonaws.com/test-file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256")
	assert.Equal(t, "PUT", resp.HTTPMethod)
}

func TestGenerateUploadURL_WithAuth(t *testing.T) {
	expectedToken := "test-token-12345"
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-upload-url", r.URL.Path)

		// Verify auth header
		assert.Equal(t, "Bearer "+expectedToken, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"uploadUrl": "https://example-bucket.s3.amazonaws.com/test-file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
			"httpMethod": "PUT"
		}`)
	}))
	defer server.Close()

	// Add token provider to client
	client.tokenProvider = &mockTokenProvider{token: expectedToken}

	request := &GenerateUploadURLRequest{
		Filename:    "test-file.txt",
		ContentType: "text/plain",
	}
	resp, err := client.GenerateUploadURL(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.UploadURL, "https://example-bucket.s3.amazonaws.com/test-file.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256")
	assert.Equal(t, "PUT", resp.HTTPMethod)
}

func TestGenerateUploadURL_Error(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-upload-url", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, `{
			"error": "invalid_request",
			"error_description": "The filename is required"
		}`)
	}))
	defer server.Close()

	request := &GenerateUploadURLRequest{
		ContentType: "text/plain",
		// Filename intentionally omitted
	}
	resp, err := client.GenerateUploadURL(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check that error is properly parsed
	errorResp, ok := err.(*apierror.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "invalid_request", errorResp.ErrorCode)
	assert.Equal(t, "The filename is required", errorResp.Description)
}

func TestGenerateUploadURL_TokenProviderError(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This won't be called because the request won't be sent
		t.Fatal("Server should not be called")
	}))
	defer server.Close()

	// Add token provider that returns an error
	client.tokenProvider = &mockTokenProvider{
		err: fmt.Errorf("token provider failed"),
	}

	request := &GenerateUploadURLRequest{
		Filename:    "test-file.txt",
		ContentType: "text/plain",
	}
	resp, err := client.GenerateUploadURL(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "token provider failed")
}

func TestGenerateDownloadURL_Success(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-download-url", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req GenerateDownloadURLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "tenant-123/files/document.pdf", req.S3Key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"downloadUrl": "https://example-bucket.s3.amazonaws.com/tenant-123/files/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
			"httpMethod": "GET"
		}`)
	}))
	defer server.Close()

	request := &GenerateDownloadURLRequest{
		S3Key: "tenant-123/files/document.pdf",
	}
	resp, err := client.GenerateDownloadURL(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.DownloadURL, "https://example-bucket.s3.amazonaws.com/tenant-123/files/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256")
	assert.Equal(t, "GET", resp.HTTPMethod)
}

func TestGenerateDownloadURL_WithAuth(t *testing.T) {
	expectedToken := "test-token-12345"
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-download-url", r.URL.Path)

		// Verify auth header
		assert.Equal(t, "Bearer "+expectedToken, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"downloadUrl": "https://example-bucket.s3.amazonaws.com/tenant-123/files/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
			"httpMethod": "GET"
		}`)
	}))
	defer server.Close()

	// Add token provider to client
	client.tokenProvider = &mockTokenProvider{token: expectedToken}

	request := &GenerateDownloadURLRequest{
		S3Key: "tenant-123/files/document.pdf",
	}
	resp, err := client.GenerateDownloadURL(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.DownloadURL, "https://example-bucket.s3.amazonaws.com/tenant-123/files/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256")
	assert.Equal(t, "GET", resp.HTTPMethod)
}

func TestGenerateDownloadURL_Error(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-download-url", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintln(w, `{
			"error": "not_found",
			"error_description": "The specified key does not exist"
		}`)
	}))
	defer server.Close()

	request := &GenerateDownloadURLRequest{
		S3Key: "tenant-123/files/non-existent.pdf",
	}
	resp, err := client.GenerateDownloadURL(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check that error is properly parsed
	errorResp, ok := err.(*apierror.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "not_found", errorResp.ErrorCode)
	assert.Equal(t, "The specified key does not exist", errorResp.Description)
}

func TestGenerateDownloadURLFromKey_Success(t *testing.T) {
	// Create a test server
	server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/generate-download-url", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var req GenerateDownloadURLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "tenant-123/files/document.pdf", req.S3Key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{
			"downloadUrl": "https://example-bucket.s3.amazonaws.com/tenant-123/files/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
			"httpMethod": "GET"
		}`)
	}))
	defer server.Close()

	resp, err := client.GenerateDownloadURLFromKey(context.Background(), "tenant-123/files/document.pdf")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.DownloadURL, "https://example-bucket.s3.amazonaws.com/tenant-123/files/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256")
	assert.Equal(t, "GET", resp.HTTPMethod)
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		errorCode   string
		description string
		want        string
	}{
		{
			name:        "with description",
			errorCode:   "invalid_request",
			description: "The request was invalid",
			want:        "invalid_request: The request was invalid",
		},
		{
			name:      "without description",
			errorCode: "server_error",
			want:      "server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &apierror.ErrorResponse{
				ErrorCode:   tt.errorCode,
				Description: tt.description,
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("ErrorResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNetworkError(t *testing.T) {
	client, err := NewClient("https://nonexistent.example.com")
	require.NoError(t, err)

	// Use a very short timeout to force a timeout error
	client.HTTPClient.Timeout = 1 * time.Millisecond

	// Try to call the API which should fail with a network error
	request := &GenerateUploadURLRequest{
		Filename:    "test-file.txt",
		ContentType: "text/plain",
	}
	resp, err := client.GenerateUploadURL(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, resp)

	// Check that the error is of type ErrorResponse
	errorResp, ok := err.(*apierror.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "request_timeout", errorResp.ErrorCode)
}

func TestHTTPStatusCodeErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
		expectedCode  string
	}{
		{
			name:          "Bad Request - Empty Body",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{}`,
			expectedError: "The request was invalid. Please check your input and try again.",
			expectedCode:  "bad_request",
		},
		{
			name:          "Bad Request - Malformed JSON",
			statusCode:    http.StatusBadRequest,
			responseBody:  `{malformed]`,
			expectedError: "The request was invalid. Please check your input and try again.",
			expectedCode:  "bad_request",
		},
		{
			name:          "Unauthorized - Empty Body",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `{}`,
			expectedError: "Authentication failed. Please check your credentials or login again.",
			expectedCode:  "unauthorized",
		},
		{
			name:          "Forbidden - Empty Body",
			statusCode:    http.StatusForbidden,
			responseBody:  `{}`,
			expectedError: "You don't have permission to access this resource.",
			expectedCode:  "forbidden",
		},
		{
			name:          "Not Found - Empty Body",
			statusCode:    http.StatusNotFound,
			responseBody:  `{}`,
			expectedError: "The requested resource was not found.",
			expectedCode:  "not_found",
		},
		{
			name:          "Rate Limited - Empty Body",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `{}`,
			expectedError: "Too many requests. Please try again later.",
			expectedCode:  "rate_limited",
		},
		{
			name:          "Server Error - Empty Body",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{}`,
			expectedError: "The service is currently unavailable. Please try again later.",
			expectedCode:  "server_error",
		},
		{
			name:          "Server Error - Malformed JSON",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{malformed]`,
			expectedError: "The service is currently unavailable. Please try again later.",
			expectedCode:  "server_error",
		},
		{
			name:          "Server Error - With Error Message",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `{"error":"database_error","error_description":"Failed to connect to database"}`,
			expectedError: "Failed to connect to database",
			expectedCode:  "database_error",
		},
		{
			name:          "Teapot - Unknown Status",
			statusCode:    http.StatusTeapot,
			responseBody:  `{}`,
			expectedError: "Unexpected HTTP status: 418",
			expectedCode:  "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client, _ := NewClient(server.URL)

			_, err := client.GenerateUploadURL(context.Background(), &GenerateUploadURLRequest{
				Filename:    "test.txt",
				ContentType: "text/plain",
			})

			require.Error(t, err)

			apiErr, ok := err.(*apierror.ErrorResponse)
			require.True(t, ok, "Error should be of type *apierror.ErrorResponse")

			assert.Equal(t, tt.expectedCode, apiErr.ErrorCode)

			// For Teapot case, just check if the status code is in the description
			if tt.statusCode == http.StatusTeapot {
				assert.Contains(t, apiErr.Description, "418")
			} else if apiErr.ErrorCode == "unknown_error" {
				// For unknown errors, description might include the body, so just check the prefix
				assert.Contains(t, apiErr.Description, tt.expectedError)
			} else {
				assert.Equal(t, tt.expectedError, apiErr.Description)
			}
		})
	}
}

func TestTokenProviderScenarios(t *testing.T) {
	tests := []struct {
		name          string
		tokenProvider TokenProvider
		expectedError string
	}{
		{
			name: "Token Provider Returns Error",
			tokenProvider: &mockTokenProvider{
				err: fmt.Errorf("failed to get token"),
			},
			expectedError: "failed to get token from provider: failed to get token",
		},
		{
			name: "Token Provider Returns Empty Token",
			tokenProvider: &mockTokenProvider{
				token: "",
			},
			expectedError: "", // Should not error, just send request without token
		},
		{
			name: "Token Provider Returns Valid Token",
			tokenProvider: &mockTokenProvider{
				token: "valid-token",
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.tokenProvider != nil {
					token, _ := tt.tokenProvider.GetToken(context.Background())
					if token != "" {
						assert.Equal(t, "Bearer "+token, r.Header.Get("Authorization"))
					} else {
						assert.Empty(t, r.Header.Get("Authorization"))
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, `{
					"uploadUrl": "https://example.com/upload",
					"httpMethod": "PUT"
				}`)
			}))
			defer server.Close()

			client.tokenProvider = tt.tokenProvider

			_, err := client.GenerateUploadURL(context.Background(), &GenerateUploadURLRequest{
				Filename:    "test.txt",
				ContentType: "text/plain",
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		expectedError string
	}{
		{
			name: "Upload URL - Empty Filename",
			request: &GenerateUploadURLRequest{
				Filename:    "",
				ContentType: "text/plain",
			},
			expectedError: "filename is required",
		},
		{
			name: "Upload URL - Empty Content Type",
			request: &GenerateUploadURLRequest{
				Filename:    "test.txt",
				ContentType: "",
			},
			expectedError: "content type is required",
		},
		{
			name: "Download URL - Empty S3 Key",
			request: &GenerateDownloadURLRequest{
				S3Key: "",
			},
			expectedError: "s3 key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, client := setupTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprintln(w, `{
					"error": "validation_error",
					"error_description": "`+tt.expectedError+`"
				}`)
			}))
			defer server.Close()

			var err error
			switch req := tt.request.(type) {
			case *GenerateUploadURLRequest:
				_, err = client.GenerateUploadURL(context.Background(), req)
			case *GenerateDownloadURLRequest:
				_, err = client.GenerateDownloadURL(context.Background(), req)
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestNetworkTimeoutError(t *testing.T) {
	client, err := NewClient("http://localhost:12345") // Non-existent server
	require.NoError(t, err)

	// Set a very short timeout to trigger timeout error
	client.HTTPClient.Timeout = 1 * time.Millisecond

	_, err = client.GenerateUploadURL(context.Background(), &GenerateUploadURLRequest{
		Filename:    "test.txt",
		ContentType: "text/plain",
	})

	require.Error(t, err)
	errorResp, ok := err.(*apierror.ErrorResponse)
	require.True(t, ok)
	assert.Equal(t, "request_timeout", errorResp.ErrorCode)
	assert.Contains(t, errorResp.Description, "The request timed out")
}

func TestURLConstruction(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
	}{
		{
			name:     "Simple path",
			baseURL:  "https://example.com",
			path:     "/generate-upload-url",
			expected: "https://example.com/generate-upload-url",
		},
		{
			name:     "Path without leading slash",
			baseURL:  "https://example.com",
			path:     "generate-upload-url",
			expected: "https://example.com/generate-upload-url",
		},
		{
			name:     "Base URL with path",
			baseURL:  "https://example.com/api/v1",
			path:     "/generate-upload-url",
			expected: "https://example.com/api/v1/generate-upload-url",
		},
		{
			name:     "Base URL with path and path without leading slash",
			baseURL:  "https://example.com/api/v1",
			path:     "generate-upload-url",
			expected: "https://example.com/api/v1/generate-upload-url",
		},
		{
			name:     "Base URL with trailing slash",
			baseURL:  "https://example.com/",
			path:     "generate-upload-url",
			expected: "https://example.com/generate-upload-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			req, err := client.newRequest(context.Background(), "GET", tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if req.URL.String() != tt.expected {
				t.Errorf("Expected URL %q, got %q", tt.expected, req.URL.String())
			}
		})
	}
}
