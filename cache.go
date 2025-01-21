package cache

import (
	"context"
	"hash/crc32"
	"time"
)

type Cache struct {
	buckets      []*bucket
	bucketsCount uint32
	capacity     int
	length       int
	hits         int
	misses       int
	gcInterval   time.Duration
}

type Config struct {
	Capacity   int
	Buckets    int
	GcInterval time.Duration
}

func NewCache(ctx context.Context, config Config) *Cache {
	return NewCacheWithConfig(ctx, Config{
		Capacity:   10000,
		Buckets:    10,
		GcInterval: 10 * time.Second,
	})
}

func NewCacheWithConfig(ctx context.Context, config Config) *Cache {
	if config.Buckets < 1 {
		config.Buckets = 1
	}

	cache := &Cache{
		capacity:     config.Capacity,
		buckets:      make([]*bucket, config.Buckets),
		bucketsCount: uint32(config.Buckets),
		gcInterval:   config.GcInterval,
	}

	bucketCapacity := config.Capacity / config.Buckets
	for i := 0; i < config.Buckets; i++ {
		cache.buckets[i] = newShard(bucketCapacity)
	}

	go func() {
		gc := time.NewTicker(cache.gcInterval)
		for {
			select {
			case <-ctx.Done():
				gc.Stop()
				return
			case <-gc.C:
				cache.gc()
			}
		}
	}()

	return cache
}

func (c *Cache) bucketByKey(k string) *bucket {
	i := crc32.ChecksumIEEE([]byte(k))
	s := i % c.bucketsCount

	return c.buckets[s]
}

func (c *Cache) Get(k string) (string, bool) {
	return c.bucketByKey(k).Get(k)
}

func (c *Cache) Set(k string, v string) {
	c.bucketByKey(k).Set(k, v)
}

func (c *Cache) Delete(k string) {
	c.bucketByKey(k).Delete(k)
}

func (c *Cache) Capacity() int {
	return c.capacity
}

func (c *Cache) Length() int {
	return c.length
}

func (c *Cache) Hits() int {
	return c.hits
}

func (c *Cache) Misses() int {
	return c.misses
}

// Очистка в бакетах значений, превышающих capacity
// Агрегация данных о текущем размере данных и попаданиях в кэш
func (c *Cache) gc() {
	var length, hits, misses int

	for _, bucket := range c.buckets {
		bucket.Cleanup()

		length += bucket.Length()
		hits += bucket.hits
		misses += bucket.misses
	}

	c.length = length
	c.hits = hits
	c.misses = misses
}
