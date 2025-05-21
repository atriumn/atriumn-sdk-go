package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("https://example.com")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.BaseURL.String() != "https://example.com" {
		t.Errorf("NewClient() BaseURL = %v, want %v", client.BaseURL.String(), "https://example.com")
	}
	if client.UserAgent != DefaultUserAgent {
		t.Errorf("NewClient() UserAgent = %v, want %v", client.UserAgent, DefaultUserAgent)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customHTTPClient := &http.Client{}
	customUserAgent := "custom-user-agent"

	client, err := NewClientWithOptions(
		"https://example.com",
		WithHTTPClient(customHTTPClient),
		WithUserAgent(customUserAgent),
	)
	if err != nil {
		t.Fatalf("NewClientWithOptions() error = %v", err)
	}

	if client.HTTPClient != customHTTPClient {
		t.Errorf("NewClientWithOptions() HTTPClient = %v, want %v", client.HTTPClient, customHTTPClient)
	}
	if client.UserAgent != customUserAgent {
		t.Errorf("NewClientWithOptions() UserAgent = %v, want %v", client.UserAgent, customUserAgent)
	}
}

func TestClient_CreatePrompt(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		if r.URL.Path != "/prompts" {
			t.Errorf("CreatePrompt() path = %v, want %v", r.URL.Path, "/prompts")
		}
		if r.Method != http.MethodPost {
			t.Errorf("CreatePrompt() method = %v, want %v", r.Method, http.MethodPost)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("CreatePrompt() Content-Type = %v, want %v", r.Header.Get("Content-Type"), "application/json")
		}

		// Decode the request body
		var requestBody CreatePromptRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Prepare mock response
		prompt := Prompt{
			ID:        "prompt-123",
			Name:      requestBody.Name,
			Template:  requestBody.Template,
			CreatedAt: "2023-01-01T00:00:00Z",
			UpdatedAt: "2023-01-01T00:00:00Z",
			Version:   1,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PromptResponse{Prompt: prompt})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test creating a prompt
	request := &CreatePromptRequest{
		Name:     "Test Prompt",
		Template: "This is a test prompt with {{variable}}",
	}

	prompt, err := client.CreatePrompt(context.Background(), request)
	if err != nil {
		t.Fatalf("CreatePrompt() error = %v", err)
	}

	if prompt.ID != "prompt-123" {
		t.Errorf("CreatePrompt() prompt.ID = %v, want %v", prompt.ID, "prompt-123")
	}
	if prompt.Name != request.Name {
		t.Errorf("CreatePrompt() prompt.Name = %v, want %v", prompt.Name, request.Name)
	}
	if prompt.Template != request.Template {
		t.Errorf("CreatePrompt() prompt.Template = %v, want %v", prompt.Template, request.Template)
	}
}

func TestClient_GetPrompt(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		if r.URL.Path != "/prompts/prompt-123" {
			t.Errorf("GetPrompt() path = %v, want %v", r.URL.Path, "/prompts/prompt-123")
		}
		if r.Method != http.MethodGet {
			t.Errorf("GetPrompt() method = %v, want %v", r.Method, http.MethodGet)
		}

		// Prepare mock response
		prompt := Prompt{
			ID:        "prompt-123",
			Name:      "Test Prompt",
			Template:  "This is a test prompt with {{variable}}",
			CreatedAt: "2023-01-01T00:00:00Z",
			UpdatedAt: "2023-01-01T00:00:00Z",
			Version:   1,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PromptResponse{Prompt: prompt})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test getting a prompt
	prompt, err := client.GetPrompt(context.Background(), "prompt-123")
	if err != nil {
		t.Fatalf("GetPrompt() error = %v", err)
	}

	if prompt.ID != "prompt-123" {
		t.Errorf("GetPrompt() prompt.ID = %v, want %v", prompt.ID, "prompt-123")
	}
	if prompt.Name != "Test Prompt" {
		t.Errorf("GetPrompt() prompt.Name = %v, want %v", prompt.Name, "Test Prompt")
	}
}

func TestClient_UpdatePrompt(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		if r.URL.Path != "/prompts/prompt-123" {
			t.Errorf("UpdatePrompt() path = %v, want %v", r.URL.Path, "/prompts/prompt-123")
		}
		if r.Method != http.MethodPut {
			t.Errorf("UpdatePrompt() method = %v, want %v", r.Method, http.MethodPut)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("UpdatePrompt() Content-Type = %v, want %v", r.Header.Get("Content-Type"), "application/json")
		}

		// Decode the request body
		var requestBody UpdatePromptRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Check that the update contains the expected field
		if requestBody.Name == nil || *requestBody.Name != "Updated Prompt" {
			t.Errorf("UpdatePrompt() requestBody.Name = %v, want %v", *requestBody.Name, "Updated Prompt")
		}

		// Prepare mock response
		updatedName := "Updated Prompt"
		prompt := Prompt{
			ID:        "prompt-123",
			Name:      updatedName,
			Template:  "This is a test prompt with {{variable}}",
			CreatedAt: "2023-01-01T00:00:00Z",
			UpdatedAt: "2023-01-02T00:00:00Z",
			Version:   2,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PromptResponse{Prompt: prompt})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test updating a prompt
	updatedName := "Updated Prompt"
	request := &UpdatePromptRequest{
		Name: &updatedName,
	}

	prompt, err := client.UpdatePrompt(context.Background(), "prompt-123", request)
	if err != nil {
		t.Fatalf("UpdatePrompt() error = %v", err)
	}

	if prompt.Name != updatedName {
		t.Errorf("UpdatePrompt() prompt.Name = %v, want %v", prompt.Name, updatedName)
	}
	if prompt.Version != 2 {
		t.Errorf("UpdatePrompt() prompt.Version = %v, want %v", prompt.Version, 2)
	}
}

