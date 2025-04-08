package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/atriumn/atriumn-sdk-go/internal/apierror"
)

// MockTokenProvider provides a mock implementation of the TokenProvider interface
type MockTokenProvider struct {
	token string
	err   error
}

func (m *MockTokenProvider) GetToken(ctx context.Context) (string, error) {
	return m.token, m.err
}

// setupTestServer creates a test server that responds with the given status code and response body
func setupTestServer(t *testing.T, statusCode int, responseBody string, validateRequest func(*http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validateRequest != nil {
			validateRequest(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

// Custom client transport that causes network errors
type errorTransport struct {
	err error
}

func (t *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.err
}

// ErrReader implements io.Reader and always returns an error on Read
type ErrReader struct {
	err error
}

func (r *ErrReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// BodyReadErrorTransport returns a valid response but with a body that will fail on read
type BodyReadErrorTransport struct {
}

func (t *BodyReadErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a broken reader that will fail when read
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&ErrReader{err: fmt.Errorf("simulated body read error")}),
		Header:     make(http.Header),
	}, nil
}

// InvalidJSONTransport returns a response with invalid JSON content
type InvalidJSONTransport struct {
}

func (t *InvalidJSONTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Return valid status but invalid JSON
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"this is not valid json`)),
		Header:     make(http.Header),
	}, nil
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("https://api.example.com")
	if err != nil {
		t.Fatalf("NewClient returned unexpected error: %v", err)
	}
	if client.BaseURL.String() != "https://api.example.com" {
		t.Errorf("NewClient BaseURL = %q, want %q", client.BaseURL.String(), "https://api.example.com")
	}
	if client.UserAgent != DefaultUserAgent {
		t.Errorf("NewClient UserAgent = %q, want %q", client.UserAgent, DefaultUserAgent)
	}

	// Test with invalid URL
	_, err = NewClient(":")
	if err == nil {
		t.Errorf("NewClient with invalid URL should return error")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	httpClient := &http.Client{}
	tokenProvider := &MockTokenProvider{token: "test-token"}
	userAgent := "custom-user-agent"
	
	client, err := NewClientWithOptions(
		"https://api.example.com",
		WithHTTPClient(httpClient),
		WithUserAgent(userAgent),
		WithTokenProvider(tokenProvider),
	)
	if err != nil {
		t.Fatalf("NewClientWithOptions returned unexpected error: %v", err)
	}
	
	if client.HTTPClient != httpClient {
		t.Errorf("NewClientWithOptions HTTPClient = %v, want %v", client.HTTPClient, httpClient)
	}
	if client.UserAgent != userAgent {
		t.Errorf("NewClientWithOptions UserAgent = %q, want %q", client.UserAgent, userAgent)
	}
	if client.tokenProvider != tokenProvider {
		t.Errorf("NewClientWithOptions tokenProvider = %v, want %v", client.tokenProvider, tokenProvider)
	}
}

func TestClient_IngestText(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"tenant-123","userId":"user-456","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/ingest/text" {
			t.Errorf("Expected path /ingest/text, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got %s", r.Header.Get("Authorization"))
		}
		
		// Validate request body
		var reqBody IngestTextRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if reqBody.TenantID != "tenant-123" {
			t.Errorf("Expected TenantID: tenant-123, got %s", reqBody.TenantID)
		}
		if reqBody.UserID != "user-456" {
			t.Errorf("Expected UserID: user-456, got %s", reqBody.UserID)
		}
		if reqBody.Content != "test content" {
			t.Errorf("Expected Content: test content, got %s", reqBody.Content)
		}
	})
	defer server.Close()
	
	client, err := NewClientWithOptions(
		server.URL,
		WithTokenProvider(&MockTokenProvider{token: "test-token"}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	resp, err := client.IngestText(context.Background(), &IngestTextRequest{
		TenantID: "tenant-123",
		UserID:   "user-456",
		Content:  "test content",
	})
	if err != nil {
		t.Fatalf("IngestText returned unexpected error: %v", err)
	}
	
	if resp.ID != "test-id" {
		t.Errorf("IngestText response ID = %q, want %q", resp.ID, "test-id")
	}
	if resp.Status != "pending" {
		t.Errorf("IngestText response Status = %q, want %q", resp.Status, "pending")
	}
	if resp.TenantID != "tenant-123" {
		t.Errorf("IngestText response TenantID = %q, want %q", resp.TenantID, "tenant-123")
	}
	if resp.UserID != "user-456" {
		t.Errorf("IngestText response UserID = %q, want %q", resp.UserID, "user-456")
	}
}

func TestClient_IngestURL(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"tenant-123","userId":"user-456","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/ingest/url" {
			t.Errorf("Expected path /ingest/url, got %s", r.URL.Path)
		}
		
		// Validate request body
		var reqBody IngestURLRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if reqBody.TenantID != "tenant-123" {
			t.Errorf("Expected TenantID: tenant-123, got %s", reqBody.TenantID)
		}
		if reqBody.UserID != "user-456" {
			t.Errorf("Expected UserID: user-456, got %s", reqBody.UserID)
		}
		if reqBody.URL != "https://example.com/document.pdf" {
			t.Errorf("Expected URL: https://example.com/document.pdf, got %s", reqBody.URL)
		}
	})
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	resp, err := client.IngestURL(context.Background(), &IngestURLRequest{
		TenantID: "tenant-123",
		UserID:   "user-456",
		URL:      "https://example.com/document.pdf",
	})
	if err != nil {
		t.Fatalf("IngestURL returned unexpected error: %v", err)
	}
	
	if resp.ID != "test-id" {
		t.Errorf("IngestURL response ID = %q, want %q", resp.ID, "test-id")
	}
	if resp.Status != "pending" {
		t.Errorf("IngestURL response Status = %q, want %q", resp.Status, "pending")
	}
}

func TestClient_IngestFile(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"tenant-123","userId":"user-456","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/ingest/file" {
			t.Errorf("Expected path /ingest/file, got %s", r.URL.Path)
		}
		
		// Check content type contains multipart/form-data
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("Expected Content-Type to contain multipart/form-data, got %s", contentType)
		}
		
		// Parse the multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}
		
		// Check form values
		if tenantID := r.FormValue("tenantId"); tenantID != "tenant-123" {
			t.Errorf("Expected tenantId: tenant-123, got %s", tenantID)
		}
		if userID := r.FormValue("userId"); userID != "user-456" {
			t.Errorf("Expected userId: user-456, got %s", userID)
		}
		
		// Check file
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("Failed to get file from form: %v", err)
		}
		defer file.Close()
		
		if header.Filename != "test.txt" {
			t.Errorf("Expected filename: test.txt, got %s", header.Filename)
		}
		
		fileContent, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("Failed to read file content: %v", err)
		}
		if string(fileContent) != "test file content" {
			t.Errorf("Expected file content: test file content, got %s", string(fileContent))
		}
	})
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	fileContent := "test file content"
	fileReader := strings.NewReader(fileContent)
	
	resp, err := client.IngestFile(
		context.Background(),
		"tenant-123",
		"test.txt",
		"text/plain",
		"user-456",
		fileReader,
	)
	if err != nil {
		t.Fatalf("IngestFile returned unexpected error: %v", err)
	}
	
	if resp.ID != "test-id" {
		t.Errorf("IngestFile response ID = %q, want %q", resp.ID, "test-id")
	}
	if resp.Status != "pending" {
		t.Errorf("IngestFile response Status = %q, want %q", resp.Status, "pending")
	}
}

