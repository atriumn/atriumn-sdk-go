// Package clientutil_test tests the client utilities.
package clientutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atriumn/atriumn-sdk-go/internal/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteRequest_Success(t *testing.T) {
	// Successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer server.Close()

	// Create HTTP client and request
	httpClient := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	require.NoError(t, err)

	// Test with nil response value
	resp, err := ExecuteRequest(context.Background(), httpClient, req, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with actual response value
	var result struct {
		Message string `json:"message"`
	}
	_, err = ExecuteRequest(context.Background(), httpClient, req, &result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result.Message)
}

func TestExecuteRequest_NetworkErrors(t *testing.T) {
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://non-existent-host-12345.local", nil)
	httpClient := &http.Client{Timeout: 5 * time.Second}

	_, err := ExecuteRequest(ctx, httpClient, req, nil)
	require.Error(t, err)
}

func TestExecuteRequest_TimeoutError(t *testing.T) {
	// Create a server that sleeps longer than the client timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer server.Close()

	// Create HTTP client with a very short timeout
	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

	// Execute the request and expect a timeout error
	_, err := ExecuteRequest(context.Background(), httpClient, req, nil)
	require.Error(t, err)
}

func TestExecuteRequest_ResponseErrors(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		wantCode     string
		wantContain  string
	}{
		{
			name:         "custom error response",
			statusCode:   400,
			responseBody: `{"error":"invalid_request","error_description":"Missing required field"}`,
			wantCode:     "invalid_request",
			wantContain:  "Missing required field",
		},
		{
			name:         "bad request with empty response",
			statusCode:   400,
			responseBody: `{}`,
			wantCode:     "bad_request",
			wantContain:  "The request was invalid",
		},
		{
			name:         "unauthorized with empty response",
			statusCode:   401,
			responseBody: `{}`,
			wantCode:     "unauthorized",
			wantContain:  "Authentication failed",
		},
		{
			name:         "forbidden with empty response",
			statusCode:   403,
			responseBody: `{}`,
			wantCode:     "forbidden",
			wantContain:  "don't have permission",
		},
		{
			name:         "not found with empty response",
			statusCode:   404,
			responseBody: `{}`,
			wantCode:     "not_found",
			wantContain:  "not found",
		},
		{
			name:         "rate limited with empty response",
			statusCode:   429,
			responseBody: `{}`,
			wantCode:     "rate_limited",
			wantContain:  "Too many requests",
		},
		{
			name:         "server error with empty response",
			statusCode:   500,
			responseBody: `{}`,
			wantCode:     "server_error",
			wantContain:  "unavailable",
		},
		{
			name:         "unknown status with empty response",
			statusCode:   418,
			responseBody: `{}`,
			wantCode:     "unknown_error",
			wantContain:  "Unexpected HTTP status: 418",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			httpClient := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

			resp, err := ExecuteRequest(context.Background(), httpClient, req, nil)
			assert.Error(t, err)
			assert.Nil(t, resp)

			// Check error code and message
			errorResp, ok := err.(*apierror.ErrorResponse)
			assert.True(t, ok, "Expected error to be *apierror.ErrorResponse")
			assert.Equal(t, tt.wantCode, errorResp.ErrorCode)
			assert.Contains(t, errorResp.Description, tt.wantContain)
		})
	}
}

func TestExecuteRequest_ResponseProcessing(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		responseBody string
		resultPtr   interface{}
		validate    func(t *testing.T, result interface{})
		expectError bool
	}{
		{
			name:         "parse json response",
			statusCode:   200,
			responseBody: `{"name":"test","value":123}`,
			resultPtr: &struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{},
			validate: func(t *testing.T, result interface{}) {
				res, ok := result.(*struct {
					Name  string `json:"name"`
					Value int    `json:"value"`
				})
				require.True(t, ok)
				assert.Equal(t, "test", res.Name)
				assert.Equal(t, 123, res.Value)
			},
			expectError: false,
		},
		{
			name:         "invalid json response",
			statusCode:   200,
			responseBody: `{invalid json`,
			resultPtr:    &struct{}{},
			validate: func(t *testing.T, result interface{}) {
				// Should not get here
				t.Fail()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			httpClient := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

			resp, err := ExecuteRequest(context.Background(), httpClient, req, tt.resultPtr)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			tt.validate(t, tt.resultPtr)
		})
	}
}

// Test for handling read errors from response body
func TestExecuteRequest_ReadBodyError(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"test"}`))
	}))
	defer server.Close()

	// Create an HTTP client with a transport that returns a body that errors on Read
	httpClient := &http.Client{
		Transport: &errorBodyTransport{},
	}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

	// Execute the request
	_, err := ExecuteRequest(context.Background(), httpClient, req, &struct{}{})
	assert.Error(t, err)
	errorResp, ok := err.(*apierror.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "read_error", errorResp.ErrorCode)
}

// Mock transport for testing read errors
type errorBodyTransport struct{}

func (t *errorBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a successful response
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       &errorReader{err: errors.New("read error")},
	}
	return resp, nil
}

// Mock reader that always errors on Read
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return nil
}