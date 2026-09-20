# common

A shared Go infrastructure library that provides production-ready adapters for caching, databases, messaging, HTTP serving, logging, and more. It follows the **ports-and-adapters** (hexagonal) pattern: `pkg/ports` defines stable interfaces; `pkg/adapters` provides concrete implementations that callers swap freely.

---

## Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
  - [Redis Cache](#redis-cache)
  - [PostgreSQL (GORM)](#postgresql-gorm)
  - [PostgreSQL (raw SQL)](#postgresql-raw-sql)
  - [MySQL (GORM)](#mysql-gorm)
  - [MySQL (raw SQL)](#mysql-raw-sql)
  - [Kafka](#kafka)
  - [RabbitMQ](#rabbitmq)
  - [gRPC Server](#grpc-server)
  - [GraphQL Handler](#graphql-handler)
  - [HTTP Server](#http-server)
  - [JSON Codec](#json-codec)
  - [Logger](#logger)
  - [In-Memory Cache](#in-memory-cache)
  - [Worker Pool](#worker-pool)
  - [Error Handling](#error-handling)
  - [Config](#config)
- [Error Model Reference](#error-model-reference)
- [Testing](#testing)
- [Benchmarks](#benchmarks)
- [Linting](#linting)
- [Design Decisions](#design-decisions)
- [Known Limitations](#known-limitations)
- [License](#license)

---

## Features

| Category           | Adapter(s)                                                  |
|--------------------|-------------------------------------------------------------|
| Cache              | Redis v9 (recommended), Redis v6 + Elastic APM             |
| Database (ORM)     | PostgreSQL via GORM, MySQL via GORM, ClickHouse via GORM   |
| Database (raw SQL) | PostgreSQL via pgx/stdlib, MySQL via go-sql-driver          |
| Message broker     | Kafka (Sarama, SCRAM-SHA-512), RabbitMQ (AMQP 0-9-1)      |
| HTTP server        | Fiber v2 with JWT, Sentry, APM, Swagger, pprof, health     |
| gRPC server        | google.golang.org/grpc with recovery + logging interceptors |
| GraphQL handler    | graph-gophers/graphql-go (no code-gen, schema-first)        |
| JSON codec         | goccy/go-json (faster drop-in for encoding/json)            |
| Logger             | Zap with trace-ID propagation                               |
| In-memory cache    | jellydator/ttlcache with generics + eviction hooks          |
| Worker pool        | Native goroutine pool, Pond                                 |
| Error utilities    | Typed `AppError`, HTTP error mapping, validation helpers    |
| Config             | Viper with env-var overrides                                |
| Encoder            | Optimus integer obfuscation                                 |
| UUID               | Google UUID v4                                              |

---

## Requirements

- **Go 1.22** or later
- External services resolved at runtime (Redis, PostgreSQL, MySQL, etc.); no build-time infrastructure needed.

---

## Installation

```bash
go get github.com/armineyvazi/common.git
```

---

## Architecture

```
pkg/
├── ports/           # Interfaces — the stable public contract
│   ├── cache.go
│   ├── database.go
│   ├── errorModel.go
│   ├── graphql.go
│   ├── grpc.go
│   ├── json.go
│   ├── kafkaBroker.go
│   ├── messageBroker.go
│   ├── httpServer.go
│   ├── logger.go
│   ├── memory.go
│   ├── workerpool.go
│   └── ...
└── adapters/        # Concrete implementations
    ├── cache/
    │   ├── redis/          # go-redis v6 + Elastic APM (legacy)
    │   └── redis_v9/       # go-redis v9 (recommended)
    ├── config/             # Viper
    ├── database/
    │   ├── clickhouse/     # GORM + ClickHouse
    │   ├── mysql/          # GORM  +  raw *sql.DB
    │   └── postgres/       # GORM  +  raw *sql.DB
    ├── encoder/optimus/
    ├── errorUtil/
    │   ├── appErr/         # AppError constructors and helpers
    │   └── httpError/      # HTTP status mapping and JSON marshalling
    ├── graphql/graphqlgo/  # graph-gophers/graphql-go handler
    ├── grpcServer/         # gRPC server with interceptors
    ├── httpServer/fiber/   # Fiber v2 HTTP server
    ├── json/goccy/         # goccy/go-json codec
    ├── logger/zap/
    ├── memory/ttlcache/    # TTL in-memory cache
    ├── messageBroker/
    │   ├── kafka/
    │   └── rabbitmq/
    ├── uuid/
    └── workerpool/         # native goroutine pool + Pond
```

### Dependency rule

Callers import only `pkg/ports`. Adapters are wired at the application's composition root (typically `main.go`). This keeps business logic free of infrastructure concerns and makes adapters interchangeable without touching domain code.

---

## Quick Start

### Redis Cache

```go
import "github.com/armineyvazi/common.git/pkg/adapters/cache/redis_v9"

cache := redis_v9.New("localhost:6379", "", 0)

if err := cache.Set(ctx, "session:42", "token-value", time.Hour); err != nil {
    log.Fatal(err)
}
val, err := cache.Get(ctx, "session:42")
```

### PostgreSQL (GORM)

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

### PostgreSQL (raw SQL)

Use this adapter when consuming `*sql.DB` directly or with sqlc-generated code.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/postgres"

sqlDB := postgres.NewSQL(postgres.SQLConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "user",
    Password: "pass",
    Database: "mydb",
    SSLMode:  "require",  // default; set "disable" only in development
})

db, err := sqlDB.GetDB(ctx)
if err != nil {
    log.Fatal(err)
}
// pass db to sqlc-generated queries or use directly
```

### MySQL (GORM)

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/mysql"

db := mysql.New("localhost:3306", "mydb", "user", "pass", "charset=utf8mb4", mysql.Config{
    PrepareStmt: true,
})
gormDB := db.GetConnection(ctx)
defer db.Close()
```

### MySQL (raw SQL)

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/mysql"

sqlDB := mysql.NewSQL(mysql.SQLConfig{
    Host:      "localhost:3306",
    Database:  "mydb",
    User:      "user",
    Password:  "pass",
    TLSConfig: "true",  // default: verify using system CAs
})

db, err := sqlDB.GetDB(ctx)
if err != nil {
    log.Fatal(err)
}
```

### Kafka

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/kafka"

// Producer
producer := kafka.NewProducer(logger, "broker1:9092,broker2:9092", "user", "pass")
producer.Produce("orders", []byte(`{"event":"order_created","id":1}`))

// Consumer
consumer := kafka.NewConsumer(logger, "broker1:9092", "user", "pass")
msgs := make(chan []byte, 100)
if err := consumer.Consume(ctx, "orders", "order-service", msgs); err != nil {
    log.Fatal(err)
}
for msg := range msgs {
    // process msg
}
```

### RabbitMQ

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/rabbitmq"

// Producer
producer, err := rabbitmq.NewProducer("amqp://guest:guest@localhost:5672/")
if err != nil {
    log.Fatal(err)
}
defer producer.Close()

err = producer.Publish(ctx,
    map[string]any{"event": "order_placed"},
    "idempotency-key-123",
    "orders-exchange", "order-queue", []string{"order.created"},
)

// Consumer
consumer, err := rabbitmq.NewConsumer("amqp://guest:guest@localhost:5672/")
if err != nil {
    log.Fatal(err)
}
defer consumer.Close()

deliveries := make(chan ports.Delivery, 100)
if err := consumer.Consume(ctx, "order-queue", "orders-exchange",
    []string{"order.created"}, deliveries); err != nil {
    log.Fatal(err)
}
for d := range deliveries {
    // process d.Body
    if err := consumer.Ack(d.Tag); err != nil {
        log.Println("ack failed:", err)
    }
}
```

### gRPC Server

```go
import (
    grpcserver "github.com/armineyvazi/common.git/pkg/adapters/grpcServer"
    pb "your/module/gen/proto"
)

srv := grpcserver.New(logger, grpcserver.Config{})

// Register any number of services before Listen.
srv.Register(&pb.YourService_ServiceDesc, &yourServiceImpl{})

// Listen blocks until Shutdown is called.
go func() {
    if err := srv.Listen(":50051"); err != nil {
        log.Fatal(err)
    }
}()

// Graceful shutdown on signal:
srv.Shutdown(ctx)
```

The adapter installs:
- **Unary interceptors**: panic recovery (returns INTERNAL), structured request logging
- **Stream interceptor**: panic recovery

### GraphQL Handler

No code generation required. Pass the SDL schema string and a resolver struct.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/graphql/graphqlgo"

const schema = `
    type Query {
        user(id: ID!): User
    }
    type User {
        id: ID!
        name: String!
    }
`

type resolver struct{}

func (r *resolver) User(args struct{ ID graphql.ID }) *userResolver {
    return &userResolver{id: string(args.ID)}
}

srv := graphqlgo.New(schema, &resolver{}, graphqlgo.Config{
    MaxDepth: 10,   // prevent deeply nested abuse
    Pretty:   false,
})

http.Handle("/graphql", srv.Handler())
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
    {Method: ports.Get,  Path: "/",   Handler: listUsers},
    {Method: ports.Post, Path: "/",   Handler: createUser},
    {Method: ports.Get,  Path: "/:id", Handler: getUser},
})

if err := server.Listen(); err != nil {
    log.Fatal(err)
}
```

### JSON Codec

`goccy/go-json` is API-compatible with `encoding/json` and typically 2–3× faster.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/json/goccy"

codec := goccy.New() // implements ports.JSONCodec

b, err := codec.Marshal(myStruct)
if err != nil {
    return err
}

var out MyStruct
if err := codec.Unmarshal(b, &out); err != nil {
    return err
}

// Streaming:
enc := codec.NewEncoder(w)
_ = enc.Encode(myStruct)

dec := codec.NewDecoder(r)
_ = dec.Decode(&out)
```

### Logger

```go
import (
    zaplogger "github.com/armineyvazi/common.git/pkg/adapters/logger/zap"
    "github.com/armineyvazi/common.git/pkg/ports"
)

logger := zaplogger.New(ports.Info, nil)
logger.Info(ctx, "request completed", "method", "GET", "path", "/users", "latency_ms", 12)
logger.Error(ctx, "database error", "err", err, "table", "users")
```

### In-Memory Cache

Generic TTL cache backed by `jellydator/ttlcache`.

```go
import (
    memcache "github.com/armineyvazi/common.git/pkg/adapters/memory/ttlcache"
    "github.com/jellydator/ttlcache/v3"
)

// Strongly-typed: key=string, value=UserProfile
cache := memcache.New[string, UserProfile](
    ttlcache.WithTTL[string, UserProfile](5 * time.Minute),
)

cache.Set("user:42", profile, ttlcache.DefaultTTL)

if item := cache.Get("user:42"); item != nil {
    fmt.Println(item.Value())
}
```

### Worker Pool

```go
import workerpool "github.com/armineyvazi/common.git/pkg/adapters/workerpool"

pool := workerpool.New(logger)

pool.RegisterTask("resize-image",
    func(req ports.JobRequest) ports.TaskResult {
        path, _ := req.Params.(string)
        // ... process ...
        return ports.TaskResult{Key: path}
    },
    /* concurrency */ 4,
    /* queue depth */ 1000,
)

pool.Run()

results := make(chan ports.TaskResult, 1)
pool.PushJob(ctx, "resize-image", results, "/uploads/photo.jpg")
r := <-results
```

### Error Handling

```go
import (
    "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
    "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/httpError"
)

// Create a typed, traceable error
err := appErr.NewNotFoundErr(errors.New("user row missing")).
    WithTrackId(ctx).
    WithMessage("user does not exist").
    WithErrorList("id", "must be a positive integer")

// Translate to HTTP response
he := httpError.MapToHttpErr(err)
c.Status(he.GetHttpStatus()).JSON(he) // he marshals to {code, message, track_id, errors}

// Predicate helpers
if appErr.IsNotFoundError(err) { ... }
if appErr.IsErrorType(err, ports.TypeDuplicate) { ... }

// Type-check the domain type directly
if ae, ok := err.(ports.AppError); ok {
    if ae.IsType(ports.TypeForbidden) { ... }
}
```

### Config

```go
import "github.com/armineyvazi/common.git/pkg/adapters/config"

type AppConfig struct {
    Port     int    `mapstructure:"port"`
    LogLevel string `mapstructure:"log_level"`
    DSN      string `mapstructure:"dsn"`
}

var cfg AppConfig
if err := config.NewViper(&cfg, "config.yaml"); err != nil {
    log.Fatal(err)
}
```

Environment variables override config-file values. A key `log_level` is overridden by the `LOG_LEVEL` environment variable (case-insensitive; dots become double underscores).

---

## Error Model Reference

### Error types (`ports.ErrorType`)

| Constant              | Value               |
|-----------------------|---------------------|
| `TypeValidation`      | `"VALIDATION"`      |
| `TypeNotFound`        | `"NOT_FOUND"`       |
| `TypeDuplicate`       | `"DUPLICATED"`      |
| `TypeUnAuthorized`    | `"UNAUTHORIZED"`    |
| `TypeForbidden`       | `"FORBIDDEN"`       |
| `TypeInternal`        | `"INTERNAL"`        |
| `TypeTooManyRequests` | `"TOO_MANY_REQUESTS"` |
| `TypeUnprocessable`   | `"UNPROCESSABLE"`   |
| `TypeMethodNotAllowed`| `"METHOD_NOT_ALLOWED"` |

### Default messages (`ports.ErrorMessage`)

| Constant                | Default message        |
|-------------------------|------------------------|
| `InternalErrorMessage`  | `internal server error`|
| `BadRequestMessage`     | `bad request`          |
| `NotFoundMessage`       | `not found`            |
| `DuplicatedMessage`     | `duplicate entry`      |
| `UnprocessableMessage`  | `unprocessable entity` |
| `PermissionDeniedMessage`| `permission denied`   |
| `UnauthenticatedMessage`| `unauthenticated`      |
| `UnauthorizedMessage`   | `unauthorized`         |
| `InvalidArgumentMessage`| `invalid argument`     |
| `TypeTooManyMessage`    | `too many requests`    |

### JSON response shape

```json
{
  "code": 10003,
  "message": "user does not exist",
  "track_id": "8427394",
  "status": false,
  "errors": [
    { "field_name": "id", "errors": ["must be a positive integer"] }
  ]
}
```

---

## Testing

Unit tests (no external services):

```bash
go test ./...
go test -race ./...
go test -cover ./...
```

Integration tests (requires running services):

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

---

## Benchmarks

Run all benchmarks with allocation reporting:

```bash
go test -bench=. -benchmem ./...
```

Run a specific package:

```bash
# JSON codec vs stdlib
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/json/goccy/

# Error creation and manipulation
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/errorUtil/appErr/

# Worker pool throughput
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/workerpool/

# In-memory cache
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/memory/ttlcache/

# UUID generation
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/uuid/
```

Example results (Apple M-series, Go 1.26, `-benchtime=5s`):

```
BenchmarkGoccy_Marshal-10          9838456    606 ns/op    256 B/op    4 allocs/op
BenchmarkStdlib_Marshal-10         3917862   1530 ns/op    256 B/op    4 allocs/op
BenchmarkGoccy_Unmarshal-10        5254432   1141 ns/op    432 B/op   11 allocs/op
BenchmarkStdlib_Unmarshal-10       2308659   2584 ns/op    544 B/op   14 allocs/op
BenchmarkNewInternalErr-10        15204448    394 ns/op    384 B/op    7 allocs/op
BenchmarkWorkerPool_Throughput-10    518136  11580 ns/op    112 B/op    4 allocs/op
BenchmarkGenV4-10                 10847218    553 ns/op     64 B/op    2 allocs/op
```

> Results vary by hardware. Always run benchmarks on the target environment.

Compare JSON codec with `-benchmem` to see the allocation advantage of `goccy/go-json` over the standard library for hot paths in your service.

---

## Linting

```bash
go vet ./...
golangci-lint run
```

CI uses golangci-lint v2 with the following linters enabled: `errcheck`, `govet`, `staticcheck`, `gocyclo`, `unparam`, `ineffassign`, `unused`, `bodyclose`, `noctx`, `gocritic`, `misspell`.

---

## Design Decisions

**Ports and adapters.** Each infrastructure concern is an interface in `pkg/ports`. Adapters implement the interface. Business logic depends only on the interface, never on a concrete library. Swapping Redis v6 for v9 (or replacing Sarama with Confluent Kafka) requires changing one line at the composition root.

**TLS by default.** The raw SQL adapters default to `SSLMode: "require"` (PostgreSQL) and `TLSConfig: "true"` (MySQL). Developers who need to disable TLS in local environments must do so explicitly, making insecure configurations visible in code review.

**Safe DSN construction.** PostgreSQL uses `net/url.URL` to build the DSN; MySQL uses `go-sql-driver/mysql`'s `Config.FormatDSN()`. Neither approach allows DSN injection through user-supplied configuration values.

**Lazy connection.** Database adapters establish the connection on the first `GetConnection` call, protected by `sync.Once`. This avoids failing at construction time when the application is not ready to connect, and prevents data races during concurrent initialisation.

**Error model.** `AppError` carries type, code, message, track ID, optional field-level validation errors, and extra context. `httpError.MapToHttpErr` translates any `AppError` to the correct HTTP status via a configurable map. The map can be extended at startup with `httpError.AddToMap`.

**gRPC interceptors.** The gRPC adapter installs panic recovery on both unary and stream handlers. A panic is caught, logged, and turned into a gRPC `INTERNAL` status — the server process stays alive.

**GraphQL at startup.** `graphqlgo.New` calls `graphql.MustParseSchema`, which panics on an invalid schema. This is intentional: a malformed schema is a programming error, not a runtime error, so it must surface at process start rather than on the first request.

**JSON codec.** `goccy/go-json` is a pure-Go, API-compatible replacement for `encoding/json` that is typically 2–3× faster on marshal and unmarshal paths. It is exposed through `ports.JSONCodec` so callers can substitute the standard library in tests or switch to a future implementation without changing call sites.

---

## Known Limitations

- **No connection pool tuning for raw SQL adapters.** `postgres.NewSQL` and `mysql.NewSQL` open a `*sql.DB` with Go's default pool settings. Applications with high concurrency should configure `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` on the returned `*sql.DB`.
- **Kafka consumer does not support consumer groups with rebalancing.** The current implementation subscribes to a single partition. Production use cases with multiple consumer instances require a group rebalance strategy.
- **HTTP/2 not enabled on the Fiber server.** Fiber uses fasthttp, which does not support HTTP/2 at the protocol level.
- **No distributed tracing out of the box.** The logger propagates a `TraceID` context value; APM integration is present in the Fiber and Redis v6 adapters but not in the gRPC or GraphQL adapters.

---

## License

MIT
