package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Cache struct {
	store *gocache.Cache
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		store: gocache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.store.Set(key, value, duration)
}

func (c *Cache) Delete(key string) {
	c.store.Delete(key)
}

func (c *Cache) DeleteByPrefix(prefix string) {
	items := c.store.Items()
	for k := range items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			c.store.Delete(k)
		}
	}
}

func (c *Cache) Flush() {
	c.store.Flush()
}

func (c *Cache) ItemCount() int {
	return c.store.ItemCount()
}

