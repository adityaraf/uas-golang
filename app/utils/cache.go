package utils

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type PermissionCache struct {
	items sync.Map
	mu    sync.RWMutex
}

var Cache *PermissionCache

func InitCache() {
	Cache = &PermissionCache{}
	// Start cleanup goroutine
	go Cache.cleanupExpired()
}

// Set menyimpan data ke cache dengan TTL
func (c *PermissionCache) Set(key string, value interface{}, ttl time.Duration) {
	expiration := time.Now().Add(ttl).UnixNano()
	c.items.Store(key, CacheItem{
		Value:      value,
		Expiration: expiration,
	})
}

// Get mengambil data dari cache
func (c *PermissionCache) Get(key string) (interface{}, bool) {
	item, found := c.items.Load(key)
	if !found {
		return nil, false
	}

	cacheItem := item.(CacheItem)

	// Check if expired
	if time.Now().UnixNano() > cacheItem.Expiration {
		c.items.Delete(key)
		return nil, false
	}

	return cacheItem.Value, true
}

// Delete menghapus data dari cache
func (c *PermissionCache) Delete(key string) {
	c.items.Delete(key)
}

// Clear menghapus semua data dari cache
func (c *PermissionCache) Clear() {
	c.items.Range(func(key, value interface{}) bool {
		c.items.Delete(key)
		return true
	})
}

// cleanupExpired membersihkan item yang sudah expired setiap 5 menit
func (c *PermissionCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().UnixNano()
		c.items.Range(func(key, value interface{}) bool {
			item := value.(CacheItem)
			if now > item.Expiration {
				c.items.Delete(key)
			}
			return true
		})
	}
}