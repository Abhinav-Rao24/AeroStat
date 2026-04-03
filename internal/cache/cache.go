package cache

import (
	"sync"
	"time"

	"go-weather-cli/internal/models"
)

type entry struct {
	data      *models.WeatherResponse
	expiresAt time.Time
}

// Cache provides a thread-safe TTL cache.
type Cache struct {
	mu    sync.RWMutex
	store map[string]entry
	ttl   time.Duration
}

func New(ttl time.Duration) *Cache {
	return &Cache{
		store: make(map[string]entry),
		ttl:   ttl,
	}
}

func (c *Cache) Get(key string) (*models.WeatherResponse, bool) {
	if c.ttl == 0 {
		return nil, false
	}

	c.mu.RLock()
	e, ok := c.store[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.data, true
}

func (c *Cache) Set(key string, data *models.WeatherResponse) {
	if c.ttl == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, e := range c.store {
		if now.After(e.expiresAt) {
			delete(c.store, k)
		}
	}

	c.store[key] = entry{
		data:      data,
		expiresAt: now.Add(c.ttl),
	}
}
