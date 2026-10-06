package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/armineyvazi/common.git/pkg/adapters/config"
)

// AppConfig is a typical application configuration struct. Field names are
// matched to YAML keys via mapstructure tags (case-insensitive by default).
type AppConfig struct {
	AppName  string `mapstructure:"app_name"`
	Port     int    `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
	Database struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
		Name string `mapstructure:"name"`
	} `mapstructure:"database"`
}

func (c AppConfig) GetConfig() AppConfig { return c }

// ExampleNewViper_basic shows loading a YAML config file into a typed struct.
func ExampleNewViper_basic() {
	dir, _ := os.MkdirTemp("", "cfg-example")
	defer os.RemoveAll(dir)

	cfgFile := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(cfgFile, []byte(`
app_name: my-service
port: 8080
log_level: info
database:
  host: localhost
  port: 5432
  name: mydb
`), 0600)

	var cfg AppConfig
	if err := config.NewViper(&cfg, cfgFile); err != nil {
		fmt.Println("load config:", err)
		return
	}

	fmt.Println("name:", cfg.AppName)
	fmt.Println("port:", cfg.Port)
	fmt.Println("db host:", cfg.Database.Host)
	// Output:
	// name: my-service
	// port: 8080
	// db host: localhost
}

// ExampleNewViper_envOverride shows that environment variables override file
// values. Key mapping replaces "." with "__" so nested keys work:
// DATABASE__HOST overrides database.host.
func ExampleNewViper_envOverride() {
	dir, _ := os.MkdirTemp("", "cfg-env")
	defer os.RemoveAll(dir)

	cfgFile := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(cfgFile, []byte("port: 3000\napp_name: default\n"), 0600)

	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	var cfg AppConfig
	if err := config.NewViper(&cfg, cfgFile); err != nil {
		fmt.Println("load config:", err)
		return
	}

	fmt.Println("port:", cfg.Port) // overridden by PORT env var
	// Output:
	// port: 9090
}
