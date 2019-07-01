package cluster

import (
	"fmt"
	"sync"
	"time"
)

type val struct {
	Identity *Identity
	Expires  time.Time
}

func makeKey(r *Registration) string {
	return fmt.Sprintf("%s:%s", r.ClusterID, r.AuthorizationToken)
}

type TimedCache struct {
	ExpireDuration time.Duration
	mutex          *sync.Mutex
	cache          map[string]val
}

// NewTimedCache constructs a new TimedCache with the given duration as the
// expire timeout
func NewTimedCache(d time.Duration) *TimedCache {
	return &TimedCache{
		ExpireDuration: d,
		mutex:          &sync.Mutex{},
		cache:          make(map[string]val),
	}
}

// Get fetches the Identity from the cache
func (c *TimedCache) Get(r *Registration) *Identity {
	now := time.Now()
	key := makeKey(r)
	c.mutex.Lock()
	i, ok := c.cache[key]
	c.mutex.Unlock()

	if !ok {
		return nil
	}

	if now.After(i.Expires) {
		return nil
	}

	return i.Identity
}

// Set caches ident for the registration
func (c *TimedCache) Set(r *Registration, ident *Identity) {
	key := makeKey(r)
	v := val{
		Identity: ident,
		Expires:  time.Now().Add(c.ExpireDuration),
	}
	c.mutex.Lock()
	c.cache[key] = v
	c.mutex.Unlock()
}

// Cache is the default cache
var Cache = NewTimedCache(time.Hour * 2)
