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
