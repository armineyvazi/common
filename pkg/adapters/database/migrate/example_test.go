package migrate_test

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	dbmigrate "github.com/armineyvazi/common.git/pkg/adapters/database/migrate"
)

// Migration files must follow the golang-migrate naming convention:
//
//	{version}_{title}.up.sql    — applied on Up / Steps(+n)
//	{version}_{title}.down.sql  — applied on Down / Steps(-n)
//
// Embed them with go:embed in your application code:
//
//	//go:embed migrations
//	var migrationsFS embed.FS

//go:embed testdata/postgres/migrations
var exPGMigrationsFS embed.FS

//go:embed testdata/mysql/migrations
var exMySQLMigrationsFS embed.FS

// ExampleNewPostgres shows the recommended production pattern for sqlc
// projects: embed versioned SQL files and apply them before starting the app.
func ExampleNewPostgres() {
	db, err := sql.Open("pgx", "postgres://app:secret@localhost:5432/mydb?sslmode=require")
	if err != nil {
		fmt.Println("open:", err)
		return
	}
	defer func() { _ = db.Close() }()

	m, err := dbmigrate.NewPostgres(db, exPGMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		fmt.Println("migrator:", err)
		return
	}
	defer func() { _ = m.Close() }()

	// Up applies all pending migrations. Returns nil when already at latest.
	if err := m.Up(); err != nil {
		fmt.Println("up:", err)
		return
	}

	version, dirty, _ := m.Version()
	fmt.Printf("schema version: %d, dirty: %v\n", version, dirty)
}

// ExampleNewMySQL shows the same pattern for MySQL databases.
func ExampleNewMySQL() {
	db, err := sql.Open("mysql", "root:secret@tcp(localhost:3306)/mydb?parseTime=True")
	if err != nil {
		fmt.Println("open:", err)
		return
	}
	defer func() { _ = db.Close() }()

	m, err := dbmigrate.NewMySQL(db, exMySQLMigrationsFS, "testdata/mysql/migrations")
	if err != nil {
		fmt.Println("migrator:", err)
		return
	}
	defer func() { _ = m.Close() }()

	if err := m.Up(); err != nil {
		fmt.Println("up:", err)
	}
}

// ExampleMigrator_Steps shows rolling back one migration step during
// a controlled deploy rollback.
func ExampleMigrator_Steps() {
	db, _ := sql.Open("pgx", "postgres://app:secret@localhost:5432/mydb?sslmode=require")
	defer func() { _ = db.Close() }()

	m, err := dbmigrate.NewPostgres(db, exPGMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		fmt.Println("migrator:", err)
		return
	}
	defer func() { _ = m.Close() }()

	// Roll back one version.
	if err := m.Steps(-1); err != nil {
		fmt.Println("rollback:", err)
		return
	}

	// Apply one version forward.
	if err := m.Steps(1); err != nil {
		fmt.Println("migrate:", err)
	}
}

// ExampleMigrator_Force shows how to recover from a dirty state after
// manually fixing a failed migration's side effects in the database.
func ExampleMigrator_Force() {
	db, _ := sql.Open("pgx", "postgres://app:secret@localhost:5432/mydb?sslmode=require")
	defer func() { _ = db.Close() }()

	m, err := dbmigrate.NewPostgres(db, exPGMigrationsFS, "testdata/postgres/migrations")
	if err != nil {
		return
	}
	defer func() { _ = m.Close() }()

	_, dirty, _ := m.Version()
	if dirty {
		// Reset the dirty flag without re-running the failed migration.
		// Only call this after you have manually fixed the database state.
		if err := m.Force(1); err != nil {
			fmt.Println("force:", err)
		}
	}
}
