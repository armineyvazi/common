package ports

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type Database interface {
	GetConnection(ctx context.Context) *gorm.DB
	Close() error
}

// GORMMigrator is implemented by GORM adapters that support GORM's AutoMigrate.
// Use it for development and staging; prefer versioned SQL files in production.
type GORMMigrator interface {
	Database
	// AutoMigrate creates missing tables and adds missing columns for the given
	// model types. It is idempotent and never drops columns or tables.
	AutoMigrate(models ...any) error
}

type MongoDatabase interface {
	GetConnection(ctx context.Context) *mongo.Client
	Close() error
}

// SQLDatabase provides a *sql.DB connection suitable for use with
// sqlc-generated query functions and any code that prefers database/sql
// over an ORM.
type SQLDatabase interface {
	GetDB() *sql.DB
	Ping(ctx context.Context) error
	Close() error
}