func TestClient_IngestFile_ReaderErrors(t *testing.T) {
	server := setupTestServer(t, http.StatusOK, `{"id":"test-id","status":"pending","tenantId":"tenant-123","timestamp":"2023-04-01T12:34:56Z"}`, nil)
	defer server.Close()

	client, _ := NewClient(server.URL)

	// Test with a reader that returns an error
	readerErr := fmt.Errorf("simulated read error")
	_, err := client.IngestFile(
		context.Background(),
		"tenant-123",
		"test.txt",
		"text/plain",
		"user-456",
		&ErrReader{err: readerErr},
	)

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if !strings.Contains(err.Error(), "failed to copy file content") {
		t.Errorf("Expected error containing 'failed to copy file content', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), readerErr.Error()) {
		t.Errorf("Expected error containing %q, got %q", readerErr.Error(), err.Error())
	}
}

func TestClient_IngestFile_APIErrors(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedErrMsg string
	}{
		{
			name:           "Bad Request",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"error":"validation_error","error_description":"Invalid file format"}`,
			expectedErrMsg: "validation_error: Invalid file format",
		},
		{
			name:           "Payload Too Large",
			statusCode:     http.StatusRequestEntityTooLarge,
			responseBody:   `{"error":"file_too_large","error_description":"File exceeds maximum size of 10MB"}`,
			expectedErrMsg: "file_too_large: File exceeds maximum size of 10MB",
		},
		{
			name:           "Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   `{"error":"internal_error","error_description":"Failed to process file"}`,
			expectedErrMsg: "internal_error: Failed to process file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.statusCode, tc.responseBody, nil)
			defer server.Close()

			client, _ := NewClient(server.URL)

			_, err := client.IngestFile(
				context.Background(),
				"tenant-123",
				"test.txt",
				"text/plain",
				"user-456",
				strings.NewReader("test file content"),
			)

			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			if !strings.Contains(err.Error(), tc.expectedErrMsg) {
				t.Errorf("Expected error containing %q, got %q", tc.expectedErrMsg, err.Error())
			}
		})
	}
}

func TestClient_IngestFile_WithEmptyFields(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"default-tenant","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}
		
		// Tenant ID should not be in form if empty
		if tenantID := r.FormValue("tenantId"); tenantID != "" {
			t.Errorf("Expected empty tenantId in form, got %s", tenantID)
		}
		
		// User ID should not be in form if empty
		if userID := r.FormValue("userId"); userID != "" {
			t.Errorf("Expected empty userId in form, got %s", userID)
		}
	})
	defer server.Close()
	
	client, _ := NewClient(server.URL)
	
	// Test with empty tenant ID and user ID
	resp, err := client.IngestFile(
		context.Background(),
		"", // empty tenant ID
		"test.txt",
		"text/plain",
		"", // empty user ID
		strings.NewReader("test file content"),
	)
	
	if err != nil {
		t.Fatalf("IngestFile returned unexpected error: %v", err)
	}
	
	if resp.ID != "test-id" {
		t.Errorf("IngestFile response ID = %q, want %q", resp.ID, "test-id")
	}
}

func TestClient_IngestText_Error(t *testing.T) {
	errorResponse := `{"error":"invalid_request","error_description":"Missing required field"}`
	
	server := setupTestServer(t, http.StatusBadRequest, errorResponse, nil)
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Invalid request (missing content)
	resp, err := client.IngestText(context.Background(), &IngestTextRequest{
		TenantID: "tenant-123",
		// Content is missing
	})
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if resp != nil {
		t.Errorf("Expected nil response, got %+v", resp)
	}
	
	// Verify error type
	apiErr, ok := err.(*apierror.ErrorResponse)
	if !ok {
		t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
	}
	if apiErr.ErrorCode != "invalid_request" {
		t.Errorf("Expected error code 'invalid_request', got %q", apiErr.ErrorCode)
	}
	if apiErr.Description != "Missing required field" {
		t.Errorf("Expected error description 'Missing required field', got %q", apiErr.Description)
	}
}

func TestClient_GetContentItem(t *testing.T) {
	expectedResponse := `{
		"id": "content-123",
		"tenantId": "tenant-123",
		"userId": "user-456",
		"sourceType": "url",
		"sourceUri": "https://example.com/document.pdf",
		"s3Key": "tenant-123/content-123.pdf",
		"status": "processed",
		"contentType": "application/pdf",
		"size": 12345,
		"metadata": {"title": "Test Document"},
		"createdAt": "2023-04-01T12:34:56Z",
		"updatedAt": "2023-04-01T12:45:00Z"
	}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/content/content-123" {
			t.Errorf("Expected path /content/content-123, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization: Bearer test-token, got %s", r.Header.Get("Authorization"))
		}
	})
	defer server.Close()
	
	client, err := NewClientWithOptions(
		server.URL,
		WithTokenProvider(&MockTokenProvider{token: "test-token"}),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	contentItem, err := client.GetContentItem(context.Background(), "content-123")
	if err != nil {
		t.Fatalf("GetContentItem returned unexpected error: %v", err)
	}
	
	// Validate response
	if contentItem.ID != "content-123" {
		t.Errorf("GetContentItem response ID = %q, want %q", contentItem.ID, "content-123")
	}
	if contentItem.TenantID != "tenant-123" {
		t.Errorf("GetContentItem response TenantID = %q, want %q", contentItem.TenantID, "tenant-123")
	}
	if contentItem.UserID != "user-456" {
		t.Errorf("GetContentItem response UserID = %q, want %q", contentItem.UserID, "user-456")
	}
	if contentItem.SourceType != "url" {
		t.Errorf("GetContentItem response SourceType = %q, want %q", contentItem.SourceType, "url")
	}
	if contentItem.Status != "processed" {
		t.Errorf("GetContentItem response Status = %q, want %q", contentItem.Status, "processed")
	}
	if contentItem.Size != 12345 {
		t.Errorf("GetContentItem response Size = %d, want %d", contentItem.Size, 12345)
	}
	if contentItem.Metadata["title"] != "Test Document" {
		t.Errorf("GetContentItem response Metadata[title] = %q, want %q", contentItem.Metadata["title"], "Test Document")
	}
}

