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

var cache = make(map[string]val)
var mutex = &sync.Mutex{}

func makeKey(r *Registration) string {
	return fmt.Sprintf("%s:%s", r.ClusterID, r.AuthorizationToken)
}

// Get fetches the Identity from the cache
func Get(r *Registration) *Identity {
	now := time.Now()
	key := makeKey(r)
	mutex.Lock()
	i, ok := cache[key]
	mutex.Unlock()

	if !ok {
		return nil
	}

	if now.After(i.Expires) {
		return nil
	}

	return i.Identity
}

// Set caches ident for the registration
func Set(r *Registration, ident *Identity) {
	key := makeKey(r)
	v := val{
		Identity: ident,
		Expires:  time.Now().Add(time.Hour * 2),
	}
	mutex.Lock()
	cache[key] = v
	mutex.Unlock()
}
