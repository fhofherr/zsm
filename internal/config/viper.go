package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// New creates a new instance of *viper.Viper configured for use within zsm.
// The returned Viper expects its default configuration file at
// /etc/zsm/config.yaml.
func New() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/zsm/")

	setDefaults(v)

	return v
}

// Read makes v read its config from the default configuration file if it
// exists.
func Read(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}
	return nil
}

// ReadFile makes v read the configuration from file. ReadFile returns an error
// if the file does not exist or reading it fails.
func ReadFile(v *viper.Viper, file string) error {
	r, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open config %s: %w", file, err)
	}
	defer r.Close()
	if err := v.ReadConfig(r); err != nil {
		return fmt.Errorf("read config %s: %w", file, err)
	}
	return nil
}
