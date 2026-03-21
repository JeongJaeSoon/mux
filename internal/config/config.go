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
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mux", "config.yaml")
}

func Exists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(ConfigPath())
	viper.SetDefault("repos_dir", filepath.Join(home, "conductor", "repos"))

	_ = viper.ReadInConfig() // ignore file-not-found

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.ReposDir = expandHome(cfg.ReposDir, home)

	return &cfg, nil
}

func Save(cfg *Config) error {
	path := ConfigPath()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	content := fmt.Sprintf("repos_dir: %s\n", cfg.ReposDir)
	return os.WriteFile(path, []byte(content), 0644)
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
