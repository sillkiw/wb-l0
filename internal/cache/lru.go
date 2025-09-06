package cache

import (
	"container/list"
	"sync"
	"time"
)

type entry[V any] struct {
	key       string
	value     V
	expiresAt time.Time
}

type LRU[V any] struct {
	mu         sync.RWMutex
	ll         *list.List
	items      map[string]*list.Element
	maxEntries int
	ttl        time.Duration
}

func NewLRU[V any](maxEntries int, ttl time.Duration) *LRU[V] {
	return &LRU[V]{
		ll:         list.New(),
		items:      make(map[string]*list.Element),
		maxEntries: maxEntries,
		ttl:        ttl,
	}
}

func (c *LRU[V]) Get(key string) (v V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, found := c.items[key]; found {
		en := ele.Value.(*entry[V])
		if c.ttl > 0 && time.Now().After(en.expiresAt) {
			c.removeElement(ele)
			var zero V
			return zero, false
		}
		c.ll.MoveToFront(ele)
		return en.value, true
	}
	var zero V
	return zero, false
}

func (c *LRU[V]) Set(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, found := c.items[key]; found {
		en := ele.Value.(*entry[V])
		en.value = value
		if c.ttl > 0 {
			en.expiresAt = time.Now().Add(c.ttl)
		}
		c.ll.MoveToFront(ele)
		return
	}

	en := &entry[V]{key: key, value: value}
	if c.ttl > 0 {
		en.expiresAt = time.Now().Add(c.ttl)
	}
	ele := c.ll.PushFront(en)
	c.items[key] = ele

	if c.maxEntries > 0 && c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *LRU[V]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, found := c.items[key]; found {
		c.removeElement(ele)
	}
}

func (c *LRU[V]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ll = list.New()
	c.items = make(map[string]*list.Element)
}

func (c *LRU[V]) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRU[V]) removeElement(e *list.Element) {
	c.ll.Remove(e)
	en := e.Value.(*entry[V])
	delete(c.items, en.key)
}
