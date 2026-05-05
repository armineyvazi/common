package ttlcache

import (
	"context"
	"github.com/jellydator/ttlcache/v3"
	"time"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type Memory[K comparable, V any] struct {
	cache *ttlcache.Cache[K, V]
}

func New[K comparable, V any](opts ...ttlcache.Option[K, V]) ports.Memory[K, V] {
	opts = append(opts, ttlcache.WithDisableTouchOnHit[K, V]())
	cache := ttlcache.New[K, V](opts...)

	memory := &Memory[K, V]{
		cache: cache,
	}
	memory.Start()
	return memory
}

func (m Memory[K, V]) Has(key K) bool {
	return m.cache.Has(key)
}

func (m Memory[K, V]) Start() {
	go m.cache.Start()
}

func (m Memory[K, V]) Set(key K, value V, ttl time.Duration) {
	m.cache.Set(key, value, ttl)
}

func (m Memory[K, V]) Get(key K) *ttlcache.Item[K, V] {
	return m.cache.Get(key)
}

func (m Memory[K, V]) GetValue(key K) V {
	return m.cache.Get(key).Value()
}

func (m Memory[K, V]) Items() map[K]*ttlcache.Item[K, V] {
	return m.cache.Items()
}

func (m Memory[K, V]) Keys() []K {
	return m.cache.Keys()
}

func (m Memory[K, V]) Delete(key K) {
	m.cache.Delete(key)
}

func (m Memory[K, V]) DeleteExpired() {
	m.cache.DeleteExpired()
}

func (m Memory[K, V]) DeleteAll() {
	m.cache.DeleteAll()
}

func (m Memory[K, V]) OnEviction(fn func(context.Context, ttlcache.EvictionReason, *ttlcache.Item[K, V])) func() {
	return m.cache.OnEviction(fn)
}

func (m Memory[K, V]) OnInsertion(fn func(context.Context, *ttlcache.Item[K, V])) func() {
	return m.cache.OnInsertion(fn)
}
