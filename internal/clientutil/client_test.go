package clientutil

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/atriumn/atriumn-sdk-go/internal/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errorTransport returns the specified error for all requests
type errorTransport struct {
	err error
}

func (t *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.err
}

func TestExecuteRequest_NetworkErrors(t *testing.T) {
	tests := []struct {
		name        string
		transport   http.RoundTripper
		wantCode    string
		wantContain string
	}{
		{
			name: "timeout error",
			transport: &errorTransport{
				err: &url.Error{
					Op:  "Get",
					URL: "https://api.example.com",
					Err: &timeoutError{},
				},
			},
			wantCode:    "request_timeout",
			wantContain: "The request timed out",
		},
		{
			name: "temporary error",
			transport: &errorTransport{
				err: &url.Error{
					Op:  "Get",
					URL: "https://api.example.com",
					Err: &temporaryError{},
				},
			},
			wantCode:    "temporary_error",
			wantContain: "temporary network error",
		},
		{
			name: "other network error",
			transport: &errorTransport{
				err: errors.New("connection refused"),
			},
			wantCode:    "network_error",
			wantContain: "Failed to connect to the service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &http.Client{
				Transport: tt.transport,
			}
			req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://api.example.com", nil)

			resp, err := ExecuteRequest(context.Background(), httpClient, req, nil)

			assert.Nil(t, resp)
			require.Error(t, err)

			apiErr, ok := err.(*apierror.ErrorResponse)
			require.True(t, ok, "Expected error to be *apierror.ErrorResponse but got %T", err)
			assert.Equal(t, tt.wantCode, apiErr.ErrorCode)
			assert.Contains(t, apiErr.Description, tt.wantContain)
		})
	}
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
			responseBody: `{"error":"validation_error","error_description":"Invalid input format"}`,
			wantCode:     "validation_error",
			wantContain:  "Invalid input format",
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
			wantContain:  "You don't have permission",
		},
		{
			name:         "not found with empty response",
			statusCode:   404,
			responseBody: `{}`,
			wantCode:     "not_found",
			wantContain:  "resource was not found",
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
			wantContain:  "service is currently unavailable",
		},
		{
			name:         "invalid error format",
			statusCode:   400,
			responseBody: `invalid json`,
			wantCode:     "bad_request",
			wantContain:  "The request was invalid",
		},
		{
			name:         "unknown error code",
			statusCode:   418, // I'm a teapot
			responseBody: `{}`,
			wantCode:     "unknown_error",
			wantContain:  "Unexpected HTTP status: 418",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			httpClient := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

			_, err := ExecuteRequest(context.Background(), httpClient, req, nil)

			require.Error(t, err)

			apiErr, ok := err.(*apierror.ErrorResponse)
			require.True(t, ok, "Expected error to be *apierror.ErrorResponse but got %T", err)
			assert.Equal(t, tt.wantCode, apiErr.ErrorCode)
			assert.Contains(t, apiErr.Description, tt.wantContain)
		})
	}
}

func TestExecuteRequest_SuccessResponse(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		resultPtr    interface{}
		validate     func(t *testing.T, result interface{})
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
		},
		{
			name:         "empty response body",
			statusCode:   204,
			responseBody: ``,
			resultPtr:    &struct{}{},
			validate: func(t *testing.T, result interface{}) {
				// Should not modify the struct but also not error
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			httpClient := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

			resp, err := ExecuteRequest(context.Background(), httpClient, req, tt.resultPtr)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.statusCode, resp.StatusCode)

			tt.validate(t, tt.resultPtr)
		})
	}
}

func TestExecuteRequest_InvalidJSONSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)

	var result struct{}
	_, err := ExecuteRequest(context.Background(), httpClient, req, &result)

	require.Error(t, err)

	apiErr, ok := err.(*apierror.ErrorResponse)
	require.True(t, ok)
	assert.Equal(t, "parse_error", apiErr.ErrorCode)
	assert.Contains(t, apiErr.Description, "Failed to parse")
}

// Mock implementations of timeout and temporary errors for testing
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout error" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

type temporaryError struct{}

func (e *temporaryError) Error() string   { return "temporary error" }
func (e *temporaryError) Timeout() bool   { return false }
func (e *temporaryError) Temporary() bool { return true }

func TestExecuteRequest_BodyReadError(t *testing.T) {
	// Create a test server that returns a valid response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer server.Close()

	// Create a custom http client that uses a transport wrapper to make the response body fail on read
	httpClient := &http.Client{
		Transport: &bodyReadErrorTransport{
			realTransport: http.DefaultTransport,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, err := ExecuteRequest(context.Background(), httpClient, req, nil)

	require.Error(t, err)
	apiErr, ok := err.(*apierror.ErrorResponse)
	require.True(t, ok)
	assert.Equal(t, "read_error", apiErr.ErrorCode)
	assert.Contains(t, apiErr.Description, "Failed to read response body")
}

// bodyReadErrorTransport wraps a transport and returns responses with a body that errors on read
type bodyReadErrorTransport struct {
	realTransport http.RoundTripper
}

func (t *bodyReadErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.realTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Replace the body with one that fails on read
	resp.Body = &errorReader{err: errors.New("read error")}
	return resp, nil
}

// errorReader always returns an error when Read is called
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return nil
}
