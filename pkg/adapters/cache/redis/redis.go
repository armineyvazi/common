// Package redis provides a Redis cache adapter backed by go-redis v6 with Elastic APM tracing.
// For new projects, prefer the redis_v9 adapter.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis"
	"go.elastic.co/apm/module/apmgoredis/v2"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Redis wraps a go-redis v6 client with APM tracing.
type Redis struct {
	address  string
	password string
	conn     apmgoredis.Client
}

func New(address, password string, database int) ports.Cache {
	client := goredis.NewClient(&goredis.Options{
		Addr:     address,
		Password: password,
		DB:       database,
	})
	return &Redis{
		address:  address,
		password: password,
		conn:     apmgoredis.Wrap(client),
	}
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	res := r.conn.WithContext(ctx).Get(key)
	if res.Err() != nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) GetAll(ctx context.Context, keys ...string) ([]string, error) {
	res := r.conn.WithContext(ctx).MGet(keys...)
	if res.Err() != nil {
		return nil, res.Err()
	}
	j, err := json.Marshal(res.Val())
	if err != nil {
		return nil, err
	}
	var values []string
	if err := json.Unmarshal(j, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func (r *Redis) SearchKeys(ctx context.Context, pattern string) ([]string, error) {
	res := r.conn.WithContext(ctx).Keys(pattern)
	if res.Err() != nil {
		return nil, res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) HGet(ctx context.Context, key, field string) (string, error) {
	res, err := r.conn.WithContext(ctx).HGet(key, field).Result()
	if errors.Is(err, goredis.Nil) {
		return "", ports.ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return res, nil
}

func (r *Redis) Set(ctx context.Context, key string, data any, exp time.Duration) error {
	return r.conn.WithContext(ctx).Set(key, data, exp).Err()
}

func (r *Redis) BulkSet(ctx context.Context, data map[string]interface{}, exp time.Duration) error {
	pipe := r.conn.WithContext(ctx).Pipeline()
	for k, v := range data {
		pipe.Set(k, v, exp)
	}
	cmds, err := pipe.Exec()
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			return cmd.Err()
		}
	}
	return nil
}

func (r *Redis) HSet(ctx context.Context, key, field string, data any) error {
	return r.conn.WithContext(ctx).HSet(key, field, data).Err()
}

func (r *Redis) HSetAll(ctx context.Context, key string, data map[string]any) error {
	return r.conn.WithContext(ctx).HMSet(key, data).Err()
}

func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	res, err := r.conn.WithContext(ctx).HGetAll(key).Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func (r *Redis) HIncrBy(ctx context.Context, key, field string, incr int64) error {
	return r.conn.WithContext(ctx).HIncrBy(key, field, incr).Err()
}

func (r *Redis) HSetStruct(ctx context.Context, key string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal struct: %w", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("unmarshal struct to map: %w", err)
	}
	return r.conn.WithContext(ctx).HMSet(key, m).Err()
}

func (r *Redis) HGetStruct(ctx context.Context, key string, out any) error {
	res, err := r.conn.WithContext(ctx).HGetAll(key).Result()
	if err != nil {
		return err
	}
	b, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("marshal hash: %w", err)
	}
	return json.Unmarshal(b, out)
}

func (r *Redis) Del(ctx context.Context, key string) error {
	return r.conn.WithContext(ctx).Del(key).Err()
}

func (r *Redis) SAdd(ctx context.Context, key string, members interface{}) error {
	return r.conn.WithContext(ctx).SAdd(key, members).Err()
}

func (r *Redis) SRem(ctx context.Context, key string, members interface{}) error {
	return r.conn.WithContext(ctx).SRem(key, members).Err()
}

func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	res := r.conn.WithContext(ctx).SIsMember(key, member)
	return res.Val(), res.Err()
}

func (r *Redis) SCard(ctx context.Context, key string) (int64, error) {
	res := r.conn.WithContext(ctx).SCard(key)
	return res.Val(), res.Err()
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	res := r.conn.WithContext(ctx).SMembers(key)
	return res.Val(), res.Err()
}

func (r *Redis) ServiceName() string {
	return fmt.Sprintf("redis_%s", r.address)
}

func (r *Redis) IsHealthy(ctx context.Context) bool {
	return r.conn.WithContext(ctx).Ping().Err() == nil
}
