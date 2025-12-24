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

func TestJsonEncode(t *testing.T) {
	// Test encoding a string
	encoded := JsonEncode("hello world")
	assert.Equal(t, []byte(`"hello world"`), encoded)

	// Test encoding an integer
	encodedInt := JsonEncode(42)
	assert.Equal(t, []byte("42"), encodedInt)

	// Test encoding a boolean
	encodedBool := JsonEncode(true)
	assert.Equal(t, []byte("true"), encodedBool)

	// Test encoding a slice
	encodedSlice := JsonEncode([]int{1, 2, 3})
	assert.Equal(t, []byte("[1,2,3]"), encodedSlice)

	// Test encoding a struct
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	encodedStruct := JsonEncode(TestStruct{Name: "test", Value: 123})
	assert.Equal(t, []byte(`{"name":"test","value":123}`), encodedStruct)
}

func TestJsonDecode(t *testing.T) {
	// Test decoding a string
	result := JsonDecode[string]([]byte(`"hello world"`))
	assert.Equal(t, "hello world", result)

	// Test decoding an integer
	resultInt := JsonDecode[int]([]byte("42"))
	assert.Equal(t, 42, resultInt)

	// Test decoding a boolean
	resultBool := JsonDecode[bool]([]byte("true"))
	assert.Equal(t, true, resultBool)

	// Test decoding a slice
	resultSlice := JsonDecode[[]int]([]byte("[1,2,3]"))
	assert.Equal(t, []int{1, 2, 3}, resultSlice)

	// Test decoding a struct
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	resultStruct := JsonDecode[TestStruct]([]byte(`{"name":"test","value":123}`))
	expected := TestStruct{Name: "test", Value: 123}
	assert.Equal(t, expected, resultStruct)
}

func TestJsonDecode_Panic(t *testing.T) {
	// Test that JsonDecode panics with invalid JSON
	assert.Panics(t, func() {
		JsonDecode[string]([]byte("invalid json"))
	})

	assert.Panics(t, func() {
		JsonDecode[int]([]byte("not a number"))
	})

	// Test with incomplete JSON
	assert.Panics(t, func() {
		JsonDecode[map[string]interface{}]([]byte(`{"key":`))
	})
}
