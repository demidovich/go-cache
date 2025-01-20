package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type Cache struct {
	mu         sync.RWMutex
	data       map[string]row
	ll         *list.List
	size       int
	gcInterval time.Duration
	hits       int
	misses     int
}

type Config struct {
	Size       int
	GcInterval time.Duration
}

func NewCache(ctx context.Context, config Config) *Cache {
	return NewCacheWithConfig(ctx, Config{
		Size:       100000,
		GcInterval: 10 * time.Second,
	})
}

func NewCacheWithConfig(ctx context.Context, config Config) *Cache {
	cache := &Cache{
		data:       make(map[string]row),
		ll:         list.New(),
		size:       config.Size,
		gcInterval: config.GcInterval,
	}

	go func() {
		cleanup := time.Tick(cache.gcInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-cleanup:
				cache.gcCleanup()
			}
		}
	}()

	return cache
}

type row struct {
	llElem *list.Element
	value  string
}

func (c *Cache) Get(k string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	row, ok := c.data[k]
	if !ok {
		c.misses++
		return "", false
	}

	c.ll.MoveToFront(row.llElem)
	c.hits++

	return row.value, true
}

func (c *Cache) Set(k string, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	row, ok := c.data[k]
	if ok {
		row.value = v
	} else {
		llElem := c.ll.PushFront(k)
		row.llElem = llElem
		row.value = v
		c.data[k] = row
	}
}

func (c *Cache) Delete(k string) {
	row, ok := c.data[k]
	if !ok {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll.Remove(row.llElem)
	delete(c.data, k)
}

func (c *Cache) Hits() int {
	return c.hits
}

func (c *Cache) Misses() int {
	return c.misses
}

func (c *Cache) gcCleanup() {
	if len(c.data) < c.size {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for len(c.data) > c.size {
		elem := c.ll.Back()
		key, _ := elem.Value.(string)
		delete(c.data, key)
		c.ll.Remove(elem)
	}
}
