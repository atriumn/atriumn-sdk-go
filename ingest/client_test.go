package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
				Description: "Authentication required. Please provide valid credentials.",
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
				Description: "An internal server error occurred. Please try again later.",
			},
		},
		{
			name:       "OtherError",
			statusCode: 418, // I'm a teapot
			responseBody: `{}`,
			expectedError: &apierror.ErrorResponse{
				ErrorCode:   "unknown_error",
				Description: "Unexpected status code: 418",
			},
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

			apiErr, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
			}

			if apiErr.ErrorCode != tc.expectedError.ErrorCode {
				t.Errorf("ErrorCode = %q, want %q", apiErr.ErrorCode, tc.expectedError.ErrorCode)
			}
			if apiErr.Description != tc.expectedError.Description {
				t.Errorf("Description = %q, want %q", apiErr.Description, tc.expectedError.Description)
			}
		})
	}
}

func TestClient_DoWithInvalidErrorResponses(t *testing.T) {
	testCases := []struct {
		name              string
		statusCode        int
		responseBody      string
		expectedErrorCode string
	}{
		{
			name:              "BadRequest with invalid JSON",
			statusCode:        http.StatusBadRequest,
			responseBody:      `<html>Bad Request</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "Unauthorized with invalid JSON",
			statusCode:        http.StatusUnauthorized,
			responseBody:      `<html>Unauthorized</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "Forbidden with invalid JSON",
			statusCode:        http.StatusForbidden,
			responseBody:      `<html>Forbidden</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "NotFound with invalid JSON",
			statusCode:        http.StatusNotFound,
			responseBody:      `<html>Not Found</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "TooManyRequests with invalid JSON",
			statusCode:        http.StatusTooManyRequests,
			responseBody:      `<html>Too Many Requests</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "InternalServerError with invalid JSON",
			statusCode:        http.StatusInternalServerError,
			responseBody:      `<html>Internal Server Error</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "Unknown status code with invalid JSON",
			statusCode:        http.StatusTeapot,
			responseBody:      `<html>I'm a teapot</html>`,
			expectedErrorCode: "unknown_error",
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

			apiErr, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
			}

			if apiErr.ErrorCode != tc.expectedErrorCode {
				t.Errorf("ErrorCode = %q, want %q", apiErr.ErrorCode, tc.expectedErrorCode)
			}

			// Also verify that the description contains the status code
			if !strings.Contains(apiErr.Description, fmt.Sprintf("HTTP error %d", tc.statusCode)) {
				t.Errorf("Expected description to contain status code %d, got: %s", tc.statusCode, apiErr.Description)
			}
		})
	}
}

func TestClient_DoWithNetworkErrors(t *testing.T) {
	testCases := []struct {
		name        string
		transportFn func() http.RoundTripper
		expectedErr string
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
			expectedErr: "failed to send request",
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
			expectedErr: "failed to send request",
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
			expectedErr: "failed to send request",
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

			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("Expected error containing %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create a server that sleeps to simulate a slow response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"test-id","status":"pending","tenantId":"tenant-123","timestamp":"2023-04-01T12:34:56Z"}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)

	// Create a context that will be canceled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.IngestText(ctx, &IngestTextRequest{
		Content: "test content",
	})

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected error containing 'context canceled', got %q", err.Error())
	}
}

func TestClient_DoResponseBodyReadError(t *testing.T) {
	client, _ := NewClientWithOptions(
		"https://example.com",
		WithHTTPClient(&http.Client{
			Transport: &BodyReadErrorTransport{},
		}),
	)

	_, err := client.IngestText(context.Background(), &IngestTextRequest{
		Content: "test content",
	})

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if !strings.Contains(err.Error(), "failed to read response body") {
		t.Errorf("Expected error containing 'failed to read response body', got %q", err.Error())
	}
}

func TestClient_DoJSONUnmarshalError(t *testing.T) {
	client, _ := NewClientWithOptions(
		"https://example.com",
		WithHTTPClient(&http.Client{
			Transport: &InvalidJSONTransport{},
		}),
	)

	_, err := client.IngestText(context.Background(), &IngestTextRequest{
		Content: "test content",
	})

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if !strings.Contains(err.Error(), "failed to unmarshal response") {
		t.Errorf("Expected error containing 'failed to unmarshal response', got %q", err.Error())
	}
}