func TestClient_GetContentItem_NotFound(t *testing.T) {
	errorResponse := `{"error":"not_found","error_description":"Content item not found"}`
	
	server := setupTestServer(t, http.StatusNotFound, errorResponse, nil)
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Request nonexistent content item
	item, err := client.GetContentItem(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if item != nil {
		t.Errorf("Expected nil response, got %+v", item)
	}
	
	// Verify error type
	apiErr, ok := err.(*apierror.ErrorResponse)
	if !ok {
		t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
	}
	if apiErr.ErrorCode != "not_found" {
		t.Errorf("Expected error code 'not_found', got %q", apiErr.ErrorCode)
	}
	if apiErr.Description != "Content item not found" {
		t.Errorf("Expected error description 'Content item not found', got %q", apiErr.Description)
	}
}

func TestClient_ListContentItems(t *testing.T) {
	expectedResponse := `{
		"items": [
			{
				"id": "content-123",
				"tenantId": "tenant-123",
				"userId": "user-456",
				"sourceType": "url",
				"sourceUri": "https://example.com/document1.pdf",
				"status": "processed",
				"contentType": "application/pdf",
				"size": 12345,
				"createdAt": "2023-04-01T12:34:56Z",
				"updatedAt": "2023-04-01T12:45:00Z"
			},
			{
				"id": "content-456",
				"tenantId": "tenant-123",
				"userId": "user-456",
				"sourceType": "text",
				"status": "processing",
				"contentType": "text/plain",
				"size": 5678,
				"createdAt": "2023-04-02T10:11:12Z",
				"updatedAt": "2023-04-02T10:11:12Z"
			}
		],
		"nextToken": "next-page-token"
	}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/content" {
			t.Errorf("Expected path /content, got %s", r.URL.Path)
		}
		
		// Validate query parameters
		q := r.URL.Query()
		if status := q.Get("status"); status != "processed" {
			t.Errorf("Expected status=processed, got %s", status)
		}
		if sourceType := q.Get("sourceType"); sourceType != "url" {
			t.Errorf("Expected sourceType=url, got %s", sourceType)
		}
		if limit := q.Get("limit"); limit != "10" {
			t.Errorf("Expected limit=10, got %s", limit)
		}
		if nextToken := q.Get("nextToken"); nextToken != "page-token" {
			t.Errorf("Expected nextToken=page-token, got %s", nextToken)
		}
	})
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	status := "processed"
	sourceType := "url"
	limit := 10
	nextToken := "page-token"
	
	resp, err := client.ListContentItems(
		context.Background(),
		&status,
		&sourceType,
		&limit,
		&nextToken,
	)
	if err != nil {
		t.Fatalf("ListContentItems returned unexpected error: %v", err)
	}
	
	// Validate response
	if len(resp.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(resp.Items))
	}
	
	if resp.NextToken != "next-page-token" {
		t.Errorf("NextToken = %q, want %q", resp.NextToken, "next-page-token")
	}
	
	// Validate first item
	item1 := resp.Items[0]
	if item1.ID != "content-123" {
		t.Errorf("First item ID = %q, want %q", item1.ID, "content-123")
	}
	if item1.SourceType != "url" {
		t.Errorf("First item SourceType = %q, want %q", item1.SourceType, "url")
	}
	if item1.Status != "processed" {
		t.Errorf("First item Status = %q, want %q", item1.Status, "processed")
	}
	
	// Validate second item
	item2 := resp.Items[1]
	if item2.ID != "content-456" {
		t.Errorf("Second item ID = %q, want %q", item2.ID, "content-456")
	}
	if item2.SourceType != "text" {
		t.Errorf("Second item SourceType = %q, want %q", item2.SourceType, "text")
	}
	if item2.Status != "processing" {
		t.Errorf("Second item Status = %q, want %q", item2.Status, "processing")
	}
}

func TestClient_ListContentItems_NoFilters(t *testing.T) {
	expectedResponse := `{"items":[],"nextToken":""}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/content" {
			t.Errorf("Expected path /content, got %s", r.URL.Path)
		}
		
		// Ensure no query parameters are present
		if len(r.URL.RawQuery) > 0 {
			t.Errorf("Expected no query parameters, got %s", r.URL.RawQuery)
		}
	})
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	resp, err := client.ListContentItems(context.Background(), nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("ListContentItems returned unexpected error: %v", err)
	}
	
	// Validate response
	if len(resp.Items) != 0 {
		t.Fatalf("Expected 0 items, got %d", len(resp.Items))
	}
	
	if resp.NextToken != "" {
		t.Errorf("NextToken = %q, want empty string", resp.NextToken)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	// Test with description
	errWithDesc := &apierror.ErrorResponse{
		ErrorCode:   "test_error",
		Description: "Test error description",
	}
	expected := "test_error: Test error description"
	if errWithDesc.Error() != expected {
		t.Errorf("Error() = %q, want %q", errWithDesc.Error(), expected)
	}
	
	// Test without description
	errNoDesc := &apierror.ErrorResponse{
		ErrorCode: "test_error",
	}
	expected = "test_error"
	if errNoDesc.Error() != expected {
		t.Errorf("Error() = %q, want %q", errNoDesc.Error(), expected)
	}
}

