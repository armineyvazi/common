# common

A shared Go infrastructure library providing production-ready adapters for caching, databases, messaging, HTTP serving, logging, and more. Designed for use across microservices, it defines clean port interfaces and ships concrete implementations behind them.

## Features

- **Cache** — Redis v9 (recommended) and Redis v6 with Elastic APM tracing
- **Databases** — MySQL (with APM tracing), PostgreSQL, ClickHouse via GORM
- **Messaging** — Kafka (Sarama, SCRAM-SHA-512) and RabbitMQ (AMQP 0-9-1)
- **HTTP Server** — Fiber v2 with JWT middleware, Sentry, APM, Swagger, pprof, health check
- **Logger** — Zap with trace ID propagation
- **In-memory cache** — TTLCache with eviction hooks
- **Worker pools** — native goroutine pool and Pond
- **Error utilities** — typed `AppError`, HTTP error mapping, validation helpers
- **Config** — Viper with environment variable overrides
- **Encoder** — Optimus integer obfuscation
- **UUID** — Google UUID v4

## Requirements

- Go 1.22+
- Dependencies auto-resolved via `go mod`

## Installation

```bash
go get github.com/armineyvazi/common.git
```

## Project Structure

```
pkg/
├── ports/          # Interfaces — the public contract
│   ├── cache.go
│   ├── database.go
│   ├── kafkaBroker.go
│   ├── messageBroker.go
│   ├── httpServer.go
│   ├── logger.go
│   └── ...
└── adapters/       # Concrete implementations
    ├── cache/
    │   ├── redis/          # go-redis v6 + APM (legacy)
    │   └── redis_v9/       # go-redis v9 (recommended)
    ├── database/
    │   ├── mysql/
    │   ├── postgres/
    │   └── clickhouse/
    ├── messageBroker/
    │   ├── kafka/
    │   └── rabbitmq/
    ├── httpServer/fiber/
    ├── logger/zap/
    ├── memory/ttlcache/
    ├── workerpool/
    ├── config/
    ├── encoder/optimus/
    └── errorUtil/
        ├── appErr/
        └── httpError/
```

## Quick Start

### Redis Cache

```go
import "github.com/armineyvazi/common.git/pkg/adapters/cache/redis_v9"

cache := redis_v9.New("localhost:6379", "", 0)

if err := cache.Set(ctx, "key", "value", time.Minute); err != nil {
    log.Fatal(err)
}
val, err := cache.Get(ctx, "key")
```

### PostgreSQL

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/postgres"

db := postgres.New("localhost", "mydb", "user", "pass", 5432, postgres.Config{
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
})

gormDB := db.GetConnection(ctx)
defer db.Close()
```

### MySQL

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/mysql"

db := mysql.New("localhost:3306", "mydb", "user", "pass", "charset=utf8mb4", mysql.Config{
    PrepareStmt: true,
})
gormDB := db.GetConnection(ctx)
defer db.Close()
```

### Kafka Producer

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/kafka"

producer := kafka.NewProducer(logger, "broker1:9092,broker2:9092", "user", "pass")
producer.Produce("my-topic", []byte(`{"event":"user_created"}`))
```

### Kafka Consumer

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/kafka"

consumer := kafka.NewConsumer(logger, "broker1:9092", "user", "pass")
msgs := make(chan []byte, 100)
if err := consumer.Consume(ctx, "my-topic", "my-group", msgs); err != nil {
    log.Fatal(err)
}
for msg := range msgs {
    // process msg
}
```

### RabbitMQ Producer

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/rabbitmq"

producer, err := rabbitmq.NewProducer("amqp://guest:guest@localhost:5672/")
if err != nil {
    log.Fatal(err)
}
defer producer.Close()

err = producer.Publish(ctx, map[string]any{"event": "order_placed"},
    "idempotency-key-123", "my-exchange", "order-queue", []string{"order.created"})
