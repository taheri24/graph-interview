package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisCacheImpl(t *testing.T) {
	// Test that RedisCacheImpl implements CacheInterface
	var _ CacheInterface[string] = (*RedisCacheImpl[string])(nil)

	// Test struct creation
	impl := &RedisCacheImpl[string]{
		sectionName: "test",
		redisCache:  nil, // nil for test
	}

	assert.Equal(t, "test", impl.sectionName)
	assert.Nil(t, impl.redisCache)
}
