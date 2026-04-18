package cache

import "sync"

type Cache[T any] struct {
	mu   sync.RWMutex
	data map[string]T
}

func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		data: make(map[string]T),
	}
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	val, ok := c.data[key]
	c.mu.RUnlock()
	return val, ok
}

func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}
