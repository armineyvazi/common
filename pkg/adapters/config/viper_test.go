package config

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	AppName string `mapstructure:"app_name"`
	Port    int    `mapstructure:"port"`
}

func (c testConfig) GetConfig() testConfig { return c }

func TestNewViper_ReadsYAML(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "config*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer f.Close()

	content := "app_name: myapp\nport: 8080\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var cfg testConfig
	if err := NewViper(&cfg, f.Name()); err != nil {
		t.Fatalf("NewViper: %v", err)
	}

	if cfg.AppName != "myapp" {
		t.Errorf("app_name: got %q, want %q", cfg.AppName, "myapp")
	}
	if cfg.Port != 8080 {
		t.Errorf("port: got %d, want 8080", cfg.Port)
	}
}

func TestNewViper_EnvOverride(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "config*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer f.Close()

	_, _ = f.WriteString("port: 3000\n")

	t.Setenv("PORT", "9090")

	var cfg testConfig
	if err := NewViper(&cfg, f.Name()); err != nil {
		t.Fatalf("NewViper: %v", err)
	}
	// Environment variable should override the file value.
	if cfg.Port != 9090 {
		t.Errorf("port: got %d, want 9090 (env override)", cfg.Port)
	}
}

func TestNewViper_MissingFile(t *testing.T) {
	var cfg testConfig
	err := NewViper(&cfg, filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Error("expected error for missing config file")
	}
}
