package mysql_test

import (
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/database/mysql"
	"github.com/armineyvazi/common.git/pkg/ports"
)

// TestMysql_ImplementsGORMMigrator is a compile-time assertion that
// *Mysql satisfies ports.GORMMigrator.
func TestMysql_ImplementsGORMMigrator(t *testing.T) {
	var _ ports.GORMMigrator = mysql.New("", "", "", "", "", mysql.Config{})
}

// TestMysql_ImplementsDatabase verifies the base Database interface.
func TestMysql_ImplementsDatabase(t *testing.T) {
	var _ ports.Database = mysql.New("", "", "", "", "", mysql.Config{})
}
