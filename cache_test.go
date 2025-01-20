package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func cacheInstance(size int, gcIntervalSeconds int) *Cache {
	ctx := context.TODO()
	return NewCacheWithConfig(
		ctx,
		Config{
			Size:       size,
			GcInterval: time.Duration(gcIntervalSeconds) * time.Second,
		},
	)
}

func TestConfig(t *testing.T) {
	ctx := context.TODO()
	c := NewCacheWithConfig(
		ctx,
		Config{
			Size:       5,
			GcInterval: 5 * time.Second,
		},
	)

	assert.Equal(t, c.size, 5)
	assert.Equal(t, c.gcInterval, 5*time.Second)
}

func TestGetSet(t *testing.T) {
	c := cacheInstance(5, 5)
	c.Set("key1", "val1")
	val, ok := c.Get("key1")

	assert.Equal(t, "val1", val)
	assert.Equal(t, true, ok)
}

func TestGetMissing(t *testing.T) {
	c := cacheInstance(5, 5)
	val, ok := c.Get("missingKey")

	assert.Equal(t, "", val)
	assert.Equal(t, false, ok)
}

func TestOversize(t *testing.T) {
	c := cacheInstance(2, 1)
	c.Set("key1", "val1") // Это вычистит GC
	c.Set("key2", "val2")
	c.Set("key3", "val3")

	time.Sleep(2 * time.Second)

	val1, ok1 := c.Get("key1")

	assert.Equal(t, "", val1)
	assert.Equal(t, false, ok1)
}
