package mysql_test

import (
	"context"
	"fmt"

	"github.com/armineyvazi/common.git/pkg/adapters/database/mysql"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// ExampleNew shows constructing a GORM MySQL adapter and running AutoMigrate
// to create or update the schema from GORM model definitions.
func ExampleNew() {
	db := mysql.New(
		"localhost:3306",
		"mydb",
		"app",
		"secret",
		"charset=utf8mb4",
		mysql.Config{PrepareStmt: true},
	)
	defer db.Close()

	// Adapter satisfies both ports.Database and ports.GORMMigrator.
	var _ ports.Database = db
	var _ ports.GORMMigrator = db

	type User struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:255;not null"`
		Email string `gorm:"uniqueIndex;size:320"`
	}

	// AutoMigrate creates missing tables and adds missing columns.
	// Safe to run on every startup — it never drops columns or tables.
	if err := db.AutoMigrate(&User{}); err != nil {
		fmt.Println("auto-migrate:", err)
		return
	}

	ctx := context.Background()
	gormDB := db.GetConnection(ctx)
	_ = gormDB
}

// ExampleNew_multiModel shows migrating several models in one call.
func ExampleNew_multiModel() {
	db := mysql.New("localhost:3306", "shop", "app", "secret", "charset=utf8mb4", mysql.Config{})
	defer db.Close()

	type Category struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100;uniqueIndex"`
	}
	type Product struct {
		ID         uint   `gorm:"primaryKey"`
		CategoryID uint   `gorm:"index"`
		Title      string `gorm:"size:500"`
		PriceCents int64
	}

	if err := db.AutoMigrate(&Category{}, &Product{}); err != nil {
		fmt.Println("auto-migrate:", err)
	}
}
