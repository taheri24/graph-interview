package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taheri24.ir/graph1/internal/models"
)

func TestNewRedisCache_Success(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	// Test basic functionality using generic functions
	testTask := models.Task{ID: uuid.New(), Title: "Test Task"}

	// Test SetT
	err = Set(cache, testTask, "task:%s", testTask.ID.String())
	assert.NoError(t, err)

	// Test GetT
	retrieved, err := Get[models.Task](cache, "task:%s", testTask.ID.String())
	assert.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, testTask.ID, retrieved.ID)
	assert.Equal(t, testTask.Title, retrieved.Title)
}

func TestNewRedisCache_ConnectionFailure(t *testing.T) {
	// Try to connect to invalid address
	cache, err := NewRedisCache("invalid:6379", "", 0)
	assert.Error(t, err)
	assert.Nil(t, cache)
}

func TestRedisCache_GetT_CacheMiss(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer cache.Close()

	// Test cache miss
	retrieved, err := Get[models.Task](cache, "nonexistent_key")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestRedisCache_SetT_JSONMarshalError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	cache, err := NewRedisCache(mr.Addr(), "", 0)
	require.NoError(t, err)
	defer cache.Close()

	// Try to set a value that cannot be marshaled (function)
	invalidValue := func() {} // functions cannot be JSON marshaled

	err = Set(cache, invalidValue, "test_key")
	assert.Error(t, err)
}
