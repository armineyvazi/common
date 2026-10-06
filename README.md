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
  - [PostgreSQL — GORM + AutoMigrate](#postgresql--gorm--automigrate)
  - [PostgreSQL — raw SQL (sqlc)](#postgresql--raw-sql-sqlc)
  - [MySQL — GORM + AutoMigrate](#mysql--gorm--automigrate)
  - [MySQL — raw SQL (sqlc)](#mysql--raw-sql-sqlc)
  - [SQL Migrations (versioned)](#sql-migrations-versioned)
  - [Confluent Kafka](#confluent-kafka)
  - [Kafka (Sarama)](#kafka-sarama)
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
- [Releasing](#releasing)
- [Design Decisions](#design-decisions)
- [Known Limitations](#known-limitations)
- [License](#license)

---

## Features

| Category                  | Adapter(s)                                                       |
|---------------------------|------------------------------------------------------------------|
| Cache                     | Redis v9 (recommended), Redis v6 + Elastic APM                  |
| Database — ORM            | PostgreSQL via GORM, MySQL via GORM, ClickHouse via GORM        |
| Database — raw SQL        | PostgreSQL via pgx/stdlib, MySQL via go-sql-driver               |
| Database — schema         | GORM `AutoMigrate` (struct-driven), versioned SQL via golang-migrate |
| Message broker            | Confluent Kafka (librdkafka, SASL/SSL), Kafka (Sarama, SCRAM-SHA-512), RabbitMQ (AMQP 0-9-1) |
| HTTP server               | Fiber v2 with JWT, Sentry, APM, Swagger, pprof, health check    |
| gRPC server               | google.golang.org/grpc with recovery + logging interceptors      |
| GraphQL handler           | graph-gophers/graphql-go (no code-gen, schema-first)             |
| JSON codec                | goccy/go-json (faster drop-in for encoding/json)                 |
| Logger                    | Zap with trace-ID propagation                                    |
| In-memory cache           | jellydator/ttlcache with generics + eviction hooks               |
| Worker pool               | Native goroutine pool, Pond                                      |
| Error utilities           | Typed `AppError`, HTTP error mapping, validation helpers         |
| Config                    | Viper with env-var overrides                                     |
| Encoder                   | Optimus integer obfuscation                                      |
| UUID                      | Google UUID v4                                                   |

---

## Requirements

- **Go 1.26** or later
- External services resolved at runtime (Redis, PostgreSQL, MySQL, etc.); no build-time infrastructure needed.
- **CGO enabled** (`CGO_ENABLED=1`) when using the Confluent Kafka adapter — it wraps librdkafka via CGO.

---

## Installation

```bash
go get github.com/armineyvazi/common.git
```

---

## Architecture

```
pkg/
├── ports/               # Interfaces — the stable public contract
│   ├── cache.go
│   ├── database.go      # Database, GORMMigrator
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
└── adapters/            # Concrete implementations
    ├── cache/
    │   ├── redis/           # go-redis v6 + Elastic APM (legacy)
    │   └── redis_v9/        # go-redis v9 (recommended)
    ├── config/              # Viper
    ├── database/
    │   ├── clickhouse/      # GORM + ClickHouse
    │   ├── migrate/         # Versioned SQL migrations (golang-migrate)
    │   ├── mysql/           # GORM + AutoMigrate + raw *sql.DB
    │   └── postgres/        # GORM + AutoMigrate + raw *sql.DB
    ├── encoder/optimus/
    ├── errorUtil/
    │   ├── appErr/          # AppError constructors and helpers
    │   └── httpError/       # HTTP status mapping and JSON marshalling
    ├── graphql/graphqlgo/   # graph-gophers/graphql-go handler
    ├── grpcServer/          # gRPC server with interceptors
    ├── httpServer/fiber/    # Fiber v2 HTTP server
    ├── json/goccy/          # goccy/go-json codec
    ├── logger/zap/
    ├── memory/ttlcache/     # TTL in-memory cache
    ├── messageBroker/
    │   ├── confluent/       # Confluent Kafka (librdkafka, SASL/SSL)
    │   ├── kafka/           # Sarama Kafka (SCRAM-SHA-512)
    │   └── rabbitmq/
    ├── uuid/
    └── workerpool/          # native goroutine pool + Pond
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

---

### PostgreSQL — GORM + AutoMigrate

`postgres.New` returns `ports.GORMMigrator`, a superset of `ports.Database` that adds `AutoMigrate`.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/postgres"

db := postgres.New("localhost", "mydb", "user", "pass", 5432, postgres.Config{
    MaxOpenConns:    25,
    MaxIdleConns:    10,
    ConnMaxLifetime: time.Hour,
})
defer func() { _ = db.Close() }()

// Define GORM model structs.
type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex;size:320"`
}
type Order struct {
    ID     uint `gorm:"primaryKey"`
    UserID uint `gorm:"index"`
    Total  int64
}

// AutoMigrate creates missing tables and adds missing columns.
// It never drops columns or tables — safe to call on every startup.
if err := db.AutoMigrate(&User{}, &Order{}); err != nil {
    log.Fatal(err)
}

gormDB := db.GetConnection(ctx) // *gorm.DB — pass to repositories
```

DSN string constructor (useful with secret managers):

```go
db := postgres.NewWithDSN(
    "host=localhost user=app password=secret dbname=mydb port=5432 sslmode=require",
    postgres.Config{PrepareStmt: true},
)
```

---

### PostgreSQL — raw SQL (sqlc)

Use this adapter when consuming `*sql.DB` directly or with sqlc-generated code.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/postgres"

sqlDB := postgres.NewSQL(postgres.SQLConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "user",
    Password: "pass",
    Database: "mydb",
    SSLMode:  "require", // default; set "disable" only in development
})
defer func() { _ = sqlDB.Close() }()

// Pass to sqlc-generated constructor:
//   queries := db.New(sqlDB.GetDB())
rawDB := sqlDB.GetDB() // *sql.DB
```

---

### MySQL — GORM + AutoMigrate

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/mysql"

db := mysql.New("localhost:3306", "mydb", "user", "pass", "charset=utf8mb4", mysql.Config{
    PrepareStmt: true,
})
defer func() { _ = db.Close() }()

type Product struct {
    ID    uint   `gorm:"primaryKey"`
    Title string `gorm:"size:500"`
    Price int64
}

if err := db.AutoMigrate(&Product{}); err != nil {
    log.Fatal(err)
}

gormDB := db.GetConnection(ctx)
```

---

### MySQL — raw SQL (sqlc)

```go
import "github.com/armineyvazi/common.git/pkg/adapters/database/mysql"

sqlDB := mysql.NewSQL(mysql.SQLConfig{
    Host:      "localhost:3306",
    Database:  "mydb",
    User:      "user",
    Password:  "pass",
    TLSConfig: "true", // default: verify using system CAs
})
defer func() { _ = sqlDB.Close() }()

rawDB := sqlDB.GetDB() // *sql.DB — pass to sqlc queries
```

---

### SQL Migrations (versioned)

`pkg/adapters/database/migrate` wraps [golang-migrate](https://github.com/golang-migrate/migrate) and accepts any `fs.FS`, making it natural to embed migration files into the binary.

**Migration file layout** (`migrations/` can live anywhere under your module):

```
migrations/
├── 0001_create_users.up.sql
├── 0001_create_users.down.sql
├── 0002_add_created_at.up.sql
└── 0002_add_created_at.down.sql
```

**Embedding and running migrations:**

```go
import (
    "database/sql"
    "embed"

    dbmigrate "github.com/armineyvazi/common.git/pkg/adapters/database/migrate"
    _ "github.com/jackc/pgx/v5/stdlib" // pgx driver
)

//go:embed migrations
var migrationsFS embed.FS

func runMigrations(db *sql.DB) error {
    m, err := dbmigrate.NewPostgres(db, migrationsFS, "migrations")
    if err != nil {
        return fmt.Errorf("create migrator: %w", err)
    }
    defer func() { _ = m.Close() }()

    // Up applies all pending migrations; returns nil if already at latest.
    return m.Up()
}
```

**MySQL:**

```go
m, err := dbmigrate.NewMySQL(db, migrationsFS, "migrations")
```

**Step-by-step control:**

```go
// Apply one migration forward.
if err := m.Steps(1); err != nil {
    log.Fatal(err)
}

// Roll back one migration.
if err := m.Steps(-1); err != nil {
    log.Fatal(err)
}

// Roll back all applied migrations.
if err := m.Down(); err != nil {
    log.Fatal(err)
}

// Check current schema version.
version, dirty, err := m.Version()

// Recover from a dirty state after manually fixing the database.
if dirty {
    if err := m.Force(int(version)); err != nil {
        log.Fatal(err)
    }
}
```

> **GORM AutoMigrate vs. versioned migrations** — `AutoMigrate` is fast and zero-config for development and simple projects (it reflects GORM models at startup). Versioned SQL migrations give you full control for complex schema changes, multi-step data migrations, and auditable rollback history. Use both: `AutoMigrate` in development, versioned files in production.

---

### Confluent Kafka

Requires `CGO_ENABLED=1` (wraps librdkafka). Supports SASL/PLAIN, SASL/SSL, and plain-text brokers.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/confluent"

// Producer
p, err := confluent.NewProducer(logger, confluent.ProducerConfig{
    BaseConfig: confluent.BaseConfig{
        Brokers:         "broker:9092",
        SecurityProtocol: "SASL_SSL",
        SaslMechanism:   "PLAIN",
        SaslUsername:    "user",
        SaslPassword:    "pass",
    },
})
if err != nil {
    log.Fatal(err)
}
defer p.Flush(5000)

if err := p.Produce("orders", []byte(`{"event":"order_created"}`)); err != nil {
    log.Fatal(err)
}

// Consumer
c, err := confluent.NewConsumer(logger, confluent.ConsumerConfig{
    BaseConfig: confluent.BaseConfig{Brokers: "broker:9092"},
    GroupID:    "order-service",
})
if err != nil {
    log.Fatal(err)
}
defer func() { _ = c.Close() }()

msgs := make(chan []byte, 100)
if err := c.Consume(ctx, "orders", "order-service", msgs); err != nil {
    log.Fatal(err)
}
for msg := range msgs {
    // unmarshal and process msg
}
```

---

### Kafka (Sarama)

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/kafka"

producer := kafka.NewProducer(logger, "broker1:9092,broker2:9092", "user", "pass")
producer.Produce("orders", []byte(`{"event":"order_created","id":1}`))

consumer := kafka.NewConsumer(logger, "broker1:9092", "user", "pass")
msgs := make(chan []byte, 100)
if err := consumer.Consume(ctx, "orders", "order-service", msgs); err != nil {
    log.Fatal(err)
}
for msg := range msgs {
    // process msg
}
```

---

### RabbitMQ

```go
import "github.com/armineyvazi/common.git/pkg/adapters/messageBroker/rabbitmq"

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

---

### gRPC Server

```go
import (
    grpcserver "github.com/armineyvazi/common.git/pkg/adapters/grpcServer"
    pb "your/module/gen/proto"
)

srv := grpcserver.New(logger, grpcserver.Config{})
srv.Register(&pb.YourService_ServiceDesc, &yourServiceImpl{})

go func() {
    if err := srv.Listen(":50051"); err != nil {
        log.Fatal(err)
    }
}()

srv.Shutdown(ctx) // graceful shutdown on signal
```

The adapter installs unary + stream panic recovery interceptors (returns `INTERNAL`) and structured request logging.

---

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

srv := graphqlgo.New(schema, &resolver{}, graphqlgo.Config{MaxDepth: 10})
http.Handle("/graphql", srv.Handler())
```

---

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
    {Method: ports.Get,  Path: "/",    Handler: listUsers},
    {Method: ports.Post, Path: "/",    Handler: createUser},
    {Method: ports.Get,  Path: "/:id", Handler: getUser},
})

if err := server.Listen(); err != nil {
    log.Fatal(err)
}
```

---

### JSON Codec

`goccy/go-json` is API-compatible with `encoding/json` and typically 2–3× faster.

```go
import "github.com/armineyvazi/common.git/pkg/adapters/json/goccy"

codec := goccy.New() // implements ports.JSONCodec

b, err := codec.Marshal(myStruct)
var out MyStruct
if err := codec.Unmarshal(b, &out); err != nil {
    return err
}

enc := codec.NewEncoder(w)
_ = enc.Encode(myStruct)

dec := codec.NewDecoder(r)
_ = dec.Decode(&out)
```

---

### Logger

```go
import (
    zaplogger "github.com/armineyvazi/common.git/pkg/adapters/logger/zap"
    "github.com/armineyvazi/common.git/pkg/ports"
)

logger := zaplogger.New(ports.Info, nil) // nil = no Sentry integration
logger.Info(ctx, "request completed", "method", "GET", "path", "/users", "latency_ms", 12)
logger.Error(ctx, "database error", "err", err, "table", "users")
```

Attach a trace ID to the context to include it in every log line:

```go
ctx = context.WithValue(ctx, ports.TraceID{}, "req-abc-123")
logger.Info(ctx, "handler called") // emits trace_id=req-abc-123
```

---

### In-Memory Cache

Generic TTL cache backed by `jellydator/ttlcache`.

```go
import (
    memcache "github.com/armineyvazi/common.git/pkg/adapters/memory/ttlcache"
    "github.com/jellydator/ttlcache/v3"
)

// Strongly typed: key=string, value=UserProfile
cache := memcache.New[string, UserProfile](
    ttlcache.WithTTL[string, UserProfile](5 * time.Minute),
)

cache.Set("user:42", profile, ttlcache.DefaultTTL)

if item := cache.Get("user:42"); item != nil {
    fmt.Println(item.Value())
}
```

---

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

---

### Error Handling

```go
import (
    "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
    "github.com/armineyvazi/common.git/pkg/adapters/errorUtil/httpError"
)

err := appErr.NewNotFoundErr(errors.New("user row missing")).
    WithTrackId(ctx).
    WithMessage("user does not exist").
    WithErrorList("id", "must be a positive integer")

he := httpError.MapToHttpErr(err)
c.Status(he.GetHttpStatus()).JSON(he) // {code, message, track_id, errors}

if appErr.IsNotFoundError(err) { ... }
if appErr.IsErrorType(err, ports.TypeDuplicate) { ... }
```

---

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

Environment variables override config-file values. Key `log_level` is overridden by `LOG_LEVEL`; nested key `database.host` is overridden by `DATABASE__HOST`.

---

## Error Model Reference

### Error types (`ports.ErrorType`)

| Constant               | Value                  |
|------------------------|------------------------|
| `TypeValidation`       | `"VALIDATION"`         |
| `TypeNotFound`         | `"NOT_FOUND"`          |
| `TypeDuplicate`        | `"DUPLICATED"`         |
| `TypeUnAuthorized`     | `"UNAUTHORIZED"`       |
| `TypeForbidden`        | `"FORBIDDEN"`          |
| `TypeInternal`         | `"INTERNAL"`           |
| `TypeTooManyRequests`  | `"TOO_MANY_REQUESTS"`  |
| `TypeUnprocessable`    | `"UNPROCESSABLE"`      |
| `TypeMethodNotAllowed` | `"METHOD_NOT_ALLOWED"` |

### Default messages (`ports.ErrorMessage`)

| Constant                  | Default message          |
|---------------------------|--------------------------|
| `InternalErrorMessage`    | `internal server error`  |
| `BadRequestMessage`       | `bad request`            |
| `NotFoundMessage`         | `not found`              |
| `DuplicatedMessage`       | `duplicate entry`        |
| `UnprocessableMessage`    | `unprocessable entity`   |
| `PermissionDeniedMessage` | `permission denied`      |
| `UnauthenticatedMessage`  | `unauthenticated`        |
| `UnauthorizedMessage`     | `unauthorized`           |
| `InvalidArgumentMessage`  | `invalid argument`       |
| `TypeTooManyMessage`      | `too many requests`      |

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
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/json/goccy/
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/errorUtil/appErr/
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/workerpool/
go test -bench=. -benchmem -benchtime=5s ./pkg/adapters/memory/ttlcache/
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

> Results vary by hardware. Always benchmark on the target environment.

---

## Linting

```bash
go vet ./...
golangci-lint run
```

CI uses golangci-lint v2 with the following linters enabled: `errcheck`, `govet`, `staticcheck`, `gocyclo`, `unparam`, `ineffassign`, `unused`, `bodyclose`, `noctx`, `gocritic`, `misspell`.

---

## Releasing

Releases are driven by annotated git tags. The `make tag` command validates semver, guards against a dirty working tree, creates the tag, and pushes it. GitHub Actions then runs the full CI gate and publishes a GitHub Release with the matching section from `CHANGELOG.md`.

```bash
# 1. Add your changes under [Unreleased] in CHANGELOG.md.
# 2. Rename [Unreleased] to [0.2.0] - YYYY-MM-DD and add a new empty [Unreleased] above it.
# 3. Commit:
git add CHANGELOG.md && git commit -m "chore(release): prepare v0.2.0"

# 4. Tag and push — this triggers the release workflow on GitHub Actions:
make tag VERSION=v0.2.0
```

Pre-releases use a suffix (`-beta.1`, `-rc.1`) and are automatically marked as pre-release on GitHub:

```bash
make tag VERSION=v0.2.0-rc.1
```

---

## Design Decisions

**Ports and adapters.** Each infrastructure concern is an interface in `pkg/ports`. Adapters implement the interface. Business logic depends only on the interface, never on a concrete library. Swapping Redis v6 for v9 requires changing one line at the composition root.

**`GORMMigrator` as a superset of `Database`.** Rather than adding `AutoMigrate` to the base `Database` interface (which would break every existing adapter), a new `GORMMigrator` interface embeds `Database` and adds `AutoMigrate(models ...any) error`. Code typed to `ports.Database` continues to compile; code that needs schema management widens the type to `ports.GORMMigrator`. Both postgres and mysql adapters satisfy both.

**GORM AutoMigrate vs. versioned migrations.** `AutoMigrate` reflects GORM model structs at startup: it creates missing tables and missing columns, but never drops anything. It is fast, zero-config, and safe to run on every startup. Versioned SQL migrations (`pkg/adapters/database/migrate`) provide full control: arbitrary SQL, explicit rollback, version tracking in a `schema_migrations` table. Use both in the same project if needed — `AutoMigrate` in development, versioned files in production CI.

**golang-migrate with `fs.FS`.** The `migrate` adapter uses the `iofs` source driver, which accepts any `fs.FS`. The intended use is `//go:embed migrations`, which embeds the SQL files into the binary at compile time — no external filesystem access at runtime, no deployment scripts needed. The underlying `*sql.DB` is not closed by `Close()`: the adapter does not own the connection, the caller does.

**TLS by default.** The raw SQL adapters default to `SSLMode: "require"` (PostgreSQL) and `TLSConfig: "true"` (MySQL). Disabling TLS requires an explicit config change, making insecure configurations visible in code review.

**Safe DSN construction.** PostgreSQL uses `net/url.URL` to build the DSN; MySQL uses `go-sql-driver/mysql`'s `Config.FormatDSN()`. Neither allows DSN injection through user-supplied configuration values.

**Lazy connection.** Database adapters establish the connection on the first `GetConnection` call, protected by `sync.Once`. This avoids failing at construction time when the service is not yet ready to connect.

**Error model.** `AppError` carries type, code, message, track ID, optional field-level validation errors, and extra context. `httpError.MapToHttpErr` translates any `AppError` to the correct HTTP status. The map can be extended at startup with `httpError.AddToMap`.

**gRPC interceptors.** The gRPC adapter installs panic recovery on both unary and stream handlers. A panic is caught, logged, and returned as a gRPC `INTERNAL` status — the server stays alive.

**GraphQL at startup.** `graphqlgo.New` calls `graphql.MustParseSchema`, which panics on an invalid schema. A malformed schema is a programming error, not a runtime error, so it must surface at process start rather than on the first request.

**JSON codec.** `goccy/go-json` is a pure-Go, API-compatible replacement for `encoding/json` that is typically 2–3× faster. Exposed through `ports.JSONCodec` so callers can substitute the standard library in tests without changing call sites.

---

## Known Limitations

- **No connection pool tuning for raw SQL adapters.** `postgres.NewSQL` and `mysql.NewSQL` use Go's default pool settings. Applications with high concurrency should configure `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` on the returned `*sql.DB`.
- **Kafka consumer (Sarama) does not support consumer-group rebalancing.** The current implementation subscribes to a single partition. Production use cases with multiple consumer instances require a group rebalance strategy; consider the Confluent adapter for that.
- **HTTP/2 not enabled on the Fiber server.** Fiber uses fasthttp, which does not support HTTP/2.
- **No distributed tracing out of the box.** The logger propagates a `TraceID` context value; APM integration is present in the Fiber and Redis v6 adapters but not in the gRPC or GraphQL adapters.
- **Confluent Kafka requires CGO.** `confluent-kafka-go/v2` wraps librdkafka via CGO. Pure-Go environments (e.g., scratch containers) must use the Sarama adapter instead.

---

## License

MIT
