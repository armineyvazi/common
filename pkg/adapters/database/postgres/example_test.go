package postgres_test

import (
	"context"
	"fmt"
	"time"

	"github.com/armineyvazi/common.git/pkg/adapters/database/postgres"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// ExampleNew demonstrates constructing a GORM PostgreSQL adapter and running
// AutoMigrate to create or update the schema from GORM model definitions.
func ExampleNew() {
	db := postgres.New("localhost", "mydb", "user", "secret", 5432, postgres.Config{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		PrepareStmt:     true,
	})
	defer db.Close()

	// Adapter satisfies both ports.Database and ports.GORMMigrator.
	var _ ports.Database = db
	var _ ports.GORMMigrator = db

	// Define GORM models inline or import them from your domain package.
	type User struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:255;not null"`
		Email string `gorm:"uniqueIndex"`
	}
	type Order struct {
		ID     uint `gorm:"primaryKey"`
		UserID uint `gorm:"index"`
		Total  int64
	}

	// AutoMigrate creates missing tables and adds missing columns.
	// Safe to call on every startup — it never drops columns or tables.
	if err := db.AutoMigrate(&User{}, &Order{}); err != nil {
		fmt.Println("auto-migrate:", err)
		return
	}

	ctx := context.Background()
	gormDB := db.GetConnection(ctx)
	_ = gormDB // pass to GORM-based repositories
}

// ExampleNewWithDSN shows construction from a pre-built DSN string, which
// is useful when the DSN is managed by a secrets manager.
func ExampleNewWithDSN() {
	dsn := "host=localhost user=app password=secret dbname=mydb port=5432 sslmode=require TimeZone=UTC"
	db := postgres.NewWithDSN(dsn, postgres.Config{PrepareStmt: true})
	defer db.Close()

	type Product struct {
		ID    uint   `gorm:"primaryKey"`
		Title string `gorm:"size:500"`
		Price int64
	}
	if err := db.AutoMigrate(&Product{}); err != nil {
		fmt.Println("auto-migrate:", err)
	}
}

// ExampleNewSQL shows the raw *sql.DB adapter, intended for use with
// sqlc-generated query functions. There is no ORM layer.
func ExampleNewSQL() {
	sqlDB := postgres.NewSQL(
		"postgres://app:secret@localhost:5432/mydb?sslmode=require",
		postgres.SQLConfig{
			MaxOpenConns:    20,
			MaxIdleConns:    5,
			ConnMaxLifetime: 10 * time.Minute,
			SSLMode:         "require",
		},
	)
	defer sqlDB.Close()

	ctx := context.Background()
	if err := sqlDB.Ping(ctx); err != nil {
		fmt.Println("ping:", err)
		return
	}

	// Pass GetDB() to the sqlc-generated constructor:
	//   queries := db.New(sqlDB.GetDB())
	db := sqlDB.GetDB()
	_ = db
}

// ExampleNewSQLFromParts builds a DSN from individual parts, URL-encoding
// any special characters in the password automatically.
func ExampleNewSQLFromParts() {
	sqlDB := postgres.NewSQLFromParts(
		"localhost", "mydb", "app", "p@$$w0rd", 5432,
		postgres.SQLConfig{SSLMode: "require"},
	)
	defer sqlDB.Close()

	db := sqlDB.GetDB() // hand to sqlc queries
	_ = db
}
