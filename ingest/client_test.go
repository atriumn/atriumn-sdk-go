package ingest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestClient_Error(t *testing.T) {
	errorResponse := `{"error":"invalid_request","error_description":"Missing required field"}`
	
	server := setupTestServer(t, http.StatusBadRequest, errorResponse, nil)
	defer server.Close()
	
	client, _ := NewClient(server.URL)
	
	_, err := client.IngestText(context.Background(), &IngestTextRequest{
		Content: "test content",
	})
	
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	
	apiErr, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("Expected *ErrorResponse, got %T", err)
	}
	
	if apiErr.ErrorCode != "invalid_request" {
		t.Errorf("ErrorCode = %q, want %q", apiErr.ErrorCode, "invalid_request")
	}
	if apiErr.Description != "Missing required field" {
		t.Errorf("Description = %q, want %q", apiErr.Description, "Missing required field")
	}
} 