// Package migrate wraps golang-migrate/migrate/v4 for use with
// the ports.SQLDatabase adapter. It supports both PostgreSQL and MySQL.
//
// Migration files follow the golang-migrate naming convention:
//
//	{version}_{title}.up.sql    — applied by Up / Steps(+n)
//	{version}_{title}.down.sql  — applied by Down / Steps(-n)
//
// Typical usage with go:embed:
//
//	//go:embed migrations
//	var migrationsFS embed.FS
//
//	sqlDB := postgres.NewSQL(dsn, cfg)
//
//	m, err := dbmigrate.NewPostgres(sqlDB.GetDB(), migrationsFS, "migrations")
//	if err != nil { log.Fatal(err) }
//	defer m.Close()
//
//	if err := m.Up(); err != nil { log.Fatal(err) }
package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"

	gmigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	gmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	gpostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrator applies versioned SQL migration files to a database.
// golang-migrate acquires an advisory lock on the database, so concurrent
// Migrators against the same database are serialised automatically.
type Migrator struct {
	m *gmigrate.Migrate
}

// NewPostgres creates a Migrator for a PostgreSQL database.
// db must be an open, connected *sql.DB (any Postgres driver works).
// migrationsFS is an fs.FS containing migration files at migrationsPath.
func NewPostgres(db *sql.DB, migrationsFS fs.FS, migrationsPath string) (*Migrator, error) {
	driver, err := gpostgres.WithInstance(db, &gpostgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("migrate postgres: init driver: %w", err)
	}
	return newMigrator(driver, migrationsFS, migrationsPath, "postgres")
}

// NewMySQL creates a Migrator for a MySQL database.
// db must be an open, connected *sql.DB.
// migrationsFS is an fs.FS containing migration files at migrationsPath.
func NewMySQL(db *sql.DB, migrationsFS fs.FS, migrationsPath string) (*Migrator, error) {
	driver, err := gmysql.WithInstance(db, &gmysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("migrate mysql: init driver: %w", err)
	}
	return newMigrator(driver, migrationsFS, migrationsPath, "mysql")
}

// Up applies all pending migrations in ascending version order.
// Returns nil when the database is already at the latest version.
func (m *Migrator) Up() error {
	err := m.m.Up()
	if errors.Is(err, gmigrate.ErrNoChange) {
		return nil
	}
	return err
}

// Down rolls back all applied migrations in descending version order.
// Returns nil when no migrations have been applied.
func (m *Migrator) Down() error {
	err := m.m.Down()
	if errors.Is(err, gmigrate.ErrNoChange) {
		return nil
	}
	return err
}

// Steps applies exactly n migration steps. Positive n migrates up,
// negative n migrates down. For example, Steps(-1) rolls back one version.
func (m *Migrator) Steps(n int) error {
	return m.m.Steps(n)
}

// Version returns the current schema version and whether the database is in
// a dirty state (a previous migration failed mid-way). Returns (0, false, nil)
// when no migrations have been applied yet.
func (m *Migrator) Version() (version uint, dirty bool, err error) {
	v, d, e := m.m.Version()
	if errors.Is(e, gmigrate.ErrNilVersion) {
		return 0, false, nil
	}
	return v, d, e
}

// Force sets the schema version without running any migration SQL.
// Use this only to recover from a dirty state after manually fixing a
// failed migration. Positive version pins to that version; -1 clears it.
func (m *Migrator) Force(version int) error {
	return m.m.Force(version)
}

// Close releases the source reader and the database handle held by the
// Migrator. The underlying *sql.DB is NOT closed — that is the caller's
// responsibility.
func (m *Migrator) Close() error {
	srcErr, dbErr := m.m.Close()
	if srcErr != nil {
		return fmt.Errorf("migrate: close source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("migrate: close database: %w", dbErr)
	}
	return nil
}

// newMigrator is the shared constructor used by NewPostgres and NewMySQL.
func newMigrator(driver database.Driver, migrationsFS fs.FS, path, dbName string) (*Migrator, error) {
	src, err := iofs.New(migrationsFS, path)
	if err != nil {
		return nil, fmt.Errorf("migrate: open source %q: %w", path, err)
	}
	m, err := gmigrate.NewWithInstance("iofs", src, dbName, driver)
	if err != nil {
		return nil, fmt.Errorf("migrate: init: %w", err)
	}
	return &Migrator{m: m}, nil
}
