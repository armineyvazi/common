package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"go.elastic.co/apm/module/apmgoredis/v2"

	"github.com/armineyvazi/common/pkg/port"
)

const (
	serviceName = "redis_%s"
)

type Redis struct {
	address  string
	password string
	conn     apmgoredis.Client
}

func New(address, password string, database int) port.Catch {
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

func (r *Redis) Get(ctx context.Context, key string) (result string, err error) {
	c := r.conn.WithContext(ctx)
	err = c.Get(key).Err()
	if err != nil {
		return result, err
	}
	result = c.Get(key).Val()
	return result, nil
}

func (r *Redis) GetAll(ctx context.Context, keys ...string) (results []string, err error) {
	c := r.conn.WithContext(ctx)
	err = c.MGet(keys...).Err()
	if err != nil {
		return results, err
	}
	res := c.MGet(keys...).Val()
	values := make([]string, 0)
	j, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func (r *Redis) SearchKeys(ctx context.Context, pattern string) (results []string, err error) {
	c := r.conn.WithContext(ctx)
	err = c.Keys(pattern).Err()
	if err != nil {
		return nil, err
	}
	results = c.Keys(pattern).Val()
	return results, nil
}

func (r *Redis) HGet(ctx context.Context, key string, field string) (result string, err error) {
	c := r.conn.WithContext(ctx)
	err = c.HGet(key, field).Err()
	if err != nil {
		return result, err
	}
	result = c.HGet(key, field).Val()
	return result, nil
}

func (r *Redis) Set(ctx context.Context, key string, data interface{}, exp time.Duration) (err error) {
	c := r.conn.WithContext(ctx)
	err = c.Set(key, data, exp).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) BulkSet(ctx context.Context, data map[string]interface{}, exp time.Duration) (err error) {
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
			return v.Err()
		}
	}
	return nil
}

func (r *Redis) HSet(ctx context.Context, key string, field string, data interface{}) (err error) {
	c := r.conn.WithContext(ctx)
	err = c.HSet(key, field, data).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) IsHealthy(ctx context.Context) (isHealthy bool) {
	return r.conn.WithContext(ctx).Ping().Err() == nil
}

func (r *Redis) ServiceName() string {
	return fmt.Sprintf(serviceName, r.address)
}
