// Package ai provides a Go client for interacting with the Atriumn AI API.
// It enables managing prompts and related configurations through a simple, idiomatic Go interface.
package ai

// Prompt represents a prompt configuration in the Atriumn AI system.
// It contains all the metadata and configuration needed for AI prompts.
type Prompt struct {
	// ID is the unique identifier for the prompt
	ID string `json:"id"`
	// Name is the human-readable name of the prompt
	Name string `json:"name"`
	// Description provides additional context about the prompt's purpose
	Description string `json:"description,omitempty"`
	// Template is the prompt template text with variables
	Template string `json:"template"`
	// ModelID is the ID of the AI model this prompt is associated with
	ModelID string `json:"modelId,omitempty"`
	// Parameters contains model-specific parameters for the prompt
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	// Variables defines the variables that can be used in the template
	Variables []PromptVariable `json:"variables,omitempty"`
	// Tags provides a way to categorize and filter prompts
	Tags []string `json:"tags,omitempty"`
	// Version is the current version of the prompt
	Version int64 `json:"version"`
	// CreatedAt is the UTC timestamp when the prompt was created
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the UTC timestamp when the prompt was last updated
	UpdatedAt string `json:"updatedAt"`
}

// PromptVariable defines a variable that can be used in a prompt template.
type PromptVariable struct {
	// Name is the name of the variable as it appears in the prompt template
	Name string `json:"name"`
	// Description provides context about what the variable represents
	Description string `json:"description,omitempty"`
	// DefaultValue is an optional default value for the variable
	DefaultValue string `json:"defaultValue,omitempty"`
	// Required indicates if the variable must be provided when using the prompt
	Required bool `json:"required,omitempty"`
}

// CreatePromptRequest represents the request payload for creating a new prompt.
type CreatePromptRequest struct {
	// Name is the human-readable name of the prompt (required)
	Name string `json:"name"`
	// Description provides additional context about the prompt's purpose
	Description string `json:"description,omitempty"`
	// Template is the prompt template text with variables (required)
	Template string `json:"template"`
	// ModelID is the ID of the AI model this prompt is associated with
	ModelID string `json:"modelId,omitempty"`
	// Parameters contains model-specific parameters for the prompt
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	// Variables defines the variables that can be used in the template
	Variables []PromptVariable `json:"variables,omitempty"`
	// Tags provides a way to categorize and filter prompts
	Tags []string `json:"tags,omitempty"`
}

// UpdatePromptRequest represents the request payload for updating an existing prompt.
// Pointer types are used to distinguish between zero values and not provided fields.
type UpdatePromptRequest struct {
	// Name is the human-readable name of the prompt
	Name *string `json:"name,omitempty"`
	// Description provides additional context about the prompt's purpose
	Description *string `json:"description,omitempty"`
	// Template is the prompt template text with variables
	Template *string `json:"template,omitempty"`
	// ModelID is the ID of the AI model this prompt is associated with
	ModelID *string `json:"modelId,omitempty"`
	// Parameters contains model-specific parameters for the prompt
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	// Variables defines the variables that can be used in the template
	Variables []PromptVariable `json:"variables,omitempty"`
	// Tags provides a way to categorize and filter prompts
	Tags []string `json:"tags,omitempty"`
}

// PromptResponse represents the response body from the API containing a single prompt.
type PromptResponse struct {
	// Prompt is the retrieved prompt configuration
	Prompt Prompt `json:"prompt"`
}

// PromptsResponse represents the response body from the API containing multiple prompts.
type PromptsResponse struct {
	// Prompts is an array of prompt configurations
	Prompts []Prompt `json:"prompts"`
	// NextToken is an optional pagination token for retrieving the next set of results
	NextToken string `json:"nextToken,omitempty"`
}

// ListPromptsOptions represents optional parameters for listing prompts.
type ListPromptsOptions struct {
	// ModelID optionally filters prompts by their associated model
	ModelID string `json:"modelId,omitempty"`
	// Tags optionally filters prompts by their tags
	Tags []string `json:"tags,omitempty"`
	// MaxResults is the maximum number of results to return per page
	MaxResults int `json:"maxResults,omitempty"`
	// NextToken is the pagination token for retrieving the next set of results
	NextToken string `json:"nextToken,omitempty"`
}
