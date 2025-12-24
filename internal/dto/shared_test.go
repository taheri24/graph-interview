package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
		message  []string
		expected ErrorResponse
	}{
		{
			name:     "without message",
			errorMsg: "test error",
			message:  nil,
			expected: ErrorResponse{Error: "test error", Message: ""},
		},
		{
			name:     "with message",
			errorMsg: "test error",
			message:  []string{"additional info"},
			expected: ErrorResponse{Error: "test error", Message: "additional info"},
		},
		{
			name:     "multiple messages (takes first)",
			errorMsg: "test error",
			message:  []string{"first", "second"},
			expected: ErrorResponse{Error: "test error", Message: "first"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErrorResponse(tt.errorMsg, tt.message...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		message  []string
		expected ErrorResponse
	}{
		{
			name:     "error without message",
			err:      assert.AnError,
			message:  nil,
			expected: ErrorResponse{Error: assert.AnError.Error(), Message: ""},
		},
		{
			name:     "error with message",
			err:      assert.AnError,
			message:  []string{"context"},
			expected: ErrorResponse{Error: assert.AnError.Error(), Message: "context"},
		},
		{
			name:     "nil error panics",
			err:      nil,
			message:  nil,
			expected: ErrorResponse{}, // Won't reach here
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				assert.Panics(t, func() { NewErr(tt.err, tt.message...) })
			} else {
				result := NewErr(tt.err, tt.message...)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{Error: "test error", Message: "test message"}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	var unmarshaled ErrorResponse
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, resp, unmarshaled)

	expectedJSON := `{"error":"test error","message":"test message"}`
	assert.JSONEq(t, expectedJSON, string(data))
}

func TestErrorResponse_JSON_EmptyMessage(t *testing.T) {
	resp := ErrorResponse{Error: "test error"}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	expectedJSON := `{"error":"test error"}`
	assert.JSONEq(t, expectedJSON, string(data))
}
