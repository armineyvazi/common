package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	gomysql "github.com/go-sql-driver/mysql" // registers the "mysql" driver for database/sql

	"github.com/armineyvazi/common.git/pkg/ports"
)

// MySQLSQLConfig configures the *sql.DB connection pool for the raw SQL adapter.
type MySQLSQLConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	// TLSConfig controls TLS for the connection.
	// Accepted values: "true" (verify cert, default), "skip-verify" (encrypt only),
	// "false" (no TLS, local dev only), or a name registered via mysql.RegisterTLSConfig.
	// Defaults to "true" when empty — system CA pool verification.
	TLSConfig string
}

type sqlMysqlDB struct {
	dsn     string
	cfg     MySQLSQLConfig
	once    sync.Once
	db      *sql.DB
	initErr error
}

// NewSQL returns a ports.SQLDatabase backed by go-sql-driver/mysql via database/sql.
// The DSN must follow the go-sql-driver format, e.g.
//
//	"user:password@tcp(host:3306)/dbname?parseTime=true&loc=Local"
func NewSQL(dsn string, cfg MySQLSQLConfig) ports.SQLDatabase {
	return &sqlMysqlDB{dsn: dsn, cfg: cfg}
}

// NewSQLFromParts constructs a DSN using mysql.Config.FormatDSN so that
// special characters in user, password, or host are escaped safely.
// charset is the connection charset, e.g. "utf8mb4".
func NewSQLFromParts(host, database, user, password, charset string, cfg MySQLSQLConfig) ports.SQLDatabase {
	tlsCfg := cfg.TLSConfig
	if tlsCfg == "" {
		tlsCfg = "true" // verify server cert against system CA pool
	}
	driverCfg := gomysql.Config{
		User:      user,
		Passwd:    password,
		Net:       "tcp",
		Addr:      host,
		DBName:    database,
		Params:    map[string]string{"charset": charset},
		ParseTime: true,
		Loc:       time.Local,
		TLSConfig: tlsCfg,
	}
	return NewSQL(driverCfg.FormatDSN(), cfg)
}

func (s *sqlMysqlDB) connect() {
	s.once.Do(func() {
		db, err := sql.Open("mysql", s.dsn)
		if err != nil {
			s.initErr = fmt.Errorf("open mysql sql.DB: %w", err)
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
func (s *sqlMysqlDB) GetDB() *sql.DB {
	s.connect()
	return s.db
}

// Ping verifies the connection is alive, establishing one if necessary.
func (s *sqlMysqlDB) Ping(ctx context.Context) error {
	s.connect()
	if s.initErr != nil {
		return s.initErr
	}
	return s.db.PingContext(ctx)
}

// Close closes all idle connections in the pool.
func (s *sqlMysqlDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
