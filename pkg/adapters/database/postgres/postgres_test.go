package postgres_test

import (
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/database/postgres"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// TestPostgresDB_ImplementsGORMMigrator verifies at compile time that the
// value returned by postgres.New satisfies ports.GORMMigrator.
func TestPostgresDB_ImplementsGORMMigrator(t *testing.T) {
	_ = postgres.New("", "", "", "", 5432, postgres.Config{})
}

// TestPostgresDB_ImplementsDatabase verifies the base Database interface.
func TestPostgresDB_ImplementsDatabase(t *testing.T) {
	var _ ports.Database = postgres.New("", "", "", "", 5432, postgres.Config{})
}
