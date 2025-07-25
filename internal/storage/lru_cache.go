package storage

import (
	"container/list"
	"github.com/mahdi-cpp/photocloud_v2/pkg/happle_models"
	"sync"
)

// LRUCache implements a Least Recently Used memory
type LRUCache struct {
	capacity int
	cache    map[int]*list.Element
	list     *list.List
	mu       sync.Mutex
}

type cacheItem struct {
	key   int
	asset *happle_models.PHAsset
}

// NewLRUCache creates a new LRU memory
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves an asset from memory
func (c *LRUCache) Get(id int) (*happle_models.PHAsset, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[id]; exists {
		c.list.MoveToFront(elem)
		return elem.Value.(*cacheItem).asset, true
	}
	return nil, false
}

// Put adds an asset to memory
func (c *LRUCache) Put(id int, asset *happle_models.PHAsset) {
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

// Remove deletes an asset from memory
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

// Len returns current memory size
func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.Len()
}
