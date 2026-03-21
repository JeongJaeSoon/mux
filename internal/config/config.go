package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ReposDir string `mapstructure:"repos_dir"`
}

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "mux", "config.yaml")
}

func Exists() bool {
	path := ConfigPath()
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	path := ConfigPath()
	if path != "" {
		viper.SetConfigFile(path)
	}
	viper.SetDefault("repos_dir", filepath.Join(home, "conductor", "repos"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("config read error: %w", err)
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.ReposDir = expandHome(cfg.ReposDir, home)

	return &cfg, nil
}

func Save(cfg *Config) error {
	path := ConfigPath()
	if path == "" {
		return fmt.Errorf("cannot determine config path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	content := fmt.Sprintf("repos_dir: %s\n", cfg.ReposDir)
	return os.WriteFile(path, []byte(content), 0600)
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
