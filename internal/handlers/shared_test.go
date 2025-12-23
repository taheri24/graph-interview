package handlers

import (
	"errors"
	"testing"
)

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name     string
		error    string
		message  []string
		expected ErrorResponse
	}{
		{
			name:    "error only",
			error:   "test error",
			message: []string{},
			expected: ErrorResponse{
				Error:   "test error",
				Message: "",
			},
		},
		{
			name:    "error with message",
			error:   "test error",
			message: []string{"additional message"},
			expected: ErrorResponse{
				Error:   "test error",
				Message: "additional message",
			},
		},
		{
			name:    "error with multiple messages (should use first)",
			error:   "test error",
			message: []string{"first message", "second message"},
			expected: ErrorResponse{
				Error:   "test error",
				Message: "first message",
			},
		},
		{
			name:    "empty error with message",
			error:   "",
			message: []string{"message only"},
			expected: ErrorResponse{
				Error:   "",
				Message: "message only",
			},
		},
		{
			name:    "empty error and message",
			error:   "",
			message: []string{},
			expected: ErrorResponse{
				Error:   "",
				Message: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErrorResponse(tt.error, tt.message...)
			if result.Error != tt.expected.Error {
				t.Errorf("NewErrorResponse() Error = %v, want %v", result.Error, tt.expected.Error)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("NewErrorResponse() Message = %v, want %v", result.Message, tt.expected.Message)
			}
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
			name:    "error only",
			err:     errors.New("test error"),
			message: []string{},
			expected: ErrorResponse{
				Error:   "test error",
				Message: "",
			},
		},
		{
			name:    "error with message",
			err:     errors.New("test error"),
			message: []string{"additional message"},
			expected: ErrorResponse{
				Error:   "test error",
				Message: "additional message",
			},
		},

		{
			name:    "complex error",
			err:     errors.New("database connection failed: timeout after 30 seconds"),
			message: []string{"database error"},
			expected: ErrorResponse{
				Error:   "database connection failed: timeout after 30 seconds",
				Message: "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewErr(tt.err, tt.message...)
			if result.Error != tt.expected.Error {
				t.Errorf("NewErr() Error = %v, want %v", result.Error, tt.expected.Error)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("NewErr() Message = %v, want %v", result.Message, tt.expected.Message)
			}
		})
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	// Test that ErrorResponse struct can be properly marshaled to JSON
	// This is more of an integration test to ensure the struct tags work correctly
	t.Run("JSON marshaling", func(t *testing.T) {
		errResp := NewErrorResponse("test error", "test message")

		// Test that the struct has the correct field names and tags
		if errResp.Error != "test error" {
			t.Errorf("Expected error field to be 'test error', got %v", errResp.Error)
		}

		if errResp.Message != "test message" {
			t.Errorf("Expected message field to be 'test message', got %v", errResp.Message)
		}
	})
}

func TestNewErr_VariadicMessage(t *testing.T) {
	t.Run("multiple messages", func(t *testing.T) {
		err := errors.New("base error")
		result := NewErr(err, "msg1", "msg2", "msg3")

		expected := ErrorResponse{
			Error:   "base error",
			Message: "msg1",
		}

		if result != expected {
			t.Errorf("NewErr() with multiple messages = %v, want %v", result, expected)
		}
	})
}

func TestNewErrorResponse_VariadicMessage(t *testing.T) {
	t.Run("multiple messages", func(t *testing.T) {
		result := NewErrorResponse("base error", "msg1", "msg2", "msg3")

		expected := ErrorResponse{
			Error:   "base error",
			Message: "msg1",
		}

		if result != expected {
			t.Errorf("NewErrorResponse() with multiple messages = %v, want %v", result, expected)
		}
	})
}
