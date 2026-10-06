package ttlcache_test

import (
	"fmt"
	"time"

	"github.com/armineyvazi/common.git/pkg/adapters/memory/ttlcache"
)

// ExampleNew_stringCache shows a typed string-keyed, string-valued cache
// with per-entry TTLs.
func ExampleNew_stringCache() {
	cache := ttlcache.New[string, string]()

	cache.Set("session:abc", "user:42", 5*time.Minute)
	cache.Set("session:xyz", "user:99", 10*time.Minute)

	if cache.Has("session:abc") {
		fmt.Println(cache.GetValue("session:abc"))
	}
	// Output:
	// user:42
}

// ExampleNew_intKeyCache shows using integer keys with struct values,
// a common pattern for user or product caches.
func ExampleNew_intKeyCache() {
	type User struct {
		Name  string
		Email string
	}

	cache := ttlcache.New[int, User]()
	cache.Set(42, User{Name: "Alice", Email: "alice@example.com"}, time.Hour)

	if item := cache.Get(42); item != nil {
		u := item.Value()
		fmt.Printf("name=%s email=%s\n", u.Name, u.Email)
	}
	// Output:
	// name=Alice email=alice@example.com
}

// ExampleNew_expiryCheck shows that Get returns nil for expired entries.
func ExampleNew_expiryCheck() {
	cache := ttlcache.New[string, int]()

	// Set with a very short TTL and wait for expiry.
	cache.Set("counter", 1, time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	if cache.Get("counter") == nil {
		fmt.Println("expired")
	}
	// Output:
	// expired
}

// ExampleNew_keysAndItems shows listing all live cache keys and entries.
func ExampleNew_keysAndItems() {
	cache := ttlcache.New[string, int]()
	cache.Set("a", 1, time.Hour)
	cache.Set("b", 2, time.Hour)

	keys := cache.Keys()
	fmt.Println("count:", len(keys))
	// Output:
	// count: 2
}

// ExampleNew_delete shows removing a single entry and verifying it is gone.
func ExampleNew_delete() {
	cache := ttlcache.New[string, bool]()
	cache.Set("flag:maintenance", true, time.Hour)
	cache.Delete("flag:maintenance")

	fmt.Println(cache.Has("flag:maintenance"))
	// Output:
	// false
}
