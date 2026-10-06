package redis_v9_test

import (
	"context"
	"fmt"
	"time"

	redis "github.com/armineyvazi/common.git/pkg/adapters/cache/redis_v9"
)

// ExampleNew_basicGetSet shows storing and retrieving a plain string value.
func ExampleNew_basicGetSet() {
	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	if err := cache.Set(ctx, "user:42:name", "Alice", time.Minute); err != nil {
		fmt.Println("set:", err)
		return
	}

	val, err := cache.Get(ctx, "user:42:name")
	if err != nil {
		fmt.Println("get:", err)
		return
	}
	fmt.Println(val)
}

// ExampleNew_hashStruct shows storing a struct as a Redis hash field-by-field
// and reading it back into a typed struct. Useful for session or profile data.
func ExampleNew_hashStruct() {
	type Session struct {
		UserID int    `redis:"user_id"`
		Role   string `redis:"role"`
	}

	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	sess := Session{UserID: 7, Role: "admin"}
	if err := cache.HSetStruct(ctx, "session:tok-abc", sess); err != nil {
		fmt.Println("hset struct:", err)
		return
	}

	var out Session
	if err := cache.HGetStruct(ctx, "session:tok-abc", &out); err != nil {
		fmt.Println("hget struct:", err)
		return
	}
	fmt.Printf("user=%d role=%s\n", out.UserID, out.Role)
}

// ExampleNew_hashFields shows reading and writing individual hash fields.
func ExampleNew_hashFields() {
	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	_ = cache.HSet(ctx, "product:10", "stock", "250")
	_ = cache.HIncrBy(ctx, "product:10", "stock", -5)

	stock, _ := cache.HGet(ctx, "product:10", "stock")
	fmt.Println("stock:", stock)
}

// ExampleNew_set shows managing a Redis set, useful for tracking unique items
// such as online users or processed event IDs.
func ExampleNew_set() {
	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	_ = cache.SAdd(ctx, "online:users", "user:1")
	_ = cache.SAdd(ctx, "online:users", "user:2")

	count, _ := cache.SCard(ctx, "online:users")
	fmt.Println("online:", count)

	isMember, _ := cache.SIsMember(ctx, "online:users", "user:1")
	fmt.Println("user:1 online:", isMember)

	_ = cache.SRem(ctx, "online:users", "user:1")
}

// ExampleNew_bulkSet shows writing multiple keys in one round trip using
// Redis MSET, which is significantly more efficient than individual SETs.
func ExampleNew_bulkSet() {
	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	batch := map[string]any{
		"config:max_retries": "3",
		"config:timeout_ms":  "5000",
		"config:debug":       "false",
	}
	if err := cache.BulkSet(ctx, batch, 10*time.Minute); err != nil {
		fmt.Println("bulk set:", err)
		return
	}
	fmt.Println("bulk write ok")
}

// ExampleNew_searchKeys shows listing keys matching a glob pattern.
func ExampleNew_searchKeys() {
	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	_ = cache.Set(ctx, "session:a", "1", time.Hour)
	_ = cache.Set(ctx, "session:b", "2", time.Hour)

	keys, err := cache.SearchKeys(ctx, "session:*")
	if err != nil {
		fmt.Println("search:", err)
		return
	}
	fmt.Println("session keys found:", len(keys) >= 2)
}

// ExampleNew_delete shows removing a key and verifying its absence.
func ExampleNew_delete() {
	cache := redis.New("localhost:6379", "", 0)
	ctx := context.Background()

	_ = cache.Set(ctx, "temp:flag", "1", time.Hour)
	_ = cache.Del(ctx, "temp:flag")

	_, err := cache.Get(ctx, "temp:flag")
	fmt.Println("deleted:", err != nil)
}
