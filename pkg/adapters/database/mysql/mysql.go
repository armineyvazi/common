package mysql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/armineyvazi/common.git/pkg/ports"

	mysql "go.elastic.co/apm/module/apmgormv2/v2/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	address string

	dbConnOnce sync.Once
	db         *gorm.DB

	config Config
}

type Config struct {
	PrepareStmt bool
}

func New(host, database, user, password, charset string, config Config) ports.Database {
	return &Mysql{
		address: createDSN(host, database, user, password, charset),
		config: Config{
			PrepareStmt: config.PrepareStmt,
		},
	}
}

func createDSN(host, database, user, password, charset string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s&parseTime=True&loc=Local", user, password, host, database, charset)
}

func (m *Mysql) GetConnection(ctx context.Context) *gorm.DB {
	if m.db == nil {
		m.dbConnOnce.Do(func() {
			var err error
			m.db, err = gorm.Open(mysql.Open(m.address), &gorm.Config{
				PrepareStmt: m.config.PrepareStmt,
				NowFunc: func() time.Time {
					ti, _ := time.LoadLocation("Asia/Tehran")
					return time.Now().In(ti)
				},
			})
			if err != nil {
				panic(err)
			}

			db, err := m.db.DB()
			if err != nil {
				panic(err)
			}
			db.SetMaxIdleConns(10)
			db.SetMaxOpenConns(25)
			db.SetConnMaxLifetime(time.Hour)
		})
	}
	return m.db.WithContext(ctx)
}

func (m *Mysql) ServiceName() string {
	return fmt.Sprintf("mysql_%s", m.address)
}

func (m *Mysql) IsHealthy(ctx context.Context) bool {
	m.GetConnection(ctx)
	db, err := m.db.DB()
	return err == nil && db.Ping() == nil
}

func (m *Mysql) Close() error {
	if m.db == nil {
		return nil
	}
	db, err := m.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
