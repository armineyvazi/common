package ports

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type Memory[K comparable, V any] interface {
	Set(key K, value V, ttl time.Duration)
	Get(key K) *ttlcache.Item[K, V]
	GetValue(key K) V
	Items() map[K]*ttlcache.Item[K, V]
	Has(key K) bool
	Start()
	Keys() []K
	Delete(key K)
	DeleteExpired()
	DeleteAll()
	OnEviction(fn func(context.Context, ttlcache.EvictionReason, *ttlcache.Item[K, V])) func()
	OnInsertion(fn func(context.Context, *ttlcache.Item[K, V])) func()
}
