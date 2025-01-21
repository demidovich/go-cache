package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func cacheInstance(capacity int, gcInterval time.Duration) *Cache {
	ctx := context.TODO()
	return NewCacheWithConfig(
		ctx,
		Config{
			Capacity:   capacity,
			GcInterval: gcInterval,
		},
	)
}

func TestConfig(t *testing.T) {
	ctx := context.TODO()
	c := NewCacheWithConfig(
		ctx,
		Config{
			Capacity:   3,
			Buckets:    2,
			GcInterval: 1 * time.Second,
		},
	)

	assert.Equal(t, 3, c.capacity)
	assert.Equal(t, 2, len(c.buckets))
	assert.Equal(t, 2, int(c.bucketsCount))
	assert.Equal(t, 1*time.Second, c.gcInterval)
}

func TestConfigWithoutBuckets(t *testing.T) {
	ctx := context.TODO()
	c := NewCacheWithConfig(
		ctx,
		Config{
			Capacity:   5,
			GcInterval: 5 * time.Second,
		},
	)

	assert.Equal(t, 1, len(c.buckets))
	assert.Equal(t, 1, int(c.bucketsCount))
}

func TestGetSet(t *testing.T) {
	c := cacheInstance(5, time.Second)

	c.Set("key1", "val1")
	val, ok := c.Get("key1")

	assert.Equal(t, "val1", val)
	assert.Equal(t, true, ok)
}

func TestGetMissing(t *testing.T) {
	c := cacheInstance(5, time.Second)

	val, ok := c.Get("missingKey")

	assert.Equal(t, "", val)
	assert.Equal(t, false, ok)
}

func TestDelete(t *testing.T) {
	c := cacheInstance(5, time.Second)

	c.Set("key1", "val1")
	c.Delete("key1")

	val1, ok1 := c.Get("key1")

	assert.Equal(t, val1, "")
	assert.Equal(t, false, ok1)
}

func TestCleanupOversize(t *testing.T) {
	c := cacheInstance(2, 10*time.Millisecond)

	c.Set("key1", "val1") // Это вычистит GC
	c.Set("key2", "val2")
	c.Set("key3", "val3")

	time.Sleep(20 * time.Millisecond)

	val1, ok1 := c.Get("key1")

	assert.Equal(t, "", val1)
	assert.Equal(t, false, ok1)
}

func TestCapacity(t *testing.T) {
	c := cacheInstance(5, 10*time.Millisecond)

	assert.Equal(t, 5, c.Capacity())
}

func TestLength(t *testing.T) {
	c := cacheInstance(5, 10*time.Millisecond)

	c.Set("key1", "val1")
	c.Set("key2", "val2")
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 2, c.Length())

	c.Delete("key1")
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 1, c.Length())
}

func TestHits(t *testing.T) {
	c := cacheInstance(5, 10*time.Millisecond)

	c.Set("key1", "val1")
	c.Get("key1")
	c.Get("key1")
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 2, c.Hits())
	assert.Equal(t, 0, c.Misses())
}

func TestMisses(t *testing.T) {
	c := cacheInstance(5, 10*time.Millisecond)

	c.Get("key1")
	c.Get("key1")
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, 2, c.Misses())
	assert.Equal(t, 0, c.Hits())
}
