package cache

import (
	"container/list"
	"sync"
)

type shard struct {
	mu       sync.RWMutex
	data     map[string]row
	ll       *list.List
	capacity int
	hits     int
	misses   int
}

func newShard(capacity int) *shard {
	return &shard{
		capacity: capacity,
		data:     make(map[string]row),
		ll:       list.New(),
	}
}

type row struct {
	llElem *list.Element
	value  string
}

func (s *shard) Get(k string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row, ok := s.data[k]
	if !ok {
		s.misses++
		return "", false
	}

	s.ll.MoveToFront(row.llElem)
	s.hits++

	return row.value, true
}

func (s *shard) Set(k string, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.data[k]
	if ok {
		row.value = v
	} else {
		llElem := s.ll.PushFront(k)
		row.llElem = llElem
		row.value = v
		s.data[k] = row
	}
}

func (s *shard) Delete(k string) {
	row, ok := s.data[k]
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ll.Remove(row.llElem)
	delete(s.data, k)
}

func (s *shard) Length() int {
	return len(s.data)
}

func (s *shard) Hits() int {
	return s.hits
}

func (s *shard) Misses() int {
	return s.misses
}

func (s *shard) GcCleanup() {
	if len(s.data) < s.capacity {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for len(s.data) > s.capacity {
		elem := s.ll.Back()
		key, _ := elem.Value.(string)
		delete(s.data, key)
		s.ll.Remove(elem)
	}
}