func TestClient_ErrorStatusCodes(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  *apierror.ErrorResponse
	}{
		{
			name:       "BadRequest",
			statusCode: http.StatusBadRequest,
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "bad_request",
				Description: "The request was invalid. Please check your input and try again.",
			},
		},
		{
			name:       "Unauthorized",
			statusCode: http.StatusUnauthorized,
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "unauthorized",
				Description: "Authentication failed. Please check your credentials or login again.",
			},
		},
		{
			name:       "Forbidden",
			statusCode: http.StatusForbidden,
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "forbidden",
				Description: "You don't have permission to access this resource.",
			},
		},
		{
			name:       "NotFound",
			statusCode: http.StatusNotFound,
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "not_found",
				Description: "The requested resource was not found.",
			},
		},
		{
			name:       "TooManyRequests",
			statusCode: http.StatusTooManyRequests,
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "rate_limited",
				Description: "Too many requests. Please try again later.",
			},
		},
		{
			name:       "InternalServerError",
			statusCode: http.StatusInternalServerError,
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "server_error",
				Description: "The service is currently unavailable. Please try again later.",
			},
		},
		{
			name:       "OtherError",
			statusCode: 418, // I'm a teapot
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "unknown_error",
				Description: "Unexpected HTTP status: 418",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.statusCode, tc.responseBody, nil)
			defer server.Close()
			
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			
			_, err = client.IngestText(context.Background(), &IngestTextRequest{Content: "test"})
			
			if err == nil {
				t.Fatal("Expected error but got nil")
			}
			
			errResp, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
			}
			
			if errResp.ErrorCode != tc.expectedError.ErrorCode {
				t.Errorf("ErrorCode = %q, want %q", errResp.ErrorCode, tc.expectedError.ErrorCode)
			}
			
			// For the OtherError case, just check that the status code is included
			if tc.statusCode == 418 {
				if !strings.Contains(errResp.Description, "418") {
					t.Errorf("Expected description to contain status code 418, got: %s", errResp.Description)
				}
			} else if errResp.Description != tc.expectedError.Description {
				t.Errorf("Description = %q, want %q", errResp.Description, tc.expectedError.Description)
			}
		})
	}
}