func TestClient_DeletePrompt(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request
		if r.URL.Path != "/prompts/prompt-123" {
			t.Errorf("DeletePrompt() path = %v, want %v", r.URL.Path, "/prompts/prompt-123")
		}
		if r.Method != http.MethodDelete {
			t.Errorf("DeletePrompt() method = %v, want %v", r.Method, http.MethodDelete)
		}

		// Return success status
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test deleting a prompt
	err = client.DeletePrompt(context.Background(), "prompt-123")
	if err != nil {
		t.Fatalf("DeletePrompt() error = %v", err)
	}
}

func TestClient_ListPrompts(t *testing.T) {
	// Variables to capture the request
	var (
		capturedPath       string
		capturedModelID    string
		capturedMaxResults string
	)

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the values for validation outside the handler
		capturedPath = r.URL.Path
		capturedModelID = r.URL.Query().Get("modelId")
		capturedMaxResults = r.URL.Query().Get("maxResults")

		// Prepare mock response
		prompts := []Prompt{
			{
				ID:        "prompt-1",
				Name:      "Prompt 1",
				Template:  "Template 1",
				CreatedAt: "2023-01-01T00:00:00Z",
				UpdatedAt: "2023-01-01T00:00:00Z",
				Version:   1,
			},
			{
				ID:        "prompt-2",
				Name:      "Prompt 2",
				Template:  "Template 2",
				CreatedAt: "2023-01-02T00:00:00Z",
				UpdatedAt: "2023-01-02T00:00:00Z",
				Version:   1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PromptsResponse{
			Prompts:   prompts,
			NextToken: "next-token-123",
		})
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test listing prompts
	options := &ListPromptsOptions{
		ModelID:    "model-123",
		MaxResults: 10,
	}

	prompts, nextToken, err := client.ListPrompts(context.Background(), options)
	if err != nil {
		t.Fatalf("ListPrompts() error = %v", err)
	}

	// Validate the captured request
	if capturedPath != "/prompts" {
		t.Errorf("ListPrompts() capturedPath = %v, want %v", capturedPath, "/prompts")
	}
	if capturedModelID != "model-123" {
		t.Errorf("ListPrompts() capturedModelID = %v, want %v", capturedModelID, "model-123")
	}
	if capturedMaxResults != "10" {
		t.Errorf("ListPrompts() capturedMaxResults = %v, want %v", capturedMaxResults, "10")
	}

	// Validate response processing
	if len(prompts) != 2 {
		t.Errorf("ListPrompts() len(prompts) = %v, want %v", len(prompts), 2)
	}
	if prompts[0].ID != "prompt-1" {
		t.Errorf("ListPrompts() prompts[0].ID = %v, want %v", prompts[0].ID, "prompt-1")
	}
	if prompts[1].ID != "prompt-2" {
		t.Errorf("ListPrompts() prompts[1].ID = %v, want %v", prompts[1].ID, "prompt-2")
	}
	if nextToken != "next-token-123" {
		t.Errorf("ListPrompts() nextToken = %v, want %v", nextToken, "next-token-123")
	}
}

func TestClient_newRequest(t *testing.T) {
	client, err := NewClient("https://example.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	body := map[string]interface{}{"key": "value"}
	req, err := client.newRequest(context.Background(), http.MethodPost, "/test", body)
	if err != nil {
		t.Fatalf("newRequest() error = %v", err)
	}

	if req.Method != http.MethodPost {
		t.Errorf("newRequest() method = %v, want %v", req.Method, http.MethodPost)
	}
	if req.URL.String() != "https://example.com/test" {
		t.Errorf("newRequest() URL = %v, want %v", req.URL.String(), "https://example.com/test")
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("newRequest() Content-Type = %v, want %v", req.Header.Get("Content-Type"), "application/json")
	}
	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("newRequest() Accept = %v, want %v", req.Header.Get("Accept"), "application/json")
	}
	if req.Header.Get("User-Agent") != DefaultUserAgent {
		t.Errorf("newRequest() User-Agent = %v, want %v", req.Header.Get("User-Agent"), DefaultUserAgent)
	}
}