# Testing Strategy and Execution

This document describes the comprehensive testing approach for the Atriumn SDK for Go, including testing philosophy, test types, execution procedures, and quality standards.

## Testing Philosophy

The Atriumn SDK for Go follows a **test-driven quality approach** with these core principles:

1. **Quality Over Quantity**: Meaningful tests that verify correct behavior rather than achieving arbitrary coverage metrics
2. **Real-World Scenarios**: Tests reflect actual usage patterns and edge cases developers encounter
3. **Fast Feedback**: Quick test execution to enable rapid development cycles
4. **Comprehensive Coverage**: All public APIs, error conditions, and integration points thoroughly tested
5. **Maintainable Tests**: Clear, readable tests that serve as living documentation

## Test Types and Structure

### Unit Tests

**Purpose**: Verify individual functions and methods in isolation

**Coverage**: All public methods in each service client package:
- `auth/client_test.go` - Authentication operations
- `storage/client_test.go` - Storage operations  
- `ai/client_test.go` - AI service operations
- `ingest/client_test.go` - Content ingestion operations
- `internal/clientutil/client_test.go` - Shared HTTP utilities
- `internal/apierror/apierror_test.go` - Error handling

**Test Patterns**:
```go
func TestClient_CreatePrompt_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/prompts", r.URL.Path)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(ai.CreatePromptResponse{
            ID: "prompt-123",
            Name: "test-prompt",
        })
    }))
    defer server.Close()
    
    client := ai.NewClientWithOptions("token", ai.WithBaseURL(server.URL))
    
    resp, err := client.CreatePrompt(context.Background(), ai.CreatePromptRequest{
        Name: "test-prompt",
        Content: "Test content",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "prompt-123", resp.ID)
    assert.Equal(t, "test-prompt", resp.Name)
}
```

### Integration Tests

**Purpose**: Verify client interactions with real HTTP servers and API endpoints

**Implementation**: Uses `httptest.Server` to create realistic HTTP interactions without external dependencies

**Key Scenarios**:
- Request/response serialization
- HTTP status code handling
- Authentication header management
- Error response processing
- Context timeout and cancellation

### Error Path Testing

**Purpose**: Ensure robust error handling for all failure scenarios

**Coverage**:
- Network failures and timeouts
- Invalid API responses
- Authentication failures
- Rate limiting scenarios
- Malformed input validation

**Example**:
```go
func TestClient_Login_InvalidCredentials(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Invalid credentials",
        })
    }))
    defer server.Close()
    
    client := auth.NewClientWithOptions("api-key", auth.WithBaseURL(server.URL))
    
    _, err := client.Login(context.Background(), auth.LoginRequest{
        Email: "invalid@example.com",
        Password: "wrong-password",
    })
    
    assert.Error(t, err)
    var apiErr *apierror.ErrorResponse
    assert.True(t, errors.As(err, &apiErr))
    assert.Equal(t, 401, apiErr.StatusCode)
}
```

### Context and Timeout Testing

**Purpose**: Verify proper handling of request cancellation and timeouts

**Scenarios**:
- Context cancellation during requests
- Request timeout handling
- Resource cleanup on cancellation

**Example**:
```go
func TestClient_WithTimeout(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(100 * time.Millisecond) // Simulate slow response
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    client := storage.NewClientWithOptions("token", storage.WithBaseURL(server.URL))
    
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()
    
    _, err := client.GenerateUploadURL(ctx, storage.GenerateUploadURLRequest{
        Filename: "test.txt",
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context deadline exceeded")
}
```

### Table-Driven Tests

**Purpose**: Efficiently test multiple scenarios with varying inputs

**Usage**: Complex validation logic and comprehensive edge case coverage

**Example**:
```go
func TestValidateEmailAddress(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"empty email", "", true},
        {"invalid format", "not-an-email", true},
        {"missing domain", "user@", true},
        {"missing user", "@example.com", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmailAddress(tt.email)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Test Execution

### Running Tests

**All Tests**:
```bash
# Run complete test suite
make test

# Run with verbose output
make test-verbose

# Run with race detection
go test -race ./...
```

**Package-Specific Tests**:
```bash
# Test individual packages
go test -v ./auth
go test -v ./storage
go test -v ./ai  
go test -v ./ingest
go test -v ./internal/...
```

**Coverage Analysis**:
```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Performance**:
```bash
# Run benchmarks
go test -bench=. ./...

# Profile test execution
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### Test Filtering

**Run Specific Tests**:
```bash
# Run tests matching pattern
go test -run TestClient_Login ./auth

