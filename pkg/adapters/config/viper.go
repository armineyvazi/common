package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config[T any] interface {
	GetConfig() T
}

func NewViper[T Config[T]](config *T, configAddress string) error {
	viper.SetConfigFile(configAddress)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshalling configuration: %w", err)
	}
	return nil
}