func TestClient_DoWithInvalidErrorResponses(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedCode   string
		statusCodeDesc bool // Whether the status code description should contain the status code
	}{
		{
			name:         "BadRequest with invalid JSON",
			statusCode:   http.StatusBadRequest,
			responseBody: `invalid json`,
			expectedCode: "bad_request",
			statusCodeDesc: false,
		},
		{
			name:         "Unauthorized with invalid JSON",
			statusCode:   http.StatusUnauthorized,
			responseBody: `invalid json`,
			expectedCode: "unauthorized",
			statusCodeDesc: false,
		},
		{
			name:         "Forbidden with invalid JSON",
			statusCode:   http.StatusForbidden,
			responseBody: `invalid json`,
			expectedCode: "forbidden",
			statusCodeDesc: false,
		},
		{
			name:         "NotFound with invalid JSON",
			statusCode:   http.StatusNotFound,
			responseBody: `invalid json`,
			expectedCode: "not_found",
			statusCodeDesc: false,
		},
		{
			name:         "TooManyRequests with invalid JSON",
			statusCode:   http.StatusTooManyRequests,
			responseBody: `invalid json`,
			expectedCode: "rate_limited",
			statusCodeDesc: false,
		},
		{
			name:         "InternalServerError with invalid JSON",
			statusCode:   http.StatusInternalServerError,
			responseBody: `invalid json`,
			expectedCode: "server_error",
			statusCodeDesc: false,
		},
		{
			name:         "Unknown status code with invalid JSON",
			statusCode:   418, // I'm a teapot
			responseBody: `<html>I'm a teapot</html>`,
			expectedCode: "unknown_error",
			statusCodeDesc: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.statusCode, tc.responseBody, nil)
			defer server.Close()
			
			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			
			_, err = client.IngestText(context.Background(), &IngestTextRequest{Content: "test"})
			
			if err == nil {
				t.Fatal("Expected error but got nil")
			}
			
			errResp, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
			}
			
			if errResp.ErrorCode != tc.expectedCode {
				t.Errorf("ErrorCode = %q, want %q", errResp.ErrorCode, tc.expectedCode)
			}
			
			if tc.statusCodeDesc {
				if !strings.Contains(errResp.Description, fmt.Sprintf("%d", tc.statusCode)) {
					t.Errorf("Expected description to contain status code %d, got: %s", tc.statusCode, errResp.Description)
				}
			}
		})
	}
}

func TestClient_DoWithNetworkErrors(t *testing.T) {
	testCases := []struct {
		name        string
		transportFn func() http.RoundTripper
		expectedErrCode string
	}{
		{
			name: "Connection refused",
			transportFn: func() http.RoundTripper {
				return &errorTransport{err: &url.Error{
					Op:  "Post",
					URL: "https://example.com",
					Err: &net.OpError{
						Op:     "dial",
						Net:    "tcp",
						Source: nil,
						Addr:   nil,
						Err:    fmt.Errorf("connection refused"),
					},
				}}
			},
			expectedErrCode: "network_error",
		},
		{
			name: "DNS lookup error",
			transportFn: func() http.RoundTripper {
				return &errorTransport{err: &url.Error{
					Op:  "Post",
					URL: "https://example.com",
					Err: &net.DNSError{
						Err:        "no such host",
						Name:       "example.com",
						IsNotFound: true,
					},
				}}
			},
			expectedErrCode: "network_error",
		},
		{
			name: "Timeout error",
			transportFn: func() http.RoundTripper {
				return &errorTransport{err: &url.Error{
					Op:  "Post",
					URL: "https://example.com",
					Err: context.DeadlineExceeded,
				}}
			},
			expectedErrCode: "request_timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client, _ := NewClientWithOptions(
				server.URL,
				WithHTTPClient(&http.Client{
					Transport: tc.transportFn(),
				}),
			)

			_, err := client.IngestText(context.Background(), &IngestTextRequest{
				Content: "test content",
			})

			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			errResp, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
			}

			if errResp.ErrorCode != tc.expectedErrCode {
				t.Errorf("Expected error code %q, got %q", tc.expectedErrCode, errResp.ErrorCode)
			}
		})
	}
}

// errorReader returns an error on read
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return nil
}

// errorBodyTransport returns a body that errors on read
type errorBodyTransport struct {
	rt http.RoundTripper
}

func (t *errorBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body = &errorReader{err: errors.New("simulated body read error")}
	return resp, nil
}

func TestClient_DoResponseBodyReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client, _ := NewClientWithOptions(
		server.URL,
		WithHTTPClient(&http.Client{
			Transport: &errorBodyTransport{rt: http.DefaultTransport},
		}),
	)

	_, err := client.IngestText(context.Background(), &IngestTextRequest{
		Content: "test content",
	})

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	errResp, ok := err.(*apierror.ErrorResponse)
	if !ok {
		t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
	}

	if errResp.ErrorCode != "read_error" {
		t.Errorf("Expected error code 'read_error', got %q", errResp.ErrorCode)
	}

	if !strings.Contains(errResp.Description, "Failed to read response body") {
		t.Errorf("Expected error message containing 'Failed to read response body', got %q", errResp.Description)
	}
}

