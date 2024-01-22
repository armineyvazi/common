package port

import (
	"context"
	"time"
)

type Catch interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, data interface{}, exp time.Duration) error
	HGet(ctx context.Context, key string, field string) (string, error)
	GetAll(ctx context.Context, keys ...string) ([]string, error)
	HSet(ctx context.Context, key string, field string, data interface{}) error
	BulkSet(ctx context.Context, data map[string]interface{}, exp time.Duration) error
	SearchKeys(ctx context.Context, pattern string) ([]string, error)
	Health
}
