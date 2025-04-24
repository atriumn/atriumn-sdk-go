// This example demonstrates how to use the Atriumn AI client
// to perform CRUD operations on prompts.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atriumn/atriumn-sdk-go/ai"
)

func main() {
	// Replace with your Atriumn AI API endpoint
	apiEndpoint := "https://api.atriumn.ai"

	// Create a client with default options
	client, err := ai.NewClient(apiEndpoint)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// For a client with custom options:
	// client, err := ai.NewClientWithOptions(
	//     apiEndpoint,
	//     ai.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
	//     ai.WithUserAgent("custom-app/1.0"),
	// )

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new prompt
	createPromptExample(ctx, client)

	// Get a prompt
	prompt := getPromptExample(ctx, client, "prompt-123")
	if prompt == nil {
		return
	}

	// Update a prompt
	updatePromptExample(ctx, client, prompt.ID)

	// List prompts
	listPromptsExample(ctx, client)

	// Delete a prompt
	deletePromptExample(ctx, client, prompt.ID)
}

func createPromptExample(ctx context.Context, client *ai.Client) {
	fmt.Println("\n=== Creating a new prompt ===")

	// Define the prompt creation request
	createRequest := &ai.CreatePromptRequest{
		Name:        "Product Description Generator",
		Description: "Generates compelling product descriptions for e-commerce",
		Template:    "Create a compelling description for {{product_name}} that highlights its {{feature}} and appeals to {{target_audience}}.",
		ModelID:     "gpt-4",
		Parameters: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  300,
		},
		Variables: []ai.PromptVariable{
			{
				Name:         "product_name",
				Description:  "The name of the product",
				Required:     true,
			},
			{
				Name:         "feature",
				Description:  "Main feature to highlight",
				DefaultValue: "unique selling point",
				Required:     true,
			},
			{
				Name:         "target_audience",
				Description:  "Target customer demographic",
				DefaultValue: "general consumers",
				Required:     false,
			},
		},
		Tags: []string{"e-commerce", "marketing", "product"},
	}

	prompt, err := client.CreatePrompt(ctx, createRequest)
	if err != nil {
		fmt.Printf("Error creating prompt: %v\n", err)
		return
	}

	fmt.Printf("Created prompt successfully:\n")
	fmt.Printf("  ID: %s\n", prompt.ID)
	fmt.Printf("  Name: %s\n", prompt.Name)
	fmt.Printf("  Version: %d\n", prompt.Version)
	fmt.Printf("  Created: %s\n", prompt.CreatedAt)
}

func getPromptExample(ctx context.Context, client *ai.Client, promptID string) *ai.Prompt {
	fmt.Println("\n=== Getting a prompt ===")

	prompt, err := client.GetPrompt(ctx, promptID)
	if err != nil {
		fmt.Printf("Error getting prompt: %v\n", err)
		return nil
	}

	fmt.Printf("Retrieved prompt:\n")
	fmt.Printf("  ID: %s\n", prompt.ID)
	fmt.Printf("  Name: %s\n", prompt.Name)
	fmt.Printf("  Template: %s\n", prompt.Template)
	fmt.Printf("  Variables: %d\n", len(prompt.Variables))
	fmt.Printf("  Created: %s\n", prompt.CreatedAt)
	fmt.Printf("  Updated: %s\n", prompt.UpdatedAt)
	fmt.Printf("  Version: %d\n", prompt.Version)

	return prompt
}

func updatePromptExample(ctx context.Context, client *ai.Client, promptID string) {
	fmt.Println("\n=== Updating a prompt ===")

	// Define the fields to update
	newName := "Enhanced Product Description Generator"
	newTemplate := "Write a compelling and SEO-friendly description for {{product_name}} that highlights its {{feature}} and appeals to {{target_audience}}. Include at least 3 benefits."
	
	updateRequest := &ai.UpdatePromptRequest{
		Name:     &newName,
		Template: &newTemplate,
		Parameters: map[string]interface{}{
			"temperature": 0.8,
			"max_tokens":  500,
		},
		Tags: []string{"e-commerce", "marketing", "product", "seo"},
	}

	prompt, err := client.UpdatePrompt(ctx, promptID, updateRequest)
	if err != nil {
		fmt.Printf("Error updating prompt: %v\n", err)
		return
	}

	fmt.Printf("Updated prompt:\n")
	fmt.Printf("  ID: %s\n", prompt.ID)
	fmt.Printf("  Name: %s\n", prompt.Name)
	fmt.Printf("  Template: %s\n", prompt.Template)
	fmt.Printf("  Tags: %v\n", prompt.Tags)
	fmt.Printf("  Version: %d\n", prompt.Version)
}

func listPromptsExample(ctx context.Context, client *ai.Client) {
	fmt.Println("\n=== Listing prompts ===")

	// Define filter options
	options := &ai.ListPromptsOptions{
		Tags:       []string{"marketing"},
		MaxResults: 10,
	}

	prompts, nextToken, err := client.ListPrompts(ctx, options)
	if err != nil {
		fmt.Printf("Error listing prompts: %v\n", err)
		return
	}

	fmt.Printf("Retrieved %d prompts:\n", len(prompts))
	for i, p := range prompts {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, p.Name, p.ID)
	}

	if nextToken != "" {
		fmt.Printf("More results available with next token: %s\n", nextToken)
	}
}

func deletePromptExample(ctx context.Context, client *ai.Client, promptID string) {
	fmt.Println("\n=== Deleting a prompt ===")

	err := client.DeletePrompt(ctx, promptID)
	if err != nil {
		fmt.Printf("Error deleting prompt: %v\n", err)
		return
	}

	fmt.Printf("Successfully deleted prompt with ID: %s\n", promptID)
} 