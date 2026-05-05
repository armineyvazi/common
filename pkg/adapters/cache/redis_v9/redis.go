package redis_v9

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type Redis struct {
	address  string
	password string
	conn     *redis.Client
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
		conn:     client,
	}
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	res := r.conn.Get(ctx, key)
	if res.Err() != nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) GetAll(ctx context.Context, keys ...string) ([]string, error) {
	res := r.conn.MGet(ctx, keys...)
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
	res := r.conn.Keys(ctx, pattern)
	if res.Err() != nil {
		return nil, res.Err()
	}
	return res.Val(), nil
}

func (r *Redis) HGet(ctx context.Context, key string, field string) (string, error) {
	res, err := r.conn.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ports.ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return res, nil
}

func (r *Redis) Set(ctx context.Context, key string, data interface{}, exp time.Duration) error {
	res := r.conn.Set(ctx, key, data, exp)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (r *Redis) BulkSet(ctx context.Context, data map[string]interface{}, exp time.Duration) error {
	pipe := r.conn.Pipeline()
	for k, v := range data {
		pipe.Set(ctx, k, v, exp)
	}
	value, err := pipe.Exec(ctx)
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
	res := r.conn.HSet(ctx, key, field, data)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (r *Redis) HSetAll(ctx context.Context, key string, data map[string]any) error {
	err := r.conn.HMSet(ctx, key, data).Err()
	return err
}

func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	mapCmd := r.conn.HGetAll(ctx, key)
	res, err := mapCmd.Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func (r *Redis) HSetStruct(ctx context.Context, key string, data any) error {
	err := r.conn.HMSet(ctx, key, data).Err()
	return err
}

func (r *Redis) HGetStruct(ctx context.Context, key string, out any) error {
	err := r.conn.HGetAll(ctx, key).Scan(out)
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) ServiceName() string {
	return fmt.Sprintf("redis_%s", r.address)
}

func (r *Redis) IsHealthy(ctx context.Context) bool {
	return r.conn.Ping(ctx).Err() == nil
}

func (r *Redis) HIncrBy(ctx context.Context, key string, field string, incr int64) error {
	return r.conn.HIncrBy(ctx, key, field, incr).Err()
}

func (r *Redis) Del(ctx context.Context, key string) error {
	return r.conn.Del(ctx, key).Err()
}

func (r *Redis) SAdd(ctx context.Context, key string, members interface{}) error {
	return r.conn.SAdd(ctx, key, members).Err()
}
func (r *Redis) SCard(ctx context.Context, key string) (int64, error) {
	return r.conn.SCard(ctx, key).Result()
}

func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.conn.SIsMember(ctx, key, member).Result()
}

func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.conn.SMembers(ctx, key).Result()
}

func (r *Redis) SRem(ctx context.Context, key string, members interface{}) error {
	return r.conn.SRem(ctx, key, members).Err()
}
