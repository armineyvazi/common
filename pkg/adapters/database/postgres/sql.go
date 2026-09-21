package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver for database/sql

	"github.com/armineyvazi/common.git/pkg/ports"
)

// SQLConfig configures the *sql.DB connection pool for the raw SQL adapter.
type SQLConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	// SSLMode is the PostgreSQL sslmode parameter.
	// Accepted values: "disable", "allow", "prefer", "require",
	// "verify-ca", "verify-full". Defaults to "require" when empty.
	// Set to "disable" only for local development; never in production.
	SSLMode string
}

type sqlDatabase struct {
	dsn     string
	cfg     SQLConfig
	once    sync.Once
	db      *sql.DB
	initErr error
}

// NewSQL returns a ports.SQLDatabase backed by pgx via database/sql.
// The DSN must be a valid PostgreSQL connection string or URL, e.g.
//
//	"postgres://user:secret@host:5432/dbname?sslmode=require"
func NewSQL(dsn string, cfg SQLConfig) ports.SQLDatabase {
	return &sqlDatabase{dsn: dsn, cfg: cfg}
}

// NewSQLFromParts constructs a DSN from individual parameters using URL
// encoding so that special characters in user, password, or host are
// handled safely. Prefer NewSQL when you already manage DSN strings.
func NewSQLFromParts(host, database, user, password string, port int, cfg SQLConfig) ports.SQLDatabase {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/" + database,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()
	return NewSQL(u.String(), cfg)
}

func (s *sqlDatabase) connect() {
	s.once.Do(func() {
		db, err := sql.Open("pgx", s.dsn)
		if err != nil {
			s.initErr = fmt.Errorf("open postgres sql.DB: %w", err)
			return
		}
		if s.cfg.MaxIdleConns > 0 {
			db.SetMaxIdleConns(s.cfg.MaxIdleConns)
		}
		if s.cfg.MaxOpenConns > 0 {
			db.SetMaxOpenConns(s.cfg.MaxOpenConns)
		}
		if s.cfg.ConnMaxLifetime > 0 {
			db.SetConnMaxLifetime(s.cfg.ConnMaxLifetime)
		}
		if s.cfg.ConnMaxIdleTime > 0 {
			db.SetConnMaxIdleTime(s.cfg.ConnMaxIdleTime)
		}
		s.db = db
	})
}

// GetDB returns the underlying *sql.DB, connecting lazily on first call.
// sqlc-generated query functions accept *sql.DB directly.
func (s *sqlDatabase) GetDB() *sql.DB {
	s.connect()
	return s.db
}

// Ping verifies the connection is alive, establishing one if necessary.
func (s *sqlDatabase) Ping(ctx context.Context) error {
	s.connect()
	if s.initErr != nil {
		return s.initErr
	}
	return s.db.PingContext(ctx)
}

// Close closes all idle connections in the pool.
func (s *sqlDatabase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
