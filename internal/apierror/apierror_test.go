package apierror

import "testing"

func TestErrorResponse_Error(t *testing.T) {
	tests := []struct {
		name        string
		errorCode   string
		description string
		want        string
	}{
		{
			name:        "with description",
			errorCode:   "test_error",
			description: "This is a test error",
			want:        "test_error: This is a test error",
		},
		{
			name:        "without description",
			errorCode:   "test_error",
			description: "",
			want:        "test_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ErrorResponse{
				ErrorCode:   tt.errorCode,
				Description: tt.description,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("ErrorResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
} 