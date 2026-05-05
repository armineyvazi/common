package clickhouse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/armineyvazi/common.git/pkg/ports"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type ClickHouse struct {
	address    string
	dbConnOnce sync.Once
	db         *gorm.DB
	config     Config
}

type Config struct {
	PrepareStmt     bool
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func New(host, database, user, password string, port int, config Config) ports.Database {
	return &ClickHouse{
		address: createDSN(host, database, user, password, port),
		config:  config,
	}
}

func createDSN(host, database, user, password string, port int) string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s?dial_timeout=120s&read_timeout=120s",
		user, password, host, port, database)
}

func (c *ClickHouse) GetConnection(ctx context.Context) *gorm.DB {
	if c.db == nil {
		c.dbConnOnce.Do(func() {
			var err error
			c.db, err = gorm.Open(clickhouse.Open(c.address), &gorm.Config{
				PrepareStmt: c.config.PrepareStmt,
				NowFunc: func() time.Time {
					ti, _ := time.LoadLocation("Asia/Tehran")
					return time.Now().In(ti)
				},
			})
			if err != nil {
				panic(err)
			}

			db, err := c.db.DB()
			if err != nil {
				panic(err)
			}
			if c.config.MaxIdleConns > 0 {
				db.SetMaxIdleConns(c.config.MaxIdleConns)
			}
			if c.config.MaxOpenConns > 0 {
				db.SetMaxOpenConns(c.config.MaxOpenConns)
			}
			if c.config.ConnMaxLifetime > 0 {
				db.SetConnMaxLifetime(c.config.ConnMaxLifetime)
			}
		})
	}
	return c.db.WithContext(ctx)
}

func (c *ClickHouse) ServiceName() string {
	return fmt.Sprintf("clickhouse_%s", c.address)
}

func (c *ClickHouse) IsHealthy(ctx context.Context) bool {
	c.GetConnection(ctx)
	db, err := c.db.DB()
	return err == nil && db.Ping() == nil
}

func (c *ClickHouse) Close() error {
	if c.db == nil {
		return nil
	}
	db, err := c.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