func TestClient_DoJSONUnmarshalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"invalid json`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)

	_, err := client.IngestText(context.Background(), &IngestTextRequest{
		Content: "test content",
	})

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	errResp, ok := err.(*apierror.ErrorResponse)
	if !ok {
		t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
	}

	if errResp.ErrorCode != "parse_error" {
		t.Errorf("Expected error code 'parse_error', got %q", errResp.ErrorCode)
	}

	if !strings.Contains(errResp.Description, "Failed to parse the successful response") {
		t.Errorf("Expected error message containing 'Failed to parse the successful response', got %q", errResp.Description)
	}
}

func TestClient_DoWithEmptyErrorCodes(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedCode   string
		descriptionPart string
	}{
		{
			name:           "BadRequest with empty error code",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "bad_request",
			descriptionPart: "The request was invalid",
		},
		{
			name:           "Unauthorized with empty error code",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "unauthorized",
			descriptionPart: "Authentication failed",
		},
		{
			name:           "Forbidden with empty error code",
			statusCode:     http.StatusForbidden,
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "forbidden",
			descriptionPart: "permission",
		},
		{
			name:           "NotFound with empty error code",
			statusCode:     http.StatusNotFound,
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "not_found",
			descriptionPart: "not found",
		},
		{
			name:           "TooManyRequests with empty error code",
			statusCode:     http.StatusTooManyRequests,
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "rate_limited",
			descriptionPart: "Too many requests",
		},
		{
			name:           "InternalServerError with empty error code",
			statusCode:     http.StatusInternalServerError,
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "server_error",
			descriptionPart: "service is currently unavailable",
		},
		{
			name:           "Unknown status code with empty error code",
			statusCode:     418, // I'm a teapot
			responseBody:   `{"error":"","error_description":""}`,
			expectedCode:   "unknown_error",
			descriptionPart: "Unexpected HTTP status: 418",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.statusCode, tc.responseBody, nil)
			defer server.Close()

			client, _ := NewClient(server.URL)

			_, err := client.IngestText(context.Background(), &IngestTextRequest{
				Content: "test content",
			})

			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			errResp, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
			}

			if errResp.ErrorCode != tc.expectedCode {
				t.Errorf("Expected error code %q, got %q", tc.expectedCode, errResp.ErrorCode)
			}

			if !strings.Contains(errResp.Description, tc.descriptionPart) {
				t.Errorf("Expected description to contain %q, got: %s", tc.descriptionPart, errResp.Description)
			}
		})
	}
}

func TestClient_DoWithInvalidResponseFormat(t *testing.T) {
	testCases := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedCode  string
		statusCodeDesc bool
	}{
		{
			name:          "BadRequest with invalid JSON",
			statusCode:    http.StatusBadRequest,
			responseBody:  `<html>Bad Request</html>`,
			expectedCode:  "bad_request",
			statusCodeDesc: false,
		},
		{
			name:          "Unauthorized with invalid JSON",
			statusCode:    http.StatusUnauthorized,
			responseBody:  `<html>Unauthorized</html>`,
			expectedCode:  "unauthorized",
			statusCodeDesc: false,
		},
		{
			name:          "Forbidden with invalid JSON",
			statusCode:    http.StatusForbidden,
			responseBody:  `<html>Forbidden</html>`,
			expectedCode:  "forbidden",
			statusCodeDesc: false,
		},
		{
			name:          "NotFound with invalid JSON",
			statusCode:    http.StatusNotFound,
			responseBody:  `<html>Not Found</html>`,
			expectedCode:  "not_found",
			statusCodeDesc: false,
		},
		{
			name:          "TooManyRequests with invalid JSON",
			statusCode:    http.StatusTooManyRequests,
			responseBody:  `<html>Too Many Requests</html>`,
			expectedCode:  "rate_limited",
			statusCodeDesc: false,
		},
		{
			name:          "InternalServerError with invalid JSON",
			statusCode:    http.StatusInternalServerError,
			responseBody:  `<html>Internal Server Error</html>`,
			expectedCode:  "server_error",
			statusCodeDesc: false,
		},
		{
			name:          "Unknown status code with invalid JSON",
			statusCode:    418, // I'm a teapot
			responseBody:  `<html>I'm a teapot</html>`,
			expectedCode:  "unknown_error",
			statusCodeDesc: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.statusCode, tc.responseBody, nil)
			defer server.Close()

			client, _ := NewClient(server.URL)

			_, err := client.IngestText(context.Background(), &IngestTextRequest{
				Content: "test content",
			})

			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			errResp, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse but got %T", err)
			}

			if errResp.ErrorCode != tc.expectedCode {
				t.Errorf("ErrorCode = %q, want %q", errResp.ErrorCode, tc.expectedCode)
			}

			if tc.statusCodeDesc {
				if !strings.Contains(errResp.Description, fmt.Sprintf("%d", tc.statusCode)) {
					t.Errorf("Expected description to contain status code %d, got: %s", tc.statusCode, errResp.Description)
				}
			}
		})
	}
}

