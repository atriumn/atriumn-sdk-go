# Atriumn Auth Client

A Go client library for the Atriumn Authentication Service. This package provides an easy way to interact with the Atriumn Auth API from other Go services.

## Installation

```bash
go get github.com/atriumn/atriumn-sdk-go/auth
```

## Usage

### Creating a Client

```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/atriumn/atriumn-sdk-go/auth"
)

func main() {
    // Create a client with default settings
    client, err := auth.NewClient("https://api.example.com")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Or create with custom options
    customClient, err := auth.NewClientWithOptions(
        "https://api.example.com",
        auth.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
        auth.WithUserAgent("my-service/1.0"),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Use the client...
}
```

### Health Check

```go
ctx := context.Background()
health, err := client.Health(ctx)
if err != nil {
    log.Fatalf("Health check failed: %v", err)
}
fmt.Printf("API status: %s\n", health.Status)
```

### Client Credentials Flow

```go
ctx := context.Background()
token, err := client.GetClientCredentialsToken(
    ctx,
    "client-id",
    "client-secret",
    "requested-scope", // Optional, can be empty string
)
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}

fmt.Printf("Access Token: %s\n", token.AccessToken)
fmt.Printf("Token Type: %s\n", token.TokenType)
fmt.Printf("Expires In: %d seconds\n", token.ExpiresIn)
```

### User Signup

```go
ctx := context.Background()
attributes := map[string]string{
    "name": "John Doe",
    "custom:role": "user",
}
result, err := client.SignupUser(ctx, "user@example.com", "password123", attributes)
if err != nil {
    log.Fatalf("Signup failed: %v", err)
}

fmt.Printf("User ID: %s\n", result.UserID)
```

### User Login

```go
ctx := context.Background()
token, err := client.LoginUser(ctx, "user@example.com", "password123")
if err != nil {
    log.Fatalf("Login failed: %v", err)
}

fmt.Printf("Access Token: %s\n", token.AccessToken)
fmt.Printf("ID Token: %s\n", token.IDToken)
fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
```

### User Logout

```go
ctx := context.Background()
err := client.LogoutUser(ctx, "access-token-from-login")
if err != nil {
    log.Fatalf("Logout failed: %v", err)
}

fmt.Println("User logged out successfully")
```

### Password Reset

```go
ctx := context.Background()

// Request password reset
response, err := client.RequestPasswordReset(ctx, "user@example.com")
if err != nil {
    log.Fatalf("Password reset request failed: %v", err)
}

fmt.Printf("Code sent to: %s\n", response.CodeDeliveryDetails.Destination)

// Confirm password reset with code
err = client.ConfirmPasswordReset(ctx, "user@example.com", "123456", "new-password")
if err != nil {
    log.Fatalf("Password reset confirmation failed: %v", err)
}

fmt.Println("Password reset successfully")
```

### Error Handling

```go
ctx := context.Background()
_, err := client.LoginUser(ctx, "wrong@example.com", "wrong-password")
if err != nil {
    // Check if it's an API error
    if apiErr, ok := err.(*auth.ErrorResponse); ok {
        fmt.Printf("API Error: %s - %s\n", apiErr.Error, apiErr.Description)
    } else {
        fmt.Printf("Network or other error: %v\n", err)
    }
}
```

## Development

### Running Tests

```bash
cd atriumn-sdk-go
go test -v ./auth
```