```

### RabbitMQ Consumer

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/rabbitmq"

consumer, err := rabbitmq.NewConsumer("amqp://guest:guest@localhost:5672/")
if err != nil {
    log.Fatal(err)
}
defer consumer.Close()

deliveries := make(chan ports.Delivery, 100)
if err := consumer.Consume(ctx, "order-queue", "my-exchange", []string{"order.created"}, deliveries); err != nil {
    log.Fatal(err)
}
for d := range deliveries {
    // process d.Body
    if err := consumer.Ack(d.Tag); err != nil {
        log.Println("ack failed:", err)
    }
}
```

### HTTP Server

```go
import (
    fiberserver "github.com/armineyvazi/common.git/pkg/adapters/httpServer/fiber"
    "github.com/armineyvazi/common.git/pkg/adapters/uuid"
    "github.com/armineyvazi/common.git/pkg/ports"
)

uuidGen := uuid.New()
server := fiberserver.NewFiberWithErrorModel(false, logger, ":8080", uuidGen, nil)

server.ActiveTraceID()
server.ActiveLogger()
server.ActiveRecover()
server.ActiveHealthCheck()

server.SetRouteGroups("v1/users", nil, []ports.Route{
    {Method: ports.Get, Path: "/", Handler: listUsers},
    {Method: ports.Post, Path: "/", Handler: createUser},
})

if err := server.Listen(); err != nil {
    log.Fatal(err)
}
```

### Logger

```go
import (
    zaplogger "github.com/armineyvazi/common.git/pkg/adapters/logger/zap"
    "github.com/armineyvazi/common.git/pkg/ports"
)

logger := zaplogger.New(ports.Info, nil)
logger.Info(ctx, "request completed", "method", "GET", "path", "/users")
logger.Errorw(ctx, "database error", "err", err, "table", "users")
```

### Error Handling

```go
import (
    "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
    "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/httpError"
)

// Create typed errors
err := appErr.NewNotFoundErr(errors.New("user not found")).
    WithTrackId(ctx).
    WithMessage("user does not exist")

// Map to HTTP response
httpErr := httpError.MapToHttpErr(err)
// httpErr.GetHttpStatus() == 404

// Check error type
if appErr.IsNotFoundError(err) {
    // handle not found
}
```

## Configuration

### Viper

```go
import "github.com/armineyvazi/common.git/pkg/adapters/config"

type AppConfig struct {
    Port     int    `mapstructure:"port"`
    LogLevel string `mapstructure:"log_level"`
}

var cfg AppConfig
if err := config.NewViper(&cfg, "config.yaml"); err != nil {
    log.Fatal(err)
}
```

Environment variables override config file values. A key `log_level` in the file is overridden by the `LOG_LEVEL` environment variable (case-insensitive, dots replaced with double underscores).

## Testing

Run unit tests:

```bash
go test ./...
go test -race ./...
go test -cover ./...
```

Run integration tests (requires running services):

```bash
REDIS_ADDR=localhost:6379 \
MYSQL_DSN="root:root@tcp(localhost:3306)/testdb?parseTime=True" \
POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=testdb port=5432 sslmode=disable" \
RABBITMQ_DSN="amqp://guest:guest@localhost:5672/" \
go test -tags integration ./...
```

Start services with Docker:

```bash
docker run -d -p 6379:6379 redis:7-alpine
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:16-alpine
docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root mysql:8
docker run -d -p 5672:5672 rabbitmq:3-alpine
```

## Linting

```bash
golangci-lint run
```

## Design Decisions

**Ports and adapters**: Each infrastructure concern is expressed as an interface in `pkg/ports`. Adapters implement these interfaces. Callers depend only on the interface, never on a specific library.

**Lazy connection**: Database adapters establish connections on the first `GetConnection` call, protected by `sync.Once` to prevent data races.

**Error model**: `AppError` carries type, code, message, track ID, and optional field-level validation errors. `httpError.MapToHttpErr` translates any `AppError` to the appropriate HTTP status code via a configurable map.

**Kafka consumer lifecycle**: The consumer goroutine exits when the context is cancelled or the AMQP channel closes. Call `Close()` to release the underlying connection.

## License

MIT
