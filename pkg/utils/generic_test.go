package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust_Success(t *testing.T) {
	// Test with successful result
	result := Must("test value", nil)
	assert.Equal(t, "test value", result)

	// Test with different types
	intResult := Must(42, nil)
	assert.Equal(t, 42, intResult)

	boolResult := Must(true, nil)
	assert.Equal(t, true, boolResult)

	sliceResult := Must([]string{"a", "b", "c"}, nil)
	assert.Equal(t, []string{"a", "b", "c"}, sliceResult)
}

func TestMust_Panic(t *testing.T) {
	// Test that Must panics when there's an error
	assert.Panics(t, func() {
		Must("", errors.New("test error"))
	})

	assert.Panics(t, func() {
		Must(0, errors.New("another error"))
	})

	// Test that the panic value is the error
	assert.PanicsWithError(t, "test error", func() {
		Must("", errors.New("test error"))
	})
}

func TestMust_NilError(t *testing.T) {
	// Test with nil error explicitly
	result := Must("success", nil)
	assert.Equal(t, "success", result)
}
