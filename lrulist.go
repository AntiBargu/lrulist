package lrulist // import "github.com/AntiBargu/lrulist"

import (
	"fmt"
	"sync"
)

type LRUListNode struct {
	key, val   interface{}
	prev, next *LRUListNode
}

type LRUList struct {
	// Capacity limit of the LRU list
	cap int
	// Mapping from cache keys to LRU nodes
	cacheMap map[interface{}]*LRUListNode
	// Head node of the LRU list
	cache *LRUListNode
	// Callback function for eviction when exceeding capacity
	evict func(interface{}) error
	// Read-write lock to protect concurrent access to the LRU list
	lock sync.RWMutex
}

func NewLRUList(cap int, evict func(interface{}) error) *LRUList {
	return &LRUList{
		cap:      cap,
		cacheMap: make(map[interface{}]*LRUListNode),
		cache:    nil,
		evict:    evict,
		lock:     sync.RWMutex{},
	}
}

func (lruc *LRUList) Set(key, val interface{}) error {
	lruc.lock.Lock()
	defer lruc.lock.Unlock()

	if item, hit := lruc.cacheMap[key]; hit {
		// If the key already exists in the cache
		if item != lruc.cache {
			// Move the node to the head of the LRU list
			item.prev.next, item.next.prev = item.next, item.prev

			item.prev, item.next = lruc.cache.prev, lruc.cache
			lruc.cache.prev.next, lruc.cache.prev = item, item
			lruc.cache = item
		}
		// Update the value of the head node
		lruc.cache.val = val
	} else {
		if len(lruc.cacheMap) < lruc.cap {
			// If the cache is not full, create new node
			item := &LRUListNode{key: key, val: val}
			lruc.cacheMap[key] = item

			if lruc.cache == nil {
				// The LRU list is empty, set the node as the head node
				item.prev, item.next = item, item
			} else {
				// Insert the node at the head of the LRU list
				item.prev, item.next = lruc.cache.prev, lruc.cache
				lruc.cache.prev.next, lruc.cache.prev = item, item
			}
			// Update the head node of the LRU list
			lruc.cache = item
		} else {
			// If the cache is full, replace the least recently used node
			lruc.cache = lruc.cache.prev

			if lruc.evict != nil {
				// Execute the callback function to handle the replaced node
				if err := lruc.evict(lruc.cache.val); err != nil {
					return err
				}
			}

			// Remove the replaced node from the cache mapping
			delete(lruc.cacheMap, lruc.cache.key)
			// Update the cache mapping
			lruc.cacheMap[key] = lruc.cache
			// Update the key and value of the head node
			lruc.cache.key, lruc.cache.val = key, val
		}
	}

	return nil
}

func (lruc *LRUList) Get(key interface{}) (interface{}, error) {
	lruc.lock.Lock()
	defer lruc.lock.Unlock()

	if item, hit := lruc.cacheMap[key]; hit {
		// If the key exists in the cache
		if item != lruc.cache {
			// Move the node to the head of the LRU list
			item.prev.next, item.next.prev = item.next, item.prev

			item.prev, item.next = lruc.cache.prev, lruc.cache
			lruc.cache.prev.next, lruc.cache.prev = item, item
			lruc.cache = item
		}
		return lruc.cache.val, nil
	} else {
		return nil, fmt.Errorf("key doesn't hit")
	}
}

func (lruc *LRUList) Traverse(visit func(interface{}) error) error {
	lruc.lock.RLock()
	defer lruc.lock.RUnlock()

	if lruc.cache == nil {
		return nil
	}

	// Visit the head node of the LRU list
	err := visit(lruc.cache.val)
	if err != nil {
		return err
	}
	for cur := lruc.cache.next; cur != lruc.cache; cur = cur.next {
		// Visit the other nodes of the LRU list
		err := visit(cur.val)
		if err != nil {
			return err
		}
	}

	return nil
}
