# Atriumn AI SDK

This package provides a Go client for interacting with the Atriumn AI API. It enables managing prompts and related configurations through a simple, idiomatic Go interface.

## Installation

```bash
go get github.com/atriumn/atriumn-sdk-go
```

## Usage

### Initialize the Client

```go
import (
    "context"
    "github.com/atriumn/atriumn-sdk-go/ai"
)

// Create a new client with the default configuration
client, err := ai.NewClient("https://api.atriumn.ai")
if err != nil {
    // Handle error
}

// Create a client with custom options
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

client, err := ai.NewClientWithOptions("https://api.atriumn.ai",
    ai.WithHTTPClient(httpClient),
    ai.WithUserAgent("my-custom-agent/1.0"),
)
if err != nil {
    // Handle error
}
```

### Create a Prompt

```go
ctx := context.Background()

// Define a new prompt
createRequest := &ai.CreatePromptRequest{
    Name:        "Customer Service Greeting",
    Description: "A friendly greeting for customer service interactions",
    Template:    "Hello {{customer_name}}, thank you for contacting us. How can I help you today?",
    Variables: []ai.PromptVariable{
        {
            Name:         "customer_name",
            Description:  "The name of the customer",
            DefaultValue: "valued customer",
            Required:     true,
        },
    },
    Tags: []string{"customer-service", "greeting"},
}

prompt, err := client.CreatePrompt(ctx, createRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Created prompt: %s (ID: %s)\n", prompt.Name, prompt.ID)
```

### Get a Prompt

```go
prompt, err := client.GetPrompt(ctx, "prompt-123")
if err != nil {
    // Handle error
}

fmt.Printf("Prompt: %s\nTemplate: %s\n", prompt.Name, prompt.Template)
```

### Update a Prompt

```go
// Note: We use pointers for fields that can be updated
// to distinguish between zero values and not provided fields
newName := "Updated Greeting"
newTemplate := "Hello {{customer_name}}, welcome to our support! How may I assist you today?"

updateRequest := &ai.UpdatePromptRequest{
    Name:     &newName,
    Template: &newTemplate,
    Tags:     []string{"customer-service", "greeting", "support"},
}

updatedPrompt, err := client.UpdatePrompt(ctx, "prompt-123", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Updated prompt: %s\nNew template: %s\n",
    updatedPrompt.Name, updatedPrompt.Template)
```

### Delete a Prompt

```go
err := client.DeletePrompt(ctx, "prompt-123")
if err != nil {
    // Handle error
}
fmt.Println("Prompt deleted successfully")
```

### List Prompts

```go
// List all prompts
prompts, nextToken, err := client.ListPrompts(ctx, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d prompts\n", len(prompts))
for _, p := range prompts {
    fmt.Printf("- %s (ID: %s)\n", p.Name, p.ID)
}

// List prompts with filtering and pagination
options := &ai.ListPromptsOptions{
    ModelID:    "model-abc",
    Tags:       []string{"customer-service"},
    MaxResults: 10,
}

prompts, nextToken, err := client.ListPrompts(ctx, options)
if err != nil {
    // Handle error
}

// For pagination, use the nextToken in subsequent requests
if nextToken != "" {
    nextPageOptions := &ai.ListPromptsOptions{
        ModelID:    "model-abc",
        Tags:       []string{"customer-service"},
        MaxResults: 10,
        NextToken:  nextToken,
    }

    morePrompts, nextToken, err := client.ListPrompts(ctx, nextPageOptions)
    // ...
}
```

## Error Handling

The client methods return specific errors that can be further inspected using the standard error handling mechanisms in Go:

```go
prompt, err := client.GetPrompt(ctx, "non-existent-prompt")
if err != nil {
    // Check if it's an API error
    if apiErr, ok := err.(*apierror.ErrorResponse); ok {
        switch apiErr.ErrorCode {
        case "not_found":
            fmt.Println("The specified prompt was not found")
        case "unauthorized":
            fmt.Println("Authentication failed")
        default:
            fmt.Printf("API error: %s - %s\n", apiErr.ErrorCode, apiErr.Description)
        }
    } else {
        // Handle network or other client-side errors
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Authentication

This client is designed to work with AWS IAM SigV4 request signing. You should provide an `http.Client` that is configured with the appropriate credentials:

```go
import (
    "context"
    "github.com/atriumn/atriumn-sdk-go/ai"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// Load AWS credentials
cfg, err := config.LoadDefaultConfig(context.TODO(),
    config.WithRegion("us-west-2"),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        "YOUR_ACCESS_KEY", "YOUR_SECRET_KEY", "",
    )),
)
if err != nil {
    // Handle error
}

// Create a transport that signs requests with SigV4
transport := &http.Transport{}
signingTransport := v4.NewSigningTransport(transport, cfg.Credentials, "execute-api", "us-west-2")

// Create HTTP client with the signing transport
httpClient := &http.Client{
    Transport: signingTransport,
}

// Create the Atriumn AI client with the signing HTTP client
client, err := ai.NewClientWithOptions("https://api.atriumn.ai",
    ai.WithHTTPClient(httpClient),
)
```