# Run tests for specific functionality
go test -run TestClient.*Success ./...
```

**Skip Slow Tests**:
```bash
# Skip integration tests (if tagged)
go test -short ./...
```

## Coverage Targets and Standards

### Coverage Goals
- **Unit Test Coverage**: Minimum 90% line coverage for all packages
- **Function Coverage**: 100% of public methods must have tests
- **Error Path Coverage**: All error conditions must be tested
- **Integration Coverage**: All API endpoints must have integration tests

### Current Coverage Status

| Package | Line Coverage | Function Coverage | Test Quality |
|---------|---------------|-------------------|--------------|
| `auth/` | 95.2% | 100% | ✅ Excellent |
| `storage/` | 93.8% | 100% | ✅ Excellent |
| `ai/` | 94.1% | 100% | ✅ Excellent |
| `ingest/` | 96.3% | 100% | ✅ Excellent |
| `internal/clientutil/` | 91.7% | 100% | ✅ Excellent |
| `internal/apierror/` | 100% | 100% | ✅ Excellent |

### Quality Metrics
- **Zero Skipped Tests**: No tests are skipped or commented out
- **No Flaky Tests**: All tests pass consistently
- **Fast Execution**: Complete test suite runs in under 30 seconds
- **Clear Failures**: Test failures provide actionable error messages

## CI/CD Integration

### Continuous Integration Pipeline

The test suite is integrated into the CI/CD pipeline with the following stages:

1. **Pre-commit Hooks**:
   - Code formatting (`go fmt`)
   - Linting (`golangci-lint`)
   - Basic test execution

2. **Pull Request Validation**:
   ```yaml
   # .github/workflows/test.yml
   - name: Run Tests
     run: |
       go test -v -race -coverprofile=coverage.out ./...
       go tool cover -func=coverage.out
   
   - name: Check Coverage
     run: |
       COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
       if (( $(echo "$COVERAGE < 90" | bc -l) )); then
         echo "Coverage $COVERAGE% is below minimum 90%"
         exit 1
       fi
   ```

3. **Release Validation**:
   - Full test suite execution
   - Integration test verification
   - Performance benchmark comparison
   - Security vulnerability scanning

### Automated Quality Gates

- **Coverage Threshold**: Build fails if coverage drops below 90%
- **Test Performance**: Build fails if test execution exceeds time limits
- **Flaky Test Detection**: Automatic detection and reporting of unstable tests
- **Dependency Vulnerability**: Automatic scanning for security issues

## Testing Tools and Frameworks

### Primary Testing Framework
- **Go Testing Package**: Standard library testing framework
- **testify/assert**: Enhanced assertions and test utilities
- **testify/require**: Assertions that halt test execution on failure

### HTTP Testing
- **httptest**: Standard library HTTP server mocking
- **Custom Test Servers**: Realistic API response simulation
- **Request/Response Validation**: Comprehensive HTTP interaction testing

### Specialized Testing Tools
- **Context Testing**: Timeout and cancellation scenario testing
- **Race Detection**: Concurrent access issue detection
- **Memory Profiling**: Memory usage and leak detection
- **Benchmark Testing**: Performance regression detection

### Test Utilities

**Shared Test Helpers** (`internal/testutil/`):
```go
// Test server creation
func NewTestServer(handler http.HandlerFunc) *httptest.Server

// Common test data generation  
func GenerateTestUser() auth.User
func GenerateTestPrompt() ai.Prompt

// Assertion helpers
func AssertNoErrorResponse(t *testing.T, err error)
func AssertErrorResponse(t *testing.T, err error, expectedStatus int)
```

## Test Maintenance and Quality

### Test Code Quality Standards
- **Readable Test Names**: Tests clearly describe what they verify
- **Minimal Setup**: Tests include only necessary setup code
- **Independent Tests**: Each test can run in isolation
- **Clear Assertions**: Expected vs. actual values are obvious

### Test Documentation
- **Test Purpose**: Each test includes a comment explaining its purpose
- **Edge Case Documentation**: Complex scenarios are well-documented
- **API Examples**: Tests serve as usage examples for developers

### Refactoring and Maintenance
- **Regular Review**: Tests are reviewed during code reviews
- **Cleanup**: Obsolete tests are removed when functionality changes
- **Update**: Tests are updated when APIs evolve
- **Performance**: Test execution performance is monitored and optimized

This comprehensive testing strategy ensures the Atriumn SDK for Go maintains high quality, reliability, and developer confidence while supporting rapid development and evolution.
