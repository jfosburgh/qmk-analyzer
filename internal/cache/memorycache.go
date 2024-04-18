package cache

import (
	"time"
)

type Cache interface {
	Get(id string) (interface{}, bool)
	Set(id string, data interface{})
	Prune(maxAge time.Duration)
}

type SessionHandler struct {
	Cache Cache
}

type CacheEntry struct {
	LastAccessed time.Time
	Value        interface{}
}

type MemoryCache map[string]CacheEntry

func (c MemoryCache) Get(id string) (interface{}, bool) {
	data, ok := c[id]

	if !ok {
		return struct{}{}, false
	}

	return data.Value, true
}

func (c MemoryCache) Set(id string, data interface{}) {
	entry := CacheEntry{
		LastAccessed: time.Now(),
		Value:        data,
	}

	c[id] = entry
}

func (c MemoryCache) Prune(maxAge time.Duration) {
	currentTime := time.Now()
	for id, entry := range c {
		if currentTime.Sub(entry.LastAccessed) > maxAge {
			delete(c, id)
		}
	}
}

func NewMemoryCache() MemoryCache {
	return make(MemoryCache, 0)
}