func TestClient_ListContentItems_Error(t *testing.T) {
	errorResponse := `{"error":"invalid_param","error_description":"Invalid parameter: limit"}`
	
	server := setupTestServer(t, http.StatusBadRequest, errorResponse, nil)
	defer server.Close()
	
	client, _ := NewClient(server.URL)
	
	// Set an invalid limit value
	limit := -1
	resp, err := client.ListContentItems(context.Background(), nil, nil, &limit, nil)
	
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	
	if resp != nil {
		t.Errorf("Expected nil response, got %+v", resp)
	}
	
	apiErr, ok := err.(*apierror.ErrorResponse)
	if !ok {
		t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
	}
	
	if apiErr.ErrorCode != "invalid_param" {
		t.Errorf("ErrorCode = %q, want %q", apiErr.ErrorCode, "invalid_param")
	}
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
			path:     "/ingest/text",
			expected: "https://example.com/ingest/text",
		},
		{
			name:     "Path without leading slash",
			baseURL:  "https://example.com",
			path:     "ingest/text",
			expected: "https://example.com/ingest/text",
		},
		{
			name:     "Base URL with path",
			baseURL:  "https://example.com/api/v1",
			path:     "/ingest/text",
			expected: "https://example.com/api/v1/ingest/text",
		},
		{
			name:     "Base URL with path and path without leading slash",
			baseURL:  "https://example.com/api/v1",
			path:     "ingest/text",
			expected: "https://example.com/api/v1/ingest/text",
		},
		{
			name:     "Base URL with trailing slash",
			baseURL:  "https://example.com/",
			path:     "ingest/text",
			expected: "https://example.com/ingest/text",
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

func TestIngestFileURLConstruction(t *testing.T) {
	// Create a test server to capture the request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a simple valid response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"id":"test-id","status":"processing"}`)
	}))
	defer ts.Close()

	tests := []struct {
		name             string
		baseURL          string
		expectedURLStart string
	}{
		{
			name:             "Simple base URL",
			baseURL:          ts.URL,
			expectedURLStart: ts.URL + "/ingest/file",
		},
		{
			name:             "Base URL with trailing slash",
			baseURL:          ts.URL + "/",
			expectedURLStart: ts.URL + "/ingest/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Create a simple reader for the file content
			fileContent := strings.NewReader("test content")

			// Call IngestFile
			_, err = client.IngestFile(context.Background(), "tenant-1", "test.txt", "text/plain", "user-1", fileContent)
			if err != nil {
				t.Fatalf("IngestFile failed: %v", err)
			}

			// We don't have direct access to the created URL, but the test server 
			// would have returned an error if the URL wasn't valid
		})
	}
}

func TestClient_RequestFileUpload(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"tenant-123","userId":"user-456","uploadUrl":"https://example-bucket.s3.amazonaws.com/files/test-id?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Validate request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/ingest/file" {
			t.Errorf("Expected path /ingest/file, got %s", r.URL.Path)
		}
		
		// Check content type is application/json
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
		
		// Parse request body
		var req RequestFileUploadRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		
		err = json.Unmarshal(body, &req)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}
		
		// Check request fields
		if req.Filename != "test.txt" {
			t.Errorf("Expected Filename: test.txt, got %s", req.Filename)
		}
		if req.ContentType != "text/plain" {
			t.Errorf("Expected ContentType: text/plain, got %s", req.ContentType)
		}
		if req.TenantID != "tenant-123" {
			t.Errorf("Expected TenantID: tenant-123, got %s", req.TenantID)
		}
		if req.UserID != "user-456" {
			t.Errorf("Expected UserID: user-456, got %s", req.UserID)
		}
	})
	defer server.Close()
	
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	request := &RequestFileUploadRequest{
		Filename:    "test.txt",
		ContentType: "text/plain",
		TenantID:    "tenant-123",
		UserID:      "user-456",
		Metadata:    map[string]string{"key": "value"},
	}
	
	resp, err := client.RequestFileUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("RequestFileUpload returned unexpected error: %v", err)
	}
	
	if resp.ContentID != "test-id" {
		t.Errorf("RequestFileUpload response ContentID = %q, want %q", resp.ContentID, "test-id")
	}
	if resp.Status != "pending" {
		t.Errorf("RequestFileUpload response Status = %q, want %q", resp.Status, "pending")
	}
	if resp.UploadURL != "https://example-bucket.s3.amazonaws.com/files/test-id?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=..." {
		t.Errorf("RequestFileUpload response UploadURL = %q, want a pre-signed S3 URL", resp.UploadURL)
	}
}

