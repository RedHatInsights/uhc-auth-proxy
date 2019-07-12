package cache

import (
	"sync"
	"time"
)

type item struct {
	Object  []byte
	Expires time.Time
}

var mutex = &sync.Mutex{}
var cache = make(map[string]item)

// Get returns the item stored by key unless it is expired
func Get(key string) []byte {
	now := time.Now()
	mutex.Lock()
	b, ok := cache[key]
	mutex.Unlock()

	if !ok {
		return nil
	}

	if now.After(b.Expires) {
		return nil
	}

	return b.Object
}

// Set stores data via key
func Set(key string, data []byte) {
	v := item{
		Object:  data,
		Expires: time.Now().Add(2 * time.Hour),
	}
	mutex.Lock()
	cache[key] = v
	mutex.Unlock()
}

// Clear resets the cache
func Clear() {
	mutex.Lock()
	cache = make(map[string]item)
	mutex.Unlock()
}
