# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.2.0] - 2026-10-24

### Changed (breaking)
- **Module path is now `github.com/armineyvazi/common`** (was
  `github.com/armineyvazi/common.git`). The go command rejects a `.git`
  suffix on github.com paths ("invalid version control suffix"), so v0.1.0
  could not be fetched with `go get`. Update imports:
  `sed -i 's#armineyvazi/common.git#armineyvazi/common#g'` on your sources.
- `ports.LoggerWithTraceID` gains `Debugw` and `Infow`. Custom
  implementations of the interface must add them.

### Added
- `Infow` / `Debugw` on the zap adapter: structured key-value logging at
  info and debug level. Previously only `Warnw` / `Errorw` existed, and
  key-value pairs passed to `Info` were rendered as `%!(EXTRA …)`.
- Functional options for all Fiber constructors: `WithTimeouts`,
  `WithBodyLimit`, `WithoutStartupMessage`, `WithFiberConfig`. Existing
  call sites compile unchanged.

### Fixed
- `NewFiberWithErrorModel` panicked on the first failed request when the
  error handler was `nil`, as shown in the README. A `nil` handler now
  disables error reporting.
- README logger example used printf-style `Info` with key-value pairs.

---

## [0.1.0] - 2026-10-06

### Added
- Adapter packages: config, cache/redis_v9, database/postgres, database/mysql,
  encoder/optimus, errorUtil/appErr, errorUtil/httpError, graphql/graphqlgo,
  grpcServer, httpServer/fiber, json/goccy, logger/zap, memory/ttlcache,
  messageBroker/confluent, messageBroker/kafka, messageBroker/rabbitmq, uuid, workerpool
- `ports` package with core interfaces: Database, GORMMigrator, Cache, Logger, MessageBroker, …
- `ports.GORMMigrator` interface — superset of `ports.Database` exposing `AutoMigrate`
- GORM `AutoMigrate` on postgres and mysql adapters (idempotent; never drops columns)
- `pkg/adapters/database/migrate` — versioned SQL migrations via `golang-migrate/migrate/v4`
  - `NewPostgres` / `NewMySQL` — driver-agnostic constructors accepting `*sql.DB`
  - `Up`, `Down`, `Steps`, `Version`, `Force`, `Close` — full migration lifecycle
  - Embed-friendly: accepts any `fs.FS` (use `//go:embed` for zero-config deploys)
- Runnable `Example*` functions for every adapter package
- GitHub Actions CI: unit tests, race detector, coverage, integration tests, golangci-lint
- GitHub Actions release workflow: triggered by `vX.Y.Z` tags, gates on CI, publishes GitHub Release
- Makefile with `test`, `test-race`, `lint`, `build`, `tidy`, `tag` targets

### Changed
- `postgres.New` / `postgres.NewWithDSN` return `ports.GORMMigrator` (backward-compatible)
- `mysql.New` returns `ports.GORMMigrator` (backward-compatible)

---

[Unreleased]: https://github.com/armineyvazi/common/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/armineyvazi/common/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/armineyvazi/common/releases/tag/v0.1.0
