package scache

import (
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

// Cache stores the internal state.
type Cache struct {
	store *cache.Cache
	chans map[string]chan Response
	mu    sync.Mutex
}

// New creates an instance based on expiration and cleanup interval.
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		store: cache.New(defaultExpiration, cleanupInterval),
		chans: make(map[string]chan Response),
	}
}

// ResponseChan opens or returns channel for the cache response.
func (c *Cache) ResponseChan(key string, retriever Retriever) <-chan Response {
	c.mu.Lock()
	ch := c.chans[key]
	if ch == nil {
		// Nobody else waiting yet
		ch = make(chan Response, 1)
		val, found := c.store.Get(key)
		if found {
			// Return if already cached
			ch <- Response{Value: val, Error: nil}
		} else {
			// Otherwise save channel for others
			c.chans[key] = ch
			go func() {
				// And asynchronously retrieve value
				val, err := retriever(key)
				c.store.Set(key, val, cache.DefaultExpiration)
				ch <- Response{Value: val, Error: err}
				delete(c.chans, key)
			}()
		}
	}
	c.mu.Unlock()
	return ch
}

// Response stores value and/or error returned by the channel.
type Response struct {
	Value interface{}
	Error error
}

// Retriever is a function to retrieve value not already cached.
type Retriever func(string) (interface{}, error)
