package postgres_test

import (
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/database/postgres"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// TestPostgresDB_ImplementsGORMMigrator is a compile-time assertion that
// *postgresDB satisfies ports.GORMMigrator. If AutoMigrate is removed or its
// signature changes, this test will fail to compile.
func TestPostgresDB_ImplementsGORMMigrator(t *testing.T) {
	var _ ports.GORMMigrator = postgres.New("", "", "", "", 5432, postgres.Config{})
}

// TestPostgresDB_ImplementsDatabase verifies the base Database interface.
func TestPostgresDB_ImplementsDatabase(t *testing.T) {
	var _ ports.Database = postgres.New("", "", "", "", 5432, postgres.Config{})
}