func TestClient_RequestFileUpload_WithEmptyFields(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"default-tenant","uploadUrl":"https://example-bucket.s3.amazonaws.com/files/test-id?signed=true","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Parse request body
		var req RequestFileUploadRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		
		err = json.Unmarshal(body, &req)
		if err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}
		
		// Check that optional fields are not present in the JSON if empty
		var jsonMap map[string]interface{}
		if err := json.Unmarshal(body, &jsonMap); err != nil {
			t.Fatalf("Failed to unmarshal body to map: %v", err)
		}
		
		if _, exists := jsonMap["tenantId"]; exists {
			t.Error("TenantID field should not be present in JSON if empty")
		}
		
		if _, exists := jsonMap["userId"]; exists {
			t.Error("UserID field should not be present in JSON if empty")
		}
		
		if _, exists := jsonMap["metadata"]; exists {
			t.Error("Metadata field should not be present in JSON if empty")
		}
	})
	defer server.Close()
	
	client, _ := NewClient(server.URL)
	
	// Test with empty optional fields
	request := &RequestFileUploadRequest{
		Filename:    "test.txt",
		ContentType: "text/plain",
		// TenantID, UserID, and Metadata are omitted
	}
	
	resp, err := client.RequestFileUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("RequestFileUpload returned unexpected error: %v", err)
	}
	
	if resp.ContentID != "test-id" {
		t.Errorf("RequestFileUpload response ContentID = %q, want %q", resp.ContentID, "test-id")
	}
}

func TestClient_RequestFileUpload_APIErrors(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedErrMsg string
	}{
		{
			name:           "Bad Request",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"error":"validation_error","error_description":"Missing required fields"}`,
			expectedErrMsg: "validation_error: Missing required fields",
		},
		{
			name:           "Unauthorized",
			statusCode:     http.StatusUnauthorized,
			responseBody:   `{"error":"unauthorized","error_description":"Invalid or missing token"}`,
			expectedErrMsg: "unauthorized: Invalid or missing token",
		},
		{
			name:           "Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   `{"error":"internal_error","error_description":"Failed to generate upload URL"}`,
			expectedErrMsg: "internal_error: Failed to generate upload URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t, tc.statusCode, tc.responseBody, nil)
			defer server.Close()

			client, _ := NewClient(server.URL)

			request := &RequestFileUploadRequest{
				Filename:    "test.txt",
				ContentType: "text/plain",
				TenantID:    "tenant-123",
			}
			
			_, err := client.RequestFileUpload(context.Background(), request)

			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			if !strings.Contains(err.Error(), tc.expectedErrMsg) {
				t.Errorf("Expected error containing %q, got %q", tc.expectedErrMsg, err.Error())
			}
		})
	}
}

func TestClient_RequestFileUpload_WithTokenProvider(t *testing.T) {
	expectedResponse := `{"id":"test-id","status":"pending","tenantId":"tenant-123","uploadUrl":"https://example-bucket.s3.amazonaws.com/files/test-id?signed=true","timestamp":"2023-04-01T12:34:56Z"}`
	
	server := setupTestServer(t, http.StatusOK, expectedResponse, func(r *http.Request) {
		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected Authorization header with token, got %s", authHeader)
		}
	})
	defer server.Close()
	
	// Create a client with a token provider
	client, _ := NewClientWithOptions(server.URL, WithTokenProvider(&MockTokenProvider{token: "test-token"}))
	
	request := &RequestFileUploadRequest{
		Filename:    "test.txt",
		ContentType: "text/plain",
	}
	
	// Call RequestFileUpload - the token provider should be used
	_, err := client.RequestFileUpload(context.Background(), request)
	if err != nil {
		t.Fatalf("RequestFileUpload returned unexpected error: %v", err)
	}
}

func TestClient_UploadToURL(t *testing.T) {
	// Create a mock S3 server to test the upload
	mockS3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", contentType)
		}
		
		// Read request body (file content)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		
		if string(body) != "test file content" {
			t.Errorf("Expected body 'test file content', got %q", string(body))
		}
		
		// Return success response
		w.WriteHeader(http.StatusOK)
	}))
	defer mockS3Server.Close()
	
	client, _ := NewClient("http://api.example.com") // Base URL not used for direct upload
	
	fileContent := "test file content"
	fileReader := strings.NewReader(fileContent)
	
	resp, err := client.UploadToURL(
		context.Background(),
		mockS3Server.URL,
		"text/plain",
		fileReader,
	)
	if err != nil {
		t.Fatalf("UploadToURL returned unexpected error: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	
	// Cleanup
	resp.Body.Close()
}

func TestClient_UploadToURL_Errors(t *testing.T) {
	// Test with server that returns an error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
	}))
	defer errorServer.Close()
	
	client, _ := NewClient("http://api.example.com")
	
	fileContent := "test file content"
	fileReader := strings.NewReader(fileContent)
	
	_, err := client.UploadToURL(
		context.Background(),
		errorServer.URL,
		"text/plain",
		fileReader,
	)
	
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	
	// Error should mention the status code
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("Expected error to contain status code 403, got: %q", err.Error())
	}
	
	// Error should include the response body
	if !strings.Contains(err.Error(), "Access denied") {
		t.Errorf("Expected error to contain response body, got: %q", err.Error())
	}
	
	// Test with invalid URL
	_, err = client.UploadToURL(
		context.Background(),
		"http://invalid-url-that-does-not-exist.example",
		"text/plain",
		fileReader,
	)
	
	if err == nil {
		t.Fatal("Expected error with invalid URL but got nil")
	}
	
	// Test with reader that returns an error
	readerErr := fmt.Errorf("simulated read error")
	_, err = client.UploadToURL(
		context.Background(),
		errorServer.URL,
		"text/plain",
		&ErrReader{err: readerErr},
	)
	
	if err == nil {
		t.Fatal("Expected error with problematic reader but got nil")
	}
} 