package ttlcache_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jellydator/ttlcache/v3"

	memcache "github.com/armineyvazi/common.git/pkg/adapters/memory/ttlcache"
)

func TestSetAndGet(t *testing.T) {
	c := memcache.New[string, int]()
	c.Set("a", 42, time.Minute)

	item := c.Get("a")
	if item == nil {
		t.Fatal("expected item, got nil")
	}
	if item.Value() != 42 {
		t.Fatalf("value: got %d, want 42", item.Value())
	}
}

func TestGetValue(t *testing.T) {
	c := memcache.New[string, string]()
	c.Set("k", "hello", time.Minute)
	if c.GetValue("k") != "hello" {
		t.Fatalf("GetValue: got %q, want %q", c.GetValue("k"), "hello")
	}
}

func TestHas(t *testing.T) {
	c := memcache.New[string, bool]()
	if c.Has("missing") {
		t.Error("Has returned true for missing key")
	}
	c.Set("present", true, time.Minute)
	if !c.Has("present") {
		t.Error("Has returned false for present key")
	}
}

func TestDelete(t *testing.T) {
	c := memcache.New[string, int]()
	c.Set("x", 1, time.Minute)
	c.Delete("x")
	if c.Has("x") {
		t.Error("key should be gone after Delete")
	}
}

func TestDeleteAll(t *testing.T) {
	c := memcache.New[string, int]()
	c.Set("a", 1, time.Minute)
	c.Set("b", 2, time.Minute)
	c.DeleteAll()
	if len(c.Keys()) != 0 {
		t.Errorf("expected 0 keys after DeleteAll, got %d", len(c.Keys()))
	}
}

func TestKeys(t *testing.T) {
	c := memcache.New[string, int]()
	c.Set("one", 1, time.Minute)
	c.Set("two", 2, time.Minute)
	keys := c.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestItems(t *testing.T) {
	c := memcache.New[string, int]()
	c.Set("m", 99, time.Minute)
	items := c.Items()
	if _, ok := items["m"]; !ok {
		t.Error("key 'm' missing from Items()")
	}
}

func TestTTLExpiry(t *testing.T) {
	c := memcache.New[string, int](ttlcache.WithTTL[string, int](30 * time.Millisecond))
	c.Set("exp", 1, ttlcache.DefaultTTL)

	time.Sleep(60 * time.Millisecond)
	c.DeleteExpired()

	if c.Has("exp") {
		t.Error("key should have expired")
	}
}

func TestOnEviction(t *testing.T) {
	c := memcache.New[string, int]()

	var mu sync.Mutex
	evicted := []string{}
	c.OnEviction(func(_ context.Context, _ ttlcache.EvictionReason, item *ttlcache.Item[string, int]) {
		mu.Lock()
		evicted = append(evicted, item.Key())
		mu.Unlock()
	})

	c.Set("ev", 1, time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	c.DeleteExpired()

	mu.Lock()
	n := len(evicted)
	mu.Unlock()
	if n == 0 {
		t.Error("eviction callback was not called")
	}
}

func TestOnInsertion(t *testing.T) {
	c := memcache.New[string, int]()

	inserted := make(chan string, 1)
	c.OnInsertion(func(_ context.Context, item *ttlcache.Item[string, int]) {
		inserted <- item.Key()
	})

	c.Set("ins", 5, time.Minute)

	select {
	case k := <-inserted:
		if k != "ins" {
			t.Errorf("insertion key: got %q, want %q", k, "ins")
		}
	case <-time.After(time.Second):
		t.Error("insertion callback timed out")
	}
}

func BenchmarkSet(b *testing.B) {
	c := memcache.New[int, int]()
	b.ReportAllocs()
	for b.Loop() {
		c.Set(b.N, b.N, time.Minute)
	}
}

func BenchmarkGet_Hit(b *testing.B) {
	c := memcache.New[string, int]()
	c.Set("key", 42, time.Minute)
	b.ReportAllocs()
	for b.Loop() {
		_ = c.Get("key")
	}
}

func BenchmarkSet_Parallel(b *testing.B) {
	c := memcache.New[int, int]()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.Set(i, i, time.Minute)
			i++
		}
	})
}
