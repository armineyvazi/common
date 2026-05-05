package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/armineyvazi/common.git/pkg/ports"
	"go.elastic.co/apm/module/apmgoredis/v2"
	"time"
)

type Redis struct {
	address  string
	password string
	conn     apmgoredis.Client
	ports.Cache
}

func New(address, password string, database int) ports.Cache {
	client := redis.NewClient(&redis.Options{
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
	c := r.conn.WithContext(ctx)
	res := c.Get(key)
	if res.Err() != nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) GetAll(ctx context.Context, keys ...string) ([]string, error) {
	c := r.conn.WithContext(ctx)
	res := c.MGet(keys...)
	if res.Err() != nil {
		return nil, res.Err()
	}
	values := make([]string, 0)
	j, err := json.Marshal(res.Val())
	if err != nil {
		return nil, res.Err()
	}
	err = json.Unmarshal(j, &values)
	if err != nil {
		return nil, res.Err()
	}
	return values, nil
}

func (r *Redis) SearchKeys(ctx context.Context, pattern string) ([]string, error) {
	c := r.conn.WithContext(ctx)
	res := c.Keys(pattern)
	if res.Err() != nil {
		return nil, res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) HGet(ctx context.Context, key string, field string) (string, error) {
	c := r.conn.WithContext(ctx)
	res, err := c.HGet(key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ports.ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return res, nil
}

func (r *Redis) Set(ctx context.Context, key string, data interface{}, exp time.Duration) error {
	c := r.conn.WithContext(ctx)
	res := c.Set(key, data, exp)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (r *Redis) BulkSet(ctx context.Context, data map[string]interface{}, exp time.Duration) error {
	pipe := r.conn.WithContext(ctx).Pipeline()
	for k, v := range data {
		pipe.Set(k, v, exp)
	}
	value, err := pipe.Exec()
	if err != nil {
		return err
	}
	for _, v := range value {
		if v.Err() != nil {
			return err
		}
	}
	return nil
}

func (r *Redis) HSet(ctx context.Context, key string, field string, data interface{}) error {
	c := r.conn.WithContext(ctx)
	res := c.HSet(key, field, data)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (r *Redis) HSetAll(ctx context.Context, key string, data map[string]any) error {
	err := r.conn.WithContext(ctx).HMSet(key, data).Err()
	return err
}

func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	mapCmd := r.conn.WithContext(ctx).HGetAll(key)
	res, err := mapCmd.Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func (r *Redis) ServiceName() string {
	return fmt.Sprintf("redis_%s", r.address)
}

func (r *Redis) IsHealthy(ctx context.Context) bool {
	return r.conn.WithContext(ctx).Ping().Err() == nil
}

func (r *Redis) HIncrBy(ctx context.Context, key string, field string, incr int64) error {
	return r.conn.WithContext(ctx).HIncrBy(key, field, incr).Err()
}

func (r *Redis) Del(ctx context.Context, key string) error {
	return r.conn.WithContext(ctx).Del(key).Err()
}

func (r *Redis) SAdd(ctx context.Context, key string, members interface{}) error {
	c := r.conn.WithContext(ctx)
	return c.SAdd(key, members).Err()
}

func (r *Redis) SRem(ctx context.Context, key string, members interface{}) error {
	c := r.conn.WithContext(ctx)
	return c.SRem(key, members).Err()
}

func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	c := r.conn.WithContext(ctx)
	res := c.SIsMember(key, member)
	if res.Err() != nil {
		return false, res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) SCard(ctx context.Context, key string) (int64, error) {
	c := r.conn.WithContext(ctx)
	res := c.SCard(key)
	if res.Err() != nil {
		return 0, res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	c := r.conn.WithContext(ctx)
	res := c.SMembers(key)
	if res.Err() != nil {
		return []string{}, res.Err()
	}
	return res.Val(), nil
}