func TestClient_DoWithEmptyErrorCodes(t *testing.T) {
	testCases := []struct {
		name              string
		statusCode        int
		responseBody      string
		expectedErrorCode string
		expectedDescContains string
	}{
		{
			name:              "BadRequest with empty error code",
			statusCode:        http.StatusBadRequest,
			responseBody:      `{}`,
			expectedErrorCode: "bad_request",
			expectedDescContains: "The request was invalid",
		},
		{
			name:              "Unauthorized with empty error code",
			statusCode:        http.StatusUnauthorized,
			responseBody:      `{}`,
			expectedErrorCode: "unauthorized",
			expectedDescContains: "Authentication required",
		},
		{
			name:              "Forbidden with empty error code",
			statusCode:        http.StatusForbidden,
			responseBody:      `{}`,
			expectedErrorCode: "forbidden",
			expectedDescContains: "You don't have permission",
		},
		{
			name:              "NotFound with empty error code",
			statusCode:        http.StatusNotFound,
			responseBody:      `{}`,
			expectedErrorCode: "not_found",
			expectedDescContains: "The requested resource was not found",
		},
		{
			name:              "TooManyRequests with empty error code",
			statusCode:        http.StatusTooManyRequests,
			responseBody:      `{}`,
			expectedErrorCode: "rate_limited",
			expectedDescContains: "Too many requests",
		},
		{
			name:              "InternalServerError with empty error code",
			statusCode:        http.StatusInternalServerError,
			responseBody:      `{}`,
			expectedErrorCode: "server_error",
			expectedDescContains: "An internal server error occurred",
		},
		{
			name:              "Unknown status code with empty error code",
			statusCode:        http.StatusTeapot,
			responseBody:      `{}`,
			expectedErrorCode: "unknown_error",
			expectedDescContains: "Unexpected status code: 418",
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

			apiErr, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
			}

			if apiErr.ErrorCode != tc.expectedErrorCode {
				t.Errorf("ErrorCode = %q, want %q", apiErr.ErrorCode, tc.expectedErrorCode)
			}

			if !strings.Contains(apiErr.Description, tc.expectedDescContains) {
				t.Errorf("Expected description to contain %q, got: %s", tc.expectedDescContains, apiErr.Description)
			}
		})
	}
}

func TestClient_DoWithInvalidResponseFormat(t *testing.T) {
	testCases := []struct {
		name              string
		statusCode        int
		responseBody      string
		expectedErrorCode string
	}{
		{
			name:              "BadRequest with invalid JSON",
			statusCode:        http.StatusBadRequest,
			responseBody:      `<html>Bad Request</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "Unauthorized with invalid JSON",
			statusCode:        http.StatusUnauthorized,
			responseBody:      `<html>Unauthorized</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "Forbidden with invalid JSON",
			statusCode:        http.StatusForbidden,
			responseBody:      `<html>Forbidden</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "NotFound with invalid JSON",
			statusCode:        http.StatusNotFound,
			responseBody:      `<html>Not Found</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "TooManyRequests with invalid JSON",
			statusCode:        http.StatusTooManyRequests,
			responseBody:      `<html>Too Many Requests</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "InternalServerError with invalid JSON",
			statusCode:        http.StatusInternalServerError,
			responseBody:      `<html>Internal Server Error</html>`,
			expectedErrorCode: "unknown_error",
		},
		{
			name:              "Unknown status code with invalid JSON",
			statusCode:        http.StatusTeapot,
			responseBody:      `<html>I'm a teapot</html>`,
			expectedErrorCode: "unknown_error",
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

			apiErr, ok := err.(*apierror.ErrorResponse)
			if !ok {
				t.Fatalf("Expected *apierror.ErrorResponse, got %T", err)
			}

			if apiErr.ErrorCode != tc.expectedErrorCode {
				t.Errorf("ErrorCode = %q, want %q", apiErr.ErrorCode, tc.expectedErrorCode)
			}

			// Also verify that the description contains the status code
			if !strings.Contains(apiErr.Description, fmt.Sprintf("HTTP error %d", tc.statusCode)) {
				t.Errorf("Expected description to contain status code %d, got: %s", tc.statusCode, apiErr.Description)
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