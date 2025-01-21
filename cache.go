package cache

import (
	"context"
	"hash/crc32"
	"time"
)

type Cache struct {
	shards      []*shard
	shardsCount uint32
	capacity    int
	length      int
	hits        int
	misses      int
	gcInterval  time.Duration
}

type Config struct {
	Capacity   int
	Shards     int
	GcInterval time.Duration
}

func NewCache(ctx context.Context, config Config) *Cache {
	return NewCacheWithConfig(ctx, Config{
		Capacity:   10000,
		Shards:     10,
		GcInterval: 10 * time.Second,
	})
}

func NewCacheWithConfig(ctx context.Context, config Config) *Cache {
	if config.Shards < 1 {
		config.Shards = 1
	}

	cache := &Cache{
		capacity:    config.Capacity,
		shards:      make([]*shard, config.Shards),
		shardsCount: uint32(config.Shards),
		gcInterval:  config.GcInterval,
	}

	shardCapacity := config.Capacity / config.Shards
	for i := 0; i < config.Shards; i++ {
		cache.shards[i] = newShard(shardCapacity)
	}

	go func() {
		maintenance := time.NewTicker(cache.gcInterval)
		for {
			select {
			case <-ctx.Done():
				maintenance.Stop()
				return
			case <-maintenance.C:
				cache.maintenance()
			}
		}
	}()

	return cache
}

func (c *Cache) shardByKey(k string) *shard {
	i := crc32.ChecksumIEEE([]byte(k))
	s := i % c.shardsCount

	return c.shards[s]
}

func (c *Cache) Get(k string) (string, bool) {
	return c.shardByKey(k).Get(k)
}

func (c *Cache) Set(k string, v string) {
	c.shardByKey(k).Set(k, v)
}

func (c *Cache) Delete(k string) {
	c.shardByKey(k).Delete(k)
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

// Очистка в шардах значений, превышающих capacity
// Агрегация данных о текущем размере данных и попаданиях в кэш
func (c *Cache) maintenance() {
	var length, hits, misses int

	for _, shard := range c.shards {
		shard.Cleanup()

		length += shard.Length()
		hits += shard.hits
		misses += shard.misses
	}

	c.length = length
	c.hits = hits
	c.misses = misses
}
