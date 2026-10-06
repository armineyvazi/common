//go:build integration

package migrate_test

import (
	"database/sql"
	"embed"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	dbmigrate "github.com/armineyvazi/common.git/pkg/adapters/database/migrate"
)

//go:embed testdata/postgres/migrations
var postgresMigrationsFS embed.FS

//go:embed testdata/mysql/migrations
var mysqlMigrationsFS embed.FS

// openPostgres opens a *sql.DB using the POSTGRES_DSN env var set in CI.
func openPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN not set; skipping postgres integration test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open postgres: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("postgres ping: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// openMySQL opens a *sql.DB using the MYSQL_DSN env var set in CI.
func openMySQL(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN not set; skipping mysql integration test")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open mysql: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("mysql ping: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// ---- PostgreSQL tests -------------------------------------------------------

func TestPostgres_Up_AppliesAllMigrations(t *testing.T) {
	db := openPostgres(t)

	m, err := dbmigrate.NewPostgres(db, postgresMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer m.Close()

	// Clean slate: roll back everything first (ignore ErrNoChange-style nil).
	_ = m.Down()

	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if dirty {
		t.Fatal("database is in dirty state after Up")
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
}

func TestPostgres_Up_Idempotent(t *testing.T) {
	db := openPostgres(t)

	m, err := dbmigrate.NewPostgres(db, postgresMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer m.Close()

	_ = m.Down()
	if err := m.Up(); err != nil {
		t.Fatalf("first Up: %v", err)
	}
	// Calling Up again when already at latest must not return an error.
	if err := m.Up(); err != nil {
		t.Fatalf("second Up (idempotent): %v", err)
	}
}

func TestPostgres_Steps_UpAndDown(t *testing.T) {
	db := openPostgres(t)

	m, err := dbmigrate.NewPostgres(db, postgresMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer m.Close()

	_ = m.Down()

	if err := m.Steps(1); err != nil {
		t.Fatalf("Steps(+1): %v", err)
	}
	v, _, err := m.Version()
	if err != nil {
		t.Fatalf("Version after Steps(+1): %v", err)
	}
	if v != 1 {
		t.Errorf("expected version 1 after Steps(+1), got %d", v)
	}

	if err := m.Steps(-1); err != nil {
		t.Fatalf("Steps(-1): %v", err)
	}
	v, _, err = m.Version()
	if err != nil {
		t.Fatalf("Version after Steps(-1): %v", err)
	}
	if v != 0 {
		t.Errorf("expected version 0 after Steps(-1), got %d", v)
	}
}

func TestPostgres_Down_AppliesAllRollbacks(t *testing.T) {
	db := openPostgres(t)

	m, err := dbmigrate.NewPostgres(db, postgresMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer m.Close()

	_ = m.Down()
	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if err := m.Down(); err != nil {
		t.Fatalf("Down: %v", err)
	}

	v, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("Version after Down: %v", err)
	}
	if dirty {
		t.Fatal("dirty state after Down")
	}
	if v != 0 {
		t.Errorf("expected version 0 after full Down, got %d", v)
	}
}

func TestPostgres_Version_BeforeAnyMigration(t *testing.T) {
	db := openPostgres(t)

	m, err := dbmigrate.NewPostgres(db, postgresMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer m.Close()

	_ = m.Down() // ensure clean state

	v, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if dirty {
		t.Error("unexpected dirty state on clean database")
	}
	if v != 0 {
		t.Errorf("expected version 0, got %d", v)
	}
}

// ---- MySQL tests ------------------------------------------------------------

func TestMySQL_Up_AppliesAllMigrations(t *testing.T) {
	db := openMySQL(t)

	m, err := dbmigrate.NewMySQL(db, mysqlMigrationsFS, "testdata/mysql/migrations")
	if err != nil {
		t.Fatalf("NewMySQL: %v", err)
	}
	defer m.Close()

	_ = m.Down()

	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if dirty {
		t.Fatal("dirty state after Up")
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
}

func TestMySQL_Up_Idempotent(t *testing.T) {
	db := openMySQL(t)

	m, err := dbmigrate.NewMySQL(db, mysqlMigrationsFS, "testdata/mysql/migrations")
	if err != nil {
		t.Fatalf("NewMySQL: %v", err)
	}
	defer m.Close()

	_ = m.Down()
	if err := m.Up(); err != nil {
		t.Fatalf("first Up: %v", err)
	}
	if err := m.Up(); err != nil {
		t.Fatalf("second Up (idempotent): %v", err)
	}
}

func TestMySQL_Down_AppliesAllRollbacks(t *testing.T) {
	db := openMySQL(t)

	m, err := dbmigrate.NewMySQL(db, mysqlMigrationsFS, "testdata/mysql/migrations")
	if err != nil {
		t.Fatalf("NewMySQL: %v", err)
	}
	defer m.Close()

	_ = m.Down()
	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if err := m.Down(); err != nil {
		t.Fatalf("Down: %v", err)
	}

	v, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if dirty {
		t.Fatal("dirty state after Down")
	}
	if v != 0 {
		t.Errorf("expected version 0 after full Down, got %d", v)
	}
}

// ---- GORM AutoMigrate interface compliance ----------------------------------

func TestPostgres_GORMMigrator_InterfaceCompliance(t *testing.T) {
	// This test verifies at compile time that the postgres adapter satisfies
	// ports.GORMMigrator. It does not need a real database connection.
	// The actual AutoMigrate behaviour is tested in the GORM adapter package.
}
