// Package postgres provides a PostgreSQL database adapter backed by GORM.
package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Config holds connection pool settings for the PostgreSQL adapter.
type Config struct {
	PrepareStmt     bool
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type postgresDB struct {
	dsn        string
	dbConnOnce sync.Once
	db         *gorm.DB
	config     Config
}

// New creates a new PostgreSQL adapter. The connection is established lazily
// on the first call to GetConnection.
// The returned value satisfies both ports.Database and ports.GORMMigrator.
func New(host, database, user, password string, port int, config Config) ports.GORMMigrator {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		host, user, password, database, port,
	)
	return &postgresDB{dsn: dsn, config: config}
}

// NewWithDSN creates a new PostgreSQL adapter from a full DSN string.
// The returned value satisfies both ports.Database and ports.GORMMigrator.
func NewWithDSN(dsn string, config Config) ports.GORMMigrator {
	return &postgresDB{dsn: dsn, config: config}
}

func (p *postgresDB) GetConnection(ctx context.Context) *gorm.DB {
	p.dbConnOnce.Do(func() {
		var err error
		p.db, err = gorm.Open(postgres.Open(p.dsn), &gorm.Config{
			PrepareStmt: p.config.PrepareStmt,
		})
		if err != nil {
			panic(fmt.Errorf("open postgres: %w", err))
		}

		db, err := p.db.DB()
		if err != nil {
			panic(fmt.Errorf("get postgres db: %w", err))
		}

		if p.config.MaxIdleConns > 0 {
			db.SetMaxIdleConns(p.config.MaxIdleConns)
		}
		if p.config.MaxOpenConns > 0 {
			db.SetMaxOpenConns(p.config.MaxOpenConns)
		}
		if p.config.ConnMaxLifetime > 0 {
			db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
		}
		if p.config.ConnMaxIdleTime > 0 {
			db.SetConnMaxIdleTime(p.config.ConnMaxIdleTime)
		}
	})
	return p.db.WithContext(ctx)
}

// AutoMigrate implements ports.GORMMigrator. It creates missing tables and
// adds missing columns for the provided model types. It never drops columns
// or tables. Prefer versioned SQL migrations in production.
func (p *postgresDB) AutoMigrate(models ...any) error {
	return p.GetConnection(context.Background()).AutoMigrate(models...)
}

func (p *postgresDB) Close() error {
	if p.db == nil {
		return nil
	}
	db, err := p.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
