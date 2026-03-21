package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ReposDir string `mapstructure:"repos_dir"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(filepath.Join(home, ".config", "mux", "config.yaml"))
	viper.SetDefault("repos_dir", filepath.Join(home, "conductor", "repos"))

	_ = viper.ReadInConfig() // ignore file-not-found

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.ReposDir = expandHome(cfg.ReposDir, home)

	return &cfg, nil
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
