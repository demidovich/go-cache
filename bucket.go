package cache

import (
	"container/list"
	"sync"
)

type bucket struct {
	mu       sync.RWMutex
	data     map[string]row
	ll       *list.List
	capacity int
	hits     int
	misses   int
}

func newShard(capacity int) *bucket {
	return &bucket{
		capacity: capacity,
		data:     make(map[string]row),
		ll:       list.New(),
	}
}

type row struct {
	llElem *list.Element
	value  string
}

func (b *bucket) Get(k string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	row, ok := b.data[k]
	if !ok {
		b.misses++
		return "", false
	}

	b.ll.MoveToFront(row.llElem)
	b.hits++

	return row.value, true
}

func (b *bucket) Set(k string, v string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	row, ok := b.data[k]
	if ok {
		row.value = v
	} else {
		llElem := b.ll.PushFront(k)
		row.llElem = llElem
		row.value = v
		b.data[k] = row
	}
}

func (b *bucket) Delete(k string) {
	row, ok := b.data[k]
	if !ok {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.ll.Remove(row.llElem)
	delete(b.data, k)
}

func (b *bucket) Length() int {
	return len(b.data)
}

func (b *bucket) Hits() int {
	return b.hits
}

func (b *bucket) Misses() int {
	return b.misses
}

// Удаление значений, превышающих capacity шарда
func (b *bucket) Cleanup() {
	if len(b.data) < b.capacity {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for len(b.data) > b.capacity {
		elem := b.ll.Back()
		key, _ := elem.Value.(string)
		delete(b.data, key)
		b.ll.Remove(elem)
	}
}
