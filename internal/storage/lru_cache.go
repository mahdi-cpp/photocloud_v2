package storage

import (
	"container/list"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
	"sync"
)

// LRUCache implements a Least Recently Used cache
type LRUCache struct {
	capacity int
	cache    map[int]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type cacheItem struct {
	key   int
	asset *model.PHAsset
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves an asset from cache
func (c *LRUCache) Get(id int) (*model.PHAsset, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[id]; exists {
		c.list.MoveToFront(elem)
		return elem.Value.(*cacheItem).asset, true
	}
	return nil, false
}

// Put adds an asset to cache
func (c *LRUCache) Put(id int, asset *model.PHAsset) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing item
	if elem, exists := c.cache[id]; exists {
		c.list.MoveToFront(elem)
		elem.Value.(*cacheItem).asset = asset
		return
	}

	// Add new item
	item := &cacheItem{key: id, asset: asset}
	elem := c.list.PushFront(item)
	c.cache[id] = elem

	// Evict if over capacity
	if c.list.Len() > c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.removeElement(oldest)
		}
	}
}

// Remove deletes an asset from cache
func (c *LRUCache) Remove(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[id]; exists {
		c.removeElement(elem)
	}
}

// removeElement removes an element from cache
func (c *LRUCache) removeElement(elem *list.Element) {
	item := c.list.Remove(elem).(*cacheItem)
	delete(c.cache, item.key)
}

// Len returns current cache size
func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.Len()
}
