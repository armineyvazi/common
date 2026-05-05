package ports

import (
	"context"
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, data any, exp time.Duration) error
	Del(ctx context.Context, key string) error
	HGet(ctx context.Context, key string, field string) (string, error)
	GetAll(ctx context.Context, keys ...string) ([]string, error)
	HSet(ctx context.Context, key string, field string, data any) error
	BulkSet(ctx context.Context, data map[string]interface{}, exp time.Duration) error
	SearchKeys(ctx context.Context, pattern string) ([]string, error)
	HIncrBy(ctx context.Context, key, field string, incr int64) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HGetStruct(ctx context.Context, key string, out any) error
	HSetAll(ctx context.Context, key string, data map[string]any) error
	HSetStruct(ctx context.Context, key string, data any) error
	SAdd(ctx context.Context, key string, members interface{}) error
	SRem(ctx context.Context, key string, members interface{}) error
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)
	SCard(ctx context.Context, key string) (int64, error)
	SMembers(ctx context.Context, key string) ([]string, error)
}
